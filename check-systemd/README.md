# check-systemd

## Description
Checks systemd if any Unit(s) failed.

## Synopsis
```
check-systemd --pattern=PROCESS_NAME --state=STATE --warning-under=N
```

## Installation
```
go get github.com/b3n4kh/go-check-plugins
cd $(go env GOPATH)/src/github.com/b3n4kh/go-check-plugins/check-systemd
go install
```

## Usage
### Options

```
  -w, --warning-over=N                Trigger a warning if over a number
  -c, --critical-over=N               Trigger a critical if over a number
  -p, --pattern=PATTERN               Match a command against this pattern
  -x, --exclude-pattern=PATTERN       Don't match against a pattern to prevent false positives
```

## For more information
Please refer to the following.

- Execute `check-systemd -h` and you can get command line options.

## Other

- This is a Go port of [Josef-Friedrich/check_systemd](https://github.com/Josef-Friedrich/check_systemd).
- [Nagios Plugins - check_systemd.py](https://exchange.nagios.org/directory/Plugins/System-Metrics/Processes/check_systemd/details)
