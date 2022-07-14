package tfsdklog

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-log/internal/hclogutils"
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
// level and additional location offset of the logger.
func NewSubsystem(ctx context.Context, subsystem string, options ...logging.Option) context.Context {
	logger := logging.GetSDKRootLogger(ctx)

	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever  the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return ctx
	}

	loggerOptions := logging.GetSDKRootLoggerOptions(ctx)
	opts := logging.ApplyLoggerOpts(options...)

	// On the off-chance that the logger options are not available, fallback
	// to creating a named logger. This will preserve the root logger options,
	// but cannot make changes beyond the level due to the hclog.Logger
	// interface.
	if loggerOptions == nil {
		subLogger := logger.Named(subsystem)

		if opts.AdditionalLocationOffset != 1 {
			logger.Warn("Unable to create logging subsystem with AdditionalLocationOffset due to missing root logger options")
		}

		if opts.Level != hclog.NoLevel {
			subLogger.SetLevel(opts.Level)
		}

		return logging.SetSDKSubsystemLogger(ctx, subsystem, subLogger)
	}

	subLoggerOptions := hclogutils.LoggerOptionsCopy(loggerOptions)
	subLoggerOptions.Name = subLoggerOptions.Name + "." + subsystem

	if opts.AdditionalLocationOffset != 1 {
		subLoggerOptions.AdditionalLocationOffset = opts.AdditionalLocationOffset
	}

	if opts.Level != hclog.NoLevel {
		subLoggerOptions.Level = opts.Level
	}

	subLogger := hclog.New(subLoggerOptions)

	if opts.IncludeRootFields {
		subLogger = subLogger.With(logger.ImpliedArgs()...)
	}

	return logging.SetSDKSubsystemLogger(ctx, subsystem, subLogger)
}

// SubsystemWith returns a new context.Context that has a modified logger for
// the specified subsystem in it which will include key and value as arguments
// in all its log output.
func SubsystemWith(ctx context.Context, subsystem, key string, value interface{}) context.Context {
	logger := logging.GetSDKSubsystemLogger(ctx, subsystem)
	if logger == nil {
		if logging.GetSDKRootLogger(ctx) == nil {
			// logging isn't set up, nothing we can do, just silently fail
			// this should basically never happen in production
			return ctx
		}
		// create a new logger if one doesn't exist
		logger = logging.GetSDKSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewSDKSubsystemLoggerWarning)
	}
	return logging.SetSDKSubsystemLogger(ctx, subsystem, logger.With(key, value))
}

// SubsystemTrace logs `msg` at the trace level to the subsystem logger
// specified in `ctx`, with optional `additionalFields` structured key-value
// fields in the log output. Fields are shallow merged with any defined on the
// subsystem logger, e.g. by the `SubsystemWith()` function, and across
// multiple maps.
func SubsystemTrace(ctx context.Context, subsystem, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetSDKSubsystemLogger(ctx, subsystem)
	if logger == nil {
		if logging.GetSDKRootLogger(ctx) == nil {
			// logging isn't set up, nothing we can do, just silently fail
			// this should basically never happen in production
			return
		}
		// create a new logger if one doesn't exist
		logger = logging.GetSDKSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewSDKSubsystemLoggerWarning)
	}
	logger.Trace(msg, hclogutils.MapsToArgs(additionalFields...)...)
}

// SubsystemDebug logs `msg` at the debug level to the subsystem logger
// specified in `ctx`, with optional `additionalFields` structured key-value
// fields in the log output. Fields are shallow merged with any defined on the
// subsystem logger, e.g. by the `SubsystemWith()` function, and across
// multiple maps.
func SubsystemDebug(ctx context.Context, subsystem, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetSDKSubsystemLogger(ctx, subsystem)
	if logger == nil {
		if logging.GetSDKRootLogger(ctx) == nil {
			// logging isn't set up, nothing we can do, just silently fail
			// this should basically never happen in production
			return
		}
		// create a new logger if one doesn't exist
		logger = logging.GetSDKSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewSDKSubsystemLoggerWarning)
	}
	logger.Debug(msg, hclogutils.MapsToArgs(additionalFields...)...)
}

// SubsystemInfo logs `msg` at the info level to the subsystem logger
// specified in `ctx`, with optional `additionalFields` structured key-value
// fields in the log output. Fields are shallow merged with any defined on the
// subsystem logger, e.g. by the `SubsystemWith()` function, and across
// multiple maps.
func SubsystemInfo(ctx context.Context, subsystem, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetSDKSubsystemLogger(ctx, subsystem)
	if logger == nil {
		if logging.GetSDKRootLogger(ctx) == nil {
			// logging isn't set up, nothing we can do, just silently fail
			// this should basically never happen in production
			return
		}
		// create a new logger if one doesn't exist
		logger = logging.GetSDKSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewSDKSubsystemLoggerWarning)
	}
	logger.Info(msg, hclogutils.MapsToArgs(additionalFields...)...)
}

// SubsystemWarn logs `msg` at the warn level to the subsystem logger
// specified in `ctx`, with optional `additionalFields` structured key-value
// fields in the log output. Fields are shallow merged with any defined on the
// subsystem logger, e.g. by the `SubsystemWith()` function, and across
// multiple maps.
func SubsystemWarn(ctx context.Context, subsystem, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetSDKSubsystemLogger(ctx, subsystem)
	if logger == nil {
		if logging.GetSDKRootLogger(ctx) == nil {
			// logging isn't set up, nothing we can do, just silently fail
			// this should basically never happen in production
			return
		}
		// create a new logger if one doesn't exist
		logger = logging.GetSDKSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewSDKSubsystemLoggerWarning)
	}
	logger.Warn(msg, hclogutils.MapsToArgs(additionalFields...)...)
}

// SubsystemError logs `msg` at the error level to the subsystem logger
// specified in `ctx`, with optional `additionalFields` structured key-value
// fields in the log output. Fields are shallow merged with any defined on the
// subsystem logger, e.g. by the `SubsystemWith()` function, and across
// multiple maps.
func SubsystemError(ctx context.Context, subsystem, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetSDKSubsystemLogger(ctx, subsystem)
	if logger == nil {
		if logging.GetSDKRootLogger(ctx) == nil {
			// logging isn't set up, nothing we can do, just silently fail
			// this should basically never happen in production
			return
		}
		// create a new logger if one doesn't exist
		logger = logging.GetSDKSubsystemLogger(NewSubsystem(ctx, subsystem), subsystem).With("new_logger_warning", logging.NewSDKSubsystemLoggerWarning)
	}
	logger.Error(msg, hclogutils.MapsToArgs(additionalFields...)...)
}
