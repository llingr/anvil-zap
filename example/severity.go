// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"go.uber.org/zap/zapcore"
)

// encodeSeverity writes zap's level as Google Cloud Logging's LogSeverity
func encodeSeverity(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
	switch level {
	case zapcore.DebugLevel:
		encoder.AppendString("DEBUG")
	case zapcore.InfoLevel:
		encoder.AppendString("INFO")
	case zapcore.WarnLevel:
		encoder.AppendString("WARNING")
	case zapcore.ErrorLevel:
		encoder.AppendString("ERROR")
	case zapcore.DPanicLevel:
		encoder.AppendString("CRITICAL")
	case zapcore.PanicLevel:
		encoder.AppendString("ALERT")
	case zapcore.FatalLevel:
		encoder.AppendString("EMERGENCY")
	default:
		encoder.AppendString("DEFAULT")
	}
}
