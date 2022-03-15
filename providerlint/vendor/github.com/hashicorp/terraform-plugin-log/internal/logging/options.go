package logging

import (
	"io"
	"os"

	"github.com/hashicorp/go-hclog"
)

// Option defines a modification to the configuration for a logger.
type Option func(LoggerOpts) LoggerOpts

// LoggerOpts is a collection of configuration settings for loggers.
type LoggerOpts struct {
	// Name is the name or "@module" of a logger.
	Name string

	// Level is the most verbose level that a logger will write logs for.
	Level hclog.Level

	// IncludeLocation indicates whether logs should include the location
	// of the logging statement or not.
	IncludeLocation bool

	// AdditionalLocationOffset is the number of additional stack levels to
	// skip when finding the file and line information for the log line.
	// Defaults to 1 to account for the tflog and tfsdklog logging functions.
	AdditionalLocationOffset int

	// Output dictates where logs are written to. Output should only ever
	// be set by tflog or tfsdklog, never by SDK authors or provider
	// developers. Where logs get written to is complex and delicate and
	// requires a deep understanding of Terraform's architecture, and it's
	// easy to mess up on accident.
	Output io.Writer

	// IncludeTime indicates whether logs should incude the time they were
	// written or not. It should only be set to true when testing tflog or
	// tfsdklog; providers and SDKs should always include the time logs
	// were written as part of the log.
	IncludeTime bool
}

// ApplyLoggerOpts generates a LoggerOpts out of a list of Option
// implementations. By default, AdditionalLocationOffset is 1, IncludeLocation
// is true, IncludeTime is true, and Output is os.Stderr.
func ApplyLoggerOpts(opts ...Option) LoggerOpts {
	// set some defaults
	l := LoggerOpts{
		AdditionalLocationOffset: 1,
		IncludeLocation:          true,
		IncludeTime:              true,
		Output:                   os.Stderr,
	}
	for _, opt := range opts {
		l = opt(l)
	}
	return l
}

// WithAdditionalLocationOffset sets the WithAdditionalLocationOffset
// configuration option, allowing implementations to fix location information
// when implementing helper functions. The default offset of 1 is automatically
// added to the provided value to account for the tflog and tfsdk logging
// functions.
func WithAdditionalLocationOffset(additionalLocationOffset int) Option {
	return func(l LoggerOpts) LoggerOpts {
		l.AdditionalLocationOffset = additionalLocationOffset + 1
		return l
	}
}

// WithOutput sets the Output configuration option, controlling where logs get
// written to. This is mostly used for testing (to write to os.Stdout, so the
// test framework can compare it against the example output) and as a helper
// when implementing safe, specific output strategies in tfsdklog.
func WithOutput(output io.Writer) Option {
	return func(l LoggerOpts) LoggerOpts {
		l.Output = output
		return l
	}
}

// WithoutLocation disables the location included with logging statements. It
// should only ever be used to make log output deterministic when testing
// terraform-plugin-log.
func WithoutLocation() Option {
	return func(l LoggerOpts) LoggerOpts {
		l.IncludeLocation = false
		return l
	}
}

// WithoutTimestamp disables the timestamp included with logging statements. It
// should only ever be used to make log output deterministic when testing
// terraform-plugin-log.
func WithoutTimestamp() Option {
	return func(l LoggerOpts) LoggerOpts {
		l.IncludeTime = false
		return l
	}
}
