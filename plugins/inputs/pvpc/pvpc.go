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

type Pvpc struct {
	TimeTrunc   string          `toml:"time_trunc"`
	GeoID       uint32          `toml:"geo_id"`
	StartDate   MarshableTime   `toml:"start_date"`
	EndDate     MarshableTime   `toml:"end_date"`
	HTTPTimeout config.Duration `toml:"http_timeout"`
	httpClient  *http.Client
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

func (s *Pvpc) Description() string {
	return "Gather Spanish electricity hourly prices."
}

//https://apidatos.ree.es/es/datos/mercados/precios-mercados-tiempo-real?start_date=2021/12/26T00:00&end_date=2021/12/26T23:59&time_trunc=hour&geo_id=5

func (s *Pvpc) SampleConfig() string {
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

func (p *Pvpc) createHttpClient() *http.Client {
	client := http.Client{Timeout: time.Duration(p.HTTPTimeout)}
	return &client
}

func (p *Pvpc) craftUrl() string {
	url := url.URL{
		Scheme: "https",
		Host:   "apidatos.ree.es",
		Path:   "/es/datos/mercados/precios-mercados-tiempo-real",
	}
	query := url.Query()
	query.Set("time_trunc", p.TimeTrunc)

	if p.StartDate.Year() == 1 || p.EndDate.Year() == 1 {
		now := time.Now()
		p.StartDate = MarshableTime{Time: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())}
		p.EndDate = MarshableTime{Time: time.Date(now.Year(), now.Month(), now.Day(), 23, 00, 0, 0, now.Location())}
	}

	query.Set("start_date", p.StartDate.Format("2006-01-02T15:04"))
	query.Set("end_date", p.EndDate.Format("2006-01-02T15:04"))

	if p.GeoID != 0 {
		query.Set("geo_id", fmt.Sprint(p.GeoID))
	}
	url.RawQuery = query.Encode()
	return url.String()
}

func (p *Pvpc) fetch() (*ReeData, error) {
	resp, err := p.httpClient.Get(p.craftUrl())

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
	if p.httpClient == nil {
		p.httpClient = p.createHttpClient()
	}

	data, err := p.fetch()
	if err != nil {
		return err
	}

	for _, price := range data.Entities[0].Attributes.Values {
		acc.AddFields("", map[string]interface{}{"value": price.Value}, nil, price.Date.Local())
	}
	return nil
}

func init() {
	inputs.Add("Pvpc", func() telegraf.Input { return &Pvpc{} })
}
