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
