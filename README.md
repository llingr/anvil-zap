# anvil-zap

An [anvil](https://github.com/llingr/anvil) logger provider backed by [zap](https://github.com/uber-go/zap):
anvil's lifecycle logs go through zap, and `shell.Logger()` is the `*zap.Logger` a service uses.

## Getting Started

```sh
go get github.com/llingr/anvil-zap
```

```go
loggerProvider := zaplog.New(zaplog.DefaultConfig())
exitCode := anvil.Run(context.Background(), "orders", configProvider, loggerProvider, wire)
os.Exit(exitCode)
```

## Features

- **Production defaults.** `zaplog.DefaultConfig()` is zap's production config with UTC millisecond
  timestamps, durations as text and no sampling: JSON in production, coloured lines in a terminal.
  It is a plain `zap.Config`, so any field can change before `zaplog.New` builds it.
- **An existing logger.** `zaplog.Wrap(logger)` uses a logger the service built itself.
- **Example.** `go run ./example` serves HTTP on 8080, logging with Google Cloud Logging's field
  names and severities.
