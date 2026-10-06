# anvil-zap

An [anvil](https://github.com/llingr/anvil) logger provider for [zap](https://github.com/uber-go/zap):
anvil's lifecycle lines go through zap, and `shell.Logger()` is the `*zap.Logger` the service uses.

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
- **Safe config logging.** `zaplog.LogConfig[Config](logger)` logs the loaded configuration through
  its `MarshalLogObject`, so only the fields it names appear, never its secrets. It plugs into
  [anvil-koanf](https://github.com/llingr/anvil-koanf)'s `conf.OnLoaded`.
- **Example.** `go run ./example` serves HTTP on 8080, logging with Google Cloud Logging's field
  names and severities.
