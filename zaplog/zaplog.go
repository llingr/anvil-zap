// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

// Package zaplog is an anvil logger provider backed by zap: New builds the logger from a zap.Config,
// and Wrap uses one the service built itself.
package zaplog

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/llingr/anvil"
)

// loggerProvider for anvil.LoggerProvider backed by *zap.Logger
type loggerProvider struct {
	logger *zap.Logger
}

// New builds a *zap.Logger wrapped in anvil.LoggerProvider
func New(cfg zap.Config) anvil.LoggerProvider[*zap.Logger] {
	logger, err := cfg.Build()
	if err != nil {
		panic(fmt.Errorf("zaplog: build logger - %w", err))
	}
	return Wrap(logger)
}

// Wrap *zap.Logger in anvil.LoggerProvider
func Wrap(logger *zap.Logger) anvil.LoggerProvider[*zap.Logger] {
	if logger == nil {
		panic("zaplog: invalid (nil) logger")
	}
	return &loggerProvider{
		logger: logger,
	}
}

// Logger is the *zap.Logger New built or Wrap was given, unchanged
func (l *loggerProvider) Logger() *zap.Logger {
	return l.logger
}

// LifecycleInfo logs anvil's lifecycle msg
func (l *loggerProvider) LifecycleInfo(_ context.Context, msg string) {
	// log caller's line, not this one
	l.logger.WithOptions(zap.AddCallerSkip(1)).Info(msg)
}

// LifecycleError logs anvil's lifecycle error along with an "error" field
func (l *loggerProvider) LifecycleError(_ context.Context, msg string, err error) {
	// log caller's line, not this one
	l.logger.WithOptions(zap.AddCallerSkip(1)).Error(msg, zap.Error(err))
}

// Flush syncs zap's buffered lines
func (l *loggerProvider) Flush() {
	_ = l.logger.Sync() // fails with "inappropriate ioctl for device" when stderr is a terminal
}

// DefaultConfig is zap's production config to stderr, info and above, with a UTC millisecond
// "time", durations as text ("5s"), no caller and no sampling. It writes JSON, or coloured lines
// when stderr is a terminal.
func DefaultConfig() zap.Config {
	cfg := zap.NewProductionConfig()
	cfg.Sampling = nil // every line, however many repeat
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = encodeTimeUTC
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	cfg.EncoderConfig.CallerKey = zapcore.OmitKey
	if stderrIsTerminal() {
		cfg.Encoding = "console"
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return cfg
}

// encodeTimeUTC writes 2026-10-05T10:36:39.070Z
func encodeTimeUTC(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(t.UTC().Format("2006-01-02T15:04:05.000Z"))
}

// stderrIsTerminal is true in an interactive shell, false under a pipe or a test
func stderrIsTerminal() bool {
	info, err := os.Stderr.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// LogConfig returns a config provider's loaded callback, such as anvil-koanf's conf.OnLoaded takes,
// logging the config through logger and the config's own MarshalLogObject
func LogConfig[C zapcore.ObjectMarshaler](logger *zap.Logger) func(ctx context.Context, config C) {
	return func(_ context.Context, config C) {
		logger.Info("configuration", zap.Object("config", config))
	}
}
