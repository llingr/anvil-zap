// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/llingr/anvil"
)

// Config is the example's every setting
type Config struct {
	Port              int
	ReadHeaderTimeout time.Duration
}

func (c Config) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddInt("port", c.Port)
	encoder.AddDuration("readHeaderTimeout", c.ReadHeaderTimeout)
	return nil
}

type fixedConfig struct {
	loaded func(ctx context.Context, config Config)
}

// newFixedConfig is a config provider whose settings are written here rather than read, calling
// loaded with them as anvil-koanf's conf.OnLoaded would
func newFixedConfig(loaded func(ctx context.Context, config Config)) anvil.ConfigProvider[Config] {
	return &fixedConfig{
		loaded: loaded,
	}
}

func (f *fixedConfig) Load(ctx context.Context) (Config, error) {
	config := Config{
		Port:              8080,
		ReadHeaderTimeout: 5 * time.Second,
	}
	f.loaded(ctx, config)
	return config, nil
}
