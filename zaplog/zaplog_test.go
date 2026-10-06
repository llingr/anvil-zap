// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

package zaplog_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/llingr/anvil-zap/zaplog"
)

// Wrap hands the application its own logger back unchanged
func TestWrapKeepsTheLogger(t *testing.T) {
	logger := zap.NewNop()
	if zaplog.Wrap(logger).Logger() != logger {
		t.Fatal("Logger() is not the wrapped logger")
	}
}

// anvil's lines through a caller-enabled logger name anvil's call site, here this file, not zaplog.go
func TestWrapCallerIsAnvilsCallSite(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	provider := zaplog.Wrap(zap.New(core, zap.AddCaller()))
	provider.LifecycleInfo(context.Background(), "starting")
	provider.LifecycleError(context.Background(), "failed", errors.New("boom"))
	for _, entry := range logs.All() {
		if !strings.HasSuffix(entry.Caller.File, "zaplog_test.go") {
			t.Errorf("%q names caller %s, want zaplog_test.go", entry.Message, entry.Caller.File)
		}
	}
	if logs.Len() != 2 {
		t.Fatalf("%d lines, want 2", logs.Len())
	}
}

// Wrap given nil panics at construction
func TestWrapNilPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Wrap(nil) did not panic")
		}
	}()
	zaplog.Wrap(nil)
}

type secretConfig struct {
	Port     int
	Password string
}

func (s secretConfig) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddInt("port", s.Port)
	return nil
}

// LogConfig logs only what the config's MarshalLogObject names, so the password never appears
func TestLogConfigUsesMarshalLogObject(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	zaplog.LogConfig[secretConfig](zap.New(core))(context.Background(), secretConfig{
		Port:     8080,
		Password: "hunter2",
	})
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("%d lines, want 1", len(entries))
	}
	logged := entries[0].ContextMap()["config"]
	want := map[string]any{
		"port": 8080,
	}
	if !reflect.DeepEqual(logged, want) {
		t.Fatalf("logged %#v, want port alone", logged)
	}
}

// New builds a working logger from the config: here the default one, writing JSON to a file
func TestNewBuildsTheConfiguredLogger(t *testing.T) {
	cfg := zaplog.DefaultConfig()
	cfg.Encoding = "json"
	path := filepath.Join(t.TempDir(), "service.log")
	cfg.OutputPaths = []string{path}
	provider := zaplog.New(cfg)
	provider.LifecycleInfo(context.Background(), "starting orders")
	provider.Flush()
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), `"msg":"starting orders"`) {
		t.Fatalf("wrote %q, want the lifecycle line", written)
	}
}

// New panics when zap rejects the config, here one with no encoding
func TestNewPanicsOnABadConfig(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil || !strings.Contains(recovered.(error).Error(), "zaplog: build logger") {
			t.Fatalf("recovered %v, want the build failure", recovered)
		}
	}()
	zaplog.New(zap.Config{})
}

// The default config never samples, omits the caller, writes the time in UTC to the millisecond
// and durations as text
func TestDefaultConfig(t *testing.T) {
	cfg := zaplog.DefaultConfig()
	if cfg.Sampling != nil || cfg.EncoderConfig.CallerKey != zapcore.OmitKey || cfg.Level.Level() != zapcore.InfoLevel {
		t.Fatalf("sampling %v, caller key %q, level %s, want none, omitted and info", cfg.Sampling, cfg.EncoderConfig.CallerKey, cfg.Level.Level())
	}
	encoder := zapcore.NewJSONEncoder(cfg.EncoderConfig)
	bst := time.FixedZone("BST", 60*60)
	entry := zapcore.Entry{Time: time.Date(2026, 10, 5, 11, 36, 39, 70_123_456, bst), Message: "configuration"}
	line, err := encoder.EncodeEntry(entry, []zapcore.Field{zap.Duration("readHeaderTimeout", 5*time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"time":"2026-10-05T10:36:39.070Z"`, `"readHeaderTimeout":"5s"`} {
		if !strings.Contains(line.String(), want) {
			t.Errorf("line %s, want %s in it", line.String(), want)
		}
	}
}

// anvil's error lines follow the logger's stack trace policy, and the trace starts at anvil's call
// site, here this test, with no zaplog frame
func TestLifecycleErrorStacktraceStartsAtTheCaller(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core, zap.AddStacktrace(zapcore.ErrorLevel))
	zaplog.Wrap(logger).LifecycleError(context.Background(), "stopped with an error", errors.New("boom"))
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("%d lines, want 1", len(entries))
	}
	stack := entries[0].Stack
	if !strings.HasPrefix(stack, "github.com/llingr/anvil-zap/zaplog_test.TestLifecycleErrorStacktraceStartsAtTheCaller") ||
		strings.Contains(stack, "zaplog.(*loggerProvider)") {
		t.Fatalf("stack %q, want it to start at this test with no zaplog frame", stack)
	}
}

// syncCounter counts the syncs Flush asks of the logger's output
type syncCounter struct {
	syncs int
}

func (s *syncCounter) Write(line []byte) (int, error) {
	return len(line), nil
}

func (s *syncCounter) Sync() error {
	s.syncs++
	return nil
}

// Flush syncs the logger, so lines buffered at exit are written
func TestFlushSyncs(t *testing.T) {
	output := &syncCounter{}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), output, zapcore.InfoLevel)
	zaplog.Wrap(zap.New(core)).Flush()
	if output.syncs != 1 {
		t.Fatalf("%d syncs, want 1", output.syncs)
	}
}
