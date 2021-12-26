# PVPC Telegraf Plugin

Gather Spanish electricity hourly prices from https://www.ree.es/es/apidatos.

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
