package logging

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

// GetSink returns the sink logger used for writing logs.
// If no sink logger has been created, it will return nil.
func GetSink(ctx context.Context) hclog.Logger {
	logger := ctx.Value(SinkKey)
	if logger == nil {
		return nil
	}

	hclogger, ok := logger.(hclog.Logger)
	if !ok {
		return nil
	}

	return hclogger
}

// GetSinkOptions returns the root logger options used for
// creating the root SDK logger. If the root logger has not been created or
// the options are not present, it will return nil.
func GetSinkOptions(ctx context.Context) *hclog.LoggerOptions {
	if GetSink(ctx) == nil {
		return nil
	}

	loggerOptionsRaw := ctx.Value(SinkOptionsKey)

	if loggerOptionsRaw == nil {
		return nil
	}

	loggerOptions, ok := loggerOptionsRaw.(*hclog.LoggerOptions)

	if !ok {
		return nil
	}

	return loggerOptions
}

// SetSink sets `logger` as the sink logger used for writing logs.
func SetSink(ctx context.Context, logger hclog.Logger) context.Context {
	return context.WithValue(ctx, SinkKey, logger)
}

// SetSinkOptions sets `loggerOptions` as the root logger options
// used for creating the SDK root logger.
func SetSinkOptions(ctx context.Context, loggerOptions *hclog.LoggerOptions) context.Context {
	return context.WithValue(ctx, SinkOptionsKey, loggerOptions)
}
