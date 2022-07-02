# PVPC Telegraf Plugin

[![golangci-lint](https://github.com/MrMarble/pvpc-telegraf-plugin/actions/workflows/lint.yml/badge.svg)](https://github.com/MrMarble/pvpc-telegraf-plugin/actions/workflows/lint.yml)

Gather Spanish electricity hourly prices from https://www.ree.es/es/apidatos.

## Build and Run

To build this plugin, just run (go environment needed, at least 1.17):

```
make
```

Which will build the binary ./bin/pvpc

You can run it with `./bin/pvpc --config plugin.conf`

## Configuration

```toml
[[inputs.Pvpc]]
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
	http_timeout="10s"

```

Once compiled and configured, you could add the plugin to Telegraf adding this configuration:

```toml
[[inputs.execd]]
  ## One program to run as daemon.
  ## NOTE: process and each argument should each be their own string
  command = ["/path/to/pvpc", "--config", "/path/to/plugin.conf"]

  ## Define how the process is signaled on each collection interval.
  ## Valid values are:
  ##   "none"    : Do not signal anything. (Recommended for service inputs)
  ##               The process must output metrics by itself.
  ##   "STDIN"   : Send a newline on STDIN. (Recommended for gather inputs)
  ##   "SIGHUP"  : Send a HUP signal. Not available on Windows. (not recommended)
  ##   "SIGUSR1" : Send a USR1 signal. Not available on Windows.
  ##   "SIGUSR2" : Send a USR2 signal. Not available on Windows.
  signal = "none"

  ## Delay before the process is restarted after an unexpected termination
  restart_delay = "10s"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
  
  # Set interval to 24h
  interval = "24h"
```

## Metrics

- pvpc
    - tags:
        - geo_id (uin32)
    - fields:
        - value (float64)

## Example Output

```
pvpc,geo_id=8741 value=291.07 1575008500000000000
pvpc,geo_id=8741 value=196.58 1575008500000000000
```
