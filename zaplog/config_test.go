// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

package zaplog_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/llingr/anvil"
	"github.com/llingr/anvil-zap/zaplog"
)

const password = "hunter2"

// valueMasked masks the password with a value receiver
type valueMasked struct {
	Username string
	Password string
}

func (v valueMasked) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("username", v.Username)
	encoder.AddString("password", "****")
	return nil
}

// pointerMasked masks the password with a pointer receiver
type pointerMasked struct {
	Username string
	Password string
}

func (p *pointerMasked) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("username", p.Username)
	encoder.AddString("password", "****")
	return nil
}

// unmasked has no MarshalLogObject, so only reflection could log it
type unmasked struct {
	Username string
	Password string
}

// stringer exposes the password through String, which reflection-free encoders also use
type stringer struct {
	Username string
	Password string
}

func (s stringer) String() string {
	return s.Username + ":" + s.Password
}

// failingMarshaler writes the username, then fails
type failingMarshaler struct {
	Username string
	Password string
}

func (f failingMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("username", f.Username)
	return errors.New("cannot encode")
}

// mapMasked is a map config masking its password key
type mapMasked map[string]string

func (m mapMasked) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	for key, value := range m {
		if key == "password" {
			value = "****"
		}
		encoder.AddString(key, value)
	}
	return nil
}

// logConfig passes config to LogConfig through a JSON encoder, returning what was written and
// any panic
func logConfig(t *testing.T, config any) (written string, panicked any) {
	t.Helper()
	var output bytes.Buffer
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zapcore.InfoLevel)
	provider, ok := zaplog.Wrap(zap.New(core)).(anvil.ConfigLogger)
	if !ok {
		t.Fatal("provider does not implement LogConfig")
	}
	defer func() {
		panicked = recover()
		written = output.String()
	}()
	provider.LogConfig(context.Background(), "configuration loaded in 1ms", config)
	return
}

// Whatever the config's type, receiver or nil-ness, the password never reaches the log, nothing
// panics, and the config is logged only through a MarshalLogObject
func TestLogConfigNeverLogsThePassword(t *testing.T) {
	var nilValueMasked *valueMasked
	var nilPointerMasked *pointerMasked
	var nilMap mapMasked
	var asInterface zapcore.ObjectMarshaler = &pointerMasked{"orders", password}
	var nilInterface zapcore.ObjectMarshaler
	const (
		masked = `"password":"****"`
		absent = "" // no "config" field
	)
	cases := []struct {
		name   string
		config any
		want   string // in the line, or absent
	}{
		{"value receiver, value", valueMasked{"orders", password}, masked},
		{"value receiver, pointer", &valueMasked{"orders", password}, masked},
		{"value receiver, nil pointer", nilValueMasked, absent},
		{"pointer receiver, value", pointerMasked{"orders", password}, absent},
		{"pointer receiver, pointer", &pointerMasked{"orders", password}, masked},
		{"pointer receiver, nil pointer", nilPointerMasked, absent},
		{"pointer to the interface", &asInterface, absent},
		{"pointer to a nil interface", &nilInterface, absent},
		{"no marshaler, value", unmasked{"orders", password}, absent},
		{"no marshaler, pointer", &unmasked{"orders", password}, absent},
		{"stringer, value", stringer{"orders", password}, absent},
		{"marshaler failing, value", failingMarshaler{"orders", password}, `"config":{"username":"orders"},"configError":"cannot encode"`},
		{"map marshaler", mapMasked{"username": "orders", "password": password}, masked},
		{"map marshaler, nil", nilMap, absent},
		{"plain map", map[string]string{"username": "orders", "password": password}, absent},
		{"plain string", "orders:" + password, absent},
		{"untyped nil", nil, absent},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			written, panicked := logConfig(t, test.config)
			t.Logf("wrote %s", strings.TrimSpace(written))
			if panicked != nil {
				t.Fatalf("panicked: %v", panicked)
			}
			if strings.Contains(written, password) {
				t.Fatalf("the password was logged: %s", written)
			}
			if !strings.Contains(written, `"msg":"configuration loaded in 1ms"`) {
				t.Fatalf("wrote %q, want the loaded line", written)
			}
			if test.want == absent && strings.Contains(written, `"config":`) {
				t.Fatalf("wrote %q, want no config field", written)
			}
			if test.want != absent && !strings.Contains(written, test.want) {
				t.Fatalf("wrote %q, want %s in it", written, test.want)
			}
		})
	}
}

// A config whose marshaler delegates to a nested marshaler keeps the nested password masked
func TestLogConfigNestedMarshalerMasks(t *testing.T) {
	credentials := valueMasked{"orders", password}
	written, panicked := logConfig(t, marshalerFunc(func(encoder zapcore.ObjectEncoder) error {
		return encoder.AddObject("credentials", credentials)
	}))
	if panicked != nil || strings.Contains(written, password) || !strings.Contains(written, `"password":"****"`) {
		t.Fatalf("wrote %q, panicked %v, want the nested password masked", written, panicked)
	}
}

// marshalerFunc adapts a function to zapcore.ObjectMarshaler
type marshalerFunc func(zapcore.ObjectEncoder) error

func (m marshalerFunc) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	return m(encoder)
}
