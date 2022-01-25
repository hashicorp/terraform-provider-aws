package tfsdklog

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-log/internal/logging"
)

// NewRootSDKLogger returns a new context.Context that contains an SDK logger
// configured with the passed options.
func NewRootSDKLogger(ctx context.Context, options ...logging.Option) context.Context {
	opts := logging.ApplyLoggerOpts(options...)
	if opts.Name == "" {
		opts.Name = "sdk"
	}
	if sink := getSink(ctx); sink != nil {
		logger := sink.Named(opts.Name)
		if opts.Level != hclog.NoLevel {
			logger.SetLevel(opts.Level)
		}
		return logging.SetSDKRootLogger(ctx, logger)
	}
	if opts.Level == hclog.NoLevel {
		opts.Level = hclog.Trace
	}
	loggerOptions := &hclog.LoggerOptions{
		Name:                     opts.Name,
		Level:                    opts.Level,
		JSONFormat:               true,
		IndependentLevels:        true,
		IncludeLocation:          opts.IncludeLocation,
		DisableTime:              !opts.IncludeTime,
		Output:                   opts.Output,
		AdditionalLocationOffset: 1,
	}
	return logging.SetSDKRootLogger(ctx, hclog.New(loggerOptions))
}

// NewRootProviderLogger returns a new context.Context that contains a provider
// logger configured with the passed options.
func NewRootProviderLogger(ctx context.Context, options ...logging.Option) context.Context {
	opts := logging.ApplyLoggerOpts(options...)
	if opts.Name == "" {
		opts.Name = "provider"
	}
	if sink := getSink(ctx); sink != nil {
		logger := sink.Named(opts.Name)
		if opts.Level != hclog.NoLevel {
			logger.SetLevel(opts.Level)
		}
		return logging.SetProviderRootLogger(ctx, logger)
	}
	if opts.Level == hclog.NoLevel {
		opts.Level = hclog.Trace
	}
	loggerOptions := &hclog.LoggerOptions{
		Name:                     opts.Name,
		Level:                    opts.Level,
		JSONFormat:               true,
		IndependentLevels:        true,
		IncludeLocation:          opts.IncludeLocation,
		DisableTime:              !opts.IncludeTime,
		Output:                   opts.Output,
		AdditionalLocationOffset: 1,
	}
	return logging.SetProviderRootLogger(ctx, hclog.New(loggerOptions))
}

// With returns a new context.Context that has a modified logger in it which
// will include key and value as arguments in all its log output.
func With(ctx context.Context, key string, value interface{}) context.Context {
	logger := logging.GetSDKRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production the root
		// logger for  code should be injected by the  in
		// question, so really this is only likely in unit tests, at
		// most so just making this a no-op is fine
		return ctx
	}
	return logging.SetSDKRootLogger(ctx, logger.With(key, value))
}

// Trace logs `msg` at the trace level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Trace(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetSDKRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production the root
		// logger for  code should be injected by the  in
		// question, so really this is only likely in unit tests, at
		// most so just making this a no-op is fine
		return
	}
	logger.Trace(msg, args...)
}

// Debug logs `msg` at the debug level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Debug(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetSDKRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production the root
		// logger for  code should be injected by the  in
		// question, so really this is only likely in unit tests, at
		// most so just making this a no-op is fine
		return
	}
	logger.Debug(msg, args...)
}

// Info logs `msg` at the info level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Info(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetSDKRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production the root
		// logger for  code should be injected by the  in
		// question, so really this is only likely in unit tests, at
		// most so just making this a no-op is fine
		return
	}
	logger.Info(msg, args...)
}

// Warn logs `msg` at the warn level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Warn(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetSDKRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production the root
		// logger for  code should be injected by the  in
		// question, so really this is only likely in unit tests, at
		// most so just making this a no-op is fine
		return
	}
	logger.Warn(msg, args...)
}

// Error logs `msg` at the error level to the logger in `ctx`, with `args` as
// structured arguments in the log output. `args` is expected to be pairs of
// key and value.
func Error(ctx context.Context, msg string, args ...interface{}) {
	logger := logging.GetSDKRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production the root
		// logger for  code should be injected by the  in
		// question, so really this is only likely in unit tests, at
		// most so just making this a no-op is fine
		return
	}
	logger.Error(msg, args...)
}
