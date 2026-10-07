// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"time"

	"go.uber.org/zap/zapcore"
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

// fixedConfig is a config provider whose settings are written here rather than read
type fixedConfig struct{}

func (fixedConfig) Load(context.Context) (Config, error) {
	return Config{
		Port:              8080,
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}
