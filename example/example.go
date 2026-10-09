// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

// An HTTP server on anvil with zap logging: zaplog.DefaultConfig adjusted for Google Cloud Logging's
// field names and severities when it writes JSON. Its configuration is fixed in code, standing in
// for a real config provider such as anvil-koanf's.
//
//	go run ./example
package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"strconv"

	"go.uber.org/zap"

	"github.com/llingr/anvil"
	"github.com/llingr/anvil-zap/zaplog"
)

type Shell = anvil.Shell[Config, *zap.Logger]

func main() {
	loggerConfig := zaplog.DefaultConfig()
	if loggerConfig.Encoding == "json" {
		// Google Cloud Logging's field names and severities
		loggerConfig.EncoderConfig.LevelKey = "severity"
		loggerConfig.EncoderConfig.MessageKey = "message"
		loggerConfig.EncoderConfig.EncodeLevel = encodeSeverity
	}

	loggerProvider := zaplog.New(loggerConfig)
	exitCode := anvil.Run(context.Background(), "anvil-zap-example", fixedConfig{}, loggerProvider, wire)
	os.Exit(exitCode)
}

// wire serves HTTP on the configured port, stopped by the server's own Shutdown
func wire(ctx context.Context, shell Shell) error {
	logger := shell.Logger()
	config := shell.Config()
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(config.Port)))
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			_, _ = writer.Write([]byte("hello from anvil\n"))
		}),
		ReadHeaderTimeout: config.ReadHeaderTimeout,
	}
	shell.AddShutdownGroup(server).Go(func(context.Context) error {
		return server.Serve(listener)
	})
	logger.Info("serving", zap.String("address", listener.Addr().String()))
	return nil
}
