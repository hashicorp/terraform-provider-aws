package tflog

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-log/internal/logging"
)

// NewSubsystem returns a new context.Context that contains a subsystem logger
// configured with the passed options, named after the subsystem argument.
//
// Subsystem loggers allow different areas of a plugin codebase to use
// different logging levels, giving developers more fine-grained control over
// what is logging and with what verbosity. They're best utilized for logical
// concerns that are sometimes helpful to log, but may generate unwanted noise
// at other times.
//
// The only Options supported for subsystems are the Options for setting the
// level of the logger.
func NewSubsystem(ctx context.Context, subsystem string, options ...logging.Option) context.Context {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return ctx
	}
	subLogger := logger.Named(subsystem)
	opts := logging.ApplyLoggerOpts(options...)
	if opts.Level != hclog.NoLevel {
		subLogger.SetLevel(opts.Level)
	}
	return logging.SetProviderSubsystemLogger(ctx, subsystem, subLogger)
}

// SubsystemWith returns a new context.Context that has a modified logger for
// the specified subsystem in it which will include key and value as arguments
// in all its log output.
func SubsystemWith(ctx context.Context, subsystem, key string, value interface{}) context.Context {
	logger := logging.GetProviderSubsystemLogger(ctx, subsystem)
	if logger == nil {
		// create a new logger if one doesn't exist
		logger = logging.GetProviderSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewProviderSubsystemLoggerWarning)
	}
	return logging.SetProviderSubsystemLogger(ctx, subsystem, logger.With(key, value))
}

// SubsystemTrace logs `msg` at the trace level to the subsystem logger
// specified in `ctx`, with `args` as structured arguments in the log output.
// `args` is expected to be pairs of key and value.
func SubsystemTrace(ctx context.Context, subsystem, msg string, args ...interface{}) {
	logger := logging.GetProviderSubsystemLogger(ctx, subsystem)
	if logger == nil {
		// create a new logger if one doesn't exist
		logger = logging.GetProviderSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewProviderSubsystemLoggerWarning)
	}
	logger.Trace(msg, args...)
}

// SubsystemDebug logs `msg` at the debug level to the subsystem logger
// specified in `ctx`, with `args` as structured arguments in the log output.
// `args` is expected to be pairs of key and value.
func SubsystemDebug(ctx context.Context, subsystem, msg string, args ...interface{}) {
	logger := logging.GetProviderSubsystemLogger(ctx, subsystem)
	if logger == nil {
		// create a new logger if one doesn't exist
		logger = logging.GetProviderSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewProviderSubsystemLoggerWarning)
	}
	logger.Debug(msg, args...)
}

// SubsystemInfo logs `msg` at the info level to the subsystem logger
// specified in `ctx`, with `args` as structured arguments in the log output.
// `args` is expected to be pairs of key and value.
func SubsystemInfo(ctx context.Context, subsystem, msg string, args ...interface{}) {
	logger := logging.GetProviderSubsystemLogger(ctx, subsystem)
	if logger == nil {
		// create a new logger if one doesn't exist
		logger = logging.GetProviderSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewProviderSubsystemLoggerWarning)
	}
	logger.Info(msg, args...)
}

// SubsystemWarn logs `msg` at the warn level to the subsystem logger
// specified in `ctx`, with `args` as structured arguments in the log output.
// `args` is expected to be pairs of key and value.
func SubsystemWarn(ctx context.Context, subsystem, msg string, args ...interface{}) {
	logger := logging.GetProviderSubsystemLogger(ctx, subsystem)
	if logger == nil {
		// create a new logger if one doesn't exist
		logger = logging.GetProviderSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewProviderSubsystemLoggerWarning)
	}
	logger.Warn(msg, args...)
}

// SubsystemError logs `msg` at the error level to the subsystem logger
// specified in `ctx`, with `args` as structured arguments in the log output.
// `args` is expected to be pairs of key and value.
func SubsystemError(ctx context.Context, subsystem, msg string, args ...interface{}) {
	logger := logging.GetProviderSubsystemLogger(ctx, subsystem)
	if logger == nil {
		// create a new logger if one doesn't exist
		logger = logging.GetProviderSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewProviderSubsystemLoggerWarning)
	}
	logger.Error(msg, args...)
}
