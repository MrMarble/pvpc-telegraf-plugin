package pvpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Pvpc struct {
	TimeTrunc   string          `toml:"time_trunc"`
	GeoID       uint32          `toml:"geo_id"`
	StartDate   MarshableTime   `toml:"start_date"`
	EndDate     MarshableTime   `toml:"end_date"`
	HTTPTimeout config.Duration `toml:"http_timeout"`
	httpClient  *http.Client

	Log telegraf.Logger `toml:"-"`
}

type MarshableTime struct {
	time.Time
}

func (m *MarshableTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	s = s[1 : len(s)-1]
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	m.Time = t
	return nil
}

type Entry struct {
	Value      float64       `json:"value"`
	Percentage float64       `json:"percentage"`
	Date       MarshableTime `json:"datetime"`
}

type Attributes struct {
	Title  string  `json:"title"`
	Values []Entry `json:"values"`
}

type Entity struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Attributes `json:"attributes"`
}

type ReeData struct {
	Entities []Entity `json:"included"`
}

func (p *Pvpc) Description() string {
	return "Gather Spanish electricity hourly prices."
}

func (p *Pvpc) SampleConfig() string {
	return `
	## Defines the time aggregation of the requested data.
	time_trunc = "hour"

	## Time range. 
	## If omitted, today's price is obtained.
	## Defines the starting date in ISO 8601 format.
	#start_date="2021-12-26T00:00:00Z"
	## Defines the ending date in ISO 8601 format.
	#end_date="2021-12-26T23:59:00Z"

	## Id of the autonomous community/electrical system. Optional
	geo_id = 8741
	
	## Http request timeout.
	http_timeout="10s"`
}

// Init is for setup, and validating config.
func (p *Pvpc) Init() error {
	// TODO: check valid config from https://www.ree.es/es/apidatos
	return nil
}

func (p *Pvpc) createHTTPClient() *http.Client {
	client := http.Client{Timeout: time.Duration(p.HTTPTimeout)}
	return &client
}

func (p *Pvpc) craftURL() string {
	url := url.URL{
		Scheme: "https",
		Host:   "apidatos.ree.es",
		Path:   "/es/datos/mercados/precios-mercados-tiempo-real",
	}
	query := url.Query()
	query.Set("time_trunc", p.TimeTrunc)

	startDate := p.StartDate
	endDate := p.EndDate

	if p.StartDate.Year() == 1 || p.EndDate.Year() == 1 {
		now := time.Now()
		startDate = MarshableTime{Time: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())}
		endDate = MarshableTime{Time: time.Date(now.Year(), now.Month(), now.Day()+1, 23, 00, 0, 0, now.Location())}
	}

	query.Set("start_date", startDate.Format("2006-01-02T15:04"))
	query.Set("end_date", endDate.Format("2006-01-02T15:04"))

	if p.GeoID != 0 {
		query.Set("geo_id", fmt.Sprint(p.GeoID))
	}
	url.RawQuery = query.Encode()
	return url.String()
}

func (p *Pvpc) fetch() (*ReeData, error) {
	resp, err := p.httpClient.Get(p.craftURL())

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var data ReeData

	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (p *Pvpc) Gather(acc telegraf.Accumulator) error {
	p.Log.Info("Gathering PVPC price")

	if p.httpClient == nil {
		p.httpClient = p.createHTTPClient()
	}

	data, err := p.fetch()
	if err != nil {
		return err
	}

	if len(data.Entities) == 0 || len(data.Entities[0].Attributes.Values) == 0 {
		p.Log.Info("No values returned")
	}

	for _, price := range data.Entities[0].Attributes.Values {
		acc.AddFields("pvpc", map[string]interface{}{"price": price.Value}, map[string]string{"geo_id": fmt.Sprint(p.GeoID)}, price.Date.Local())
	}
	return nil
}

func init() {
	inputs.Add("Pvpc", func() telegraf.Input { return &Pvpc{} })
}
