package tflog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/internal/logging"
)

// With returns a new context.Context that has a modified logger in it which
// will include key and value as arguments in all its log output.
func With(ctx context.Context, key string, value interface{}) context.Context {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return ctx
	}
	return logging.SetProviderRootLogger(ctx, logger.With(key, value))
}

// Trace logs `msg` at the trace level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Trace(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}
	logger.Trace(msg, args...)
}

// Debug logs `msg` at the debug level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Debug(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}
	logger.Debug(msg, args...)
}

// Info logs `msg` at the info level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Info(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}
	logger.Info(msg, args...)
}

// Warn logs `msg` at the warn level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Warn(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}
	logger.Warn(msg, args...)
}

// Error logs `msg` at the error level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Error(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}
	logger.Error(msg, args...)
}
