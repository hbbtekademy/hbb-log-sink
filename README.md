# Log Sink

## Reads from `stdin` and writes to specified log file.

### Usage

`./some-process | ./logsink -logfile ./some-process.log -max-size-mb 10 -compress`

## Features

- Logs rollover after reaching specified max-size MB.
- Rolled over logs can be automatically gzipped.
- Using logsink is almost twice faster than redirecting to file
  `./some-process > some-process.log`

### Run Benchmarks

Execute following commands on Linux to run benchmarks:

```bash
go build -o loggen.exe loggen/main.go
go build -o logsink.exe main.go
time ./loggen.exe > test1.log
time ./loggen.exe | ./logsink.exe -logfile test2.log -max-size-mb 500 -compress
```
