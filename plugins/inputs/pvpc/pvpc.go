package pvpc

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Pvpc struct {
	Ok  bool            `toml:"ok"`
	Log telegraf.Logger `toml:"-"`
}

func (s *Pvpc) Description() string {
	return "a demo plugin"
}

func (s *Pvpc) SampleConfig() string {
	return `
  ## Indicate if everything is fine
  ok = true
`
}

// Init is for setup, and validating config.
func (s *Pvpc) Init() error {
	return nil
}

func (s *Pvpc) Gather(acc telegraf.Accumulator) error {
	if s.Ok {
		acc.AddFields("state", map[string]interface{}{"value": "pretty good"}, nil)
	} else {
		acc.AddFields("state", map[string]interface{}{"value": "not great"}, nil)
	}

	return nil
}

func init() {
	inputs.Add("Pvpc", func() telegraf.Input { return &Pvpc{} })
}
