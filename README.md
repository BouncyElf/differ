# Differ
Differ proxy local request and forward request to 2 remote server and compare the response diff.

# Usage
***Make sure you are testing an api without any side-effect***

1. start the differ server
```bash
go install github.com/BouncyElf/differ@latest
differ <config_file>
```

2. curl the differ server and watch differ's log

# Configuration
The Config file is using yaml.
You can check the demo config below.
```yaml
# local proxy server setting
proxy_config:
  port: 8888
  enable_proxy_log: false

# the main remote server you want to compare as base standard
origin_scheme_and_host: 'https://example.com'

# the second remote server you want to check the diff
remote_scheme_and_host: 'https://example.com'

# the response headers you want to ignore when comparing the diff
# NOTE: highly recommand exclude `Date` header cause it will cause diff in most cases
exclude_headers:
  - Date

# async or sync call, optional
async_call: false
```
