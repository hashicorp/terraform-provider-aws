package logging

import (
	"io"
	"os"
	"regexp"

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

	// IncludeTime indicates whether logs should include the time they were
	// written or not. It should only be set to true when testing tflog or
	// tfsdklog; providers and SDKs should always include the time logs
	// were written as part of the log.
	IncludeTime bool

	// Fields indicates the key/value pairs to be added to each of its log output.
	Fields map[string]interface{}

	// IncludeRootFields indicates whether a new subsystem logger should
	// copy existing fields from the root logger. This is only performed
	// at the time of new subsystem creation.
	IncludeRootFields bool

	// OmitLogWithFieldKeys indicates that the logger should omit to write
	// any log when any of the given keys is found within the fields.
	//
	// Example:
	//
	//   OmitLogWithFieldKeys = `['foo', 'baz']`
	//
	//   log1 = `{ msg = "...", fields = { 'foo', '...', 'bar', '...' }`   -> omitted
	//   log2 = `{ msg = "...", fields = { 'bar', '...' }`                 -> printed
	//   log3 = `{ msg = "...", fields = { 'baz`', '...', 'boo', '...' }`  -> omitted
	//
	OmitLogWithFieldKeys []string

	// OmitLogWithMessageRegexes indicates that the logger should omit to write
	// any log that matches any of the given *regexp.Regexp.
	//
	// Example:
	//
	//   OmitLogWithMessageRegexes = `[regexp.MustCompile("(foo|bar)")]`
	//
	//   log1 = `{ msg = "banana apple foo", fields = {...}`     -> omitted
	//   log2 = `{ msg = "pineapple mango", fields = {...}`      -> printed
	//   log3 = `{ msg = "pineapple mango bar", fields = {...}`  -> omitted
	//
	OmitLogWithMessageRegexes []*regexp.Regexp

	// OmitLogWithMessageStrings indicates that the logger should omit to write
	// any log that matches any of the given string.
	//
	// Example:
	//
	//   OmitLogWithMessageStrings = `['foo', 'bar']`
	//
	//   log1 = `{ msg = "banana apple foo", fields = {...}`     -> omitted
	//   log2 = `{ msg = "pineapple mango", fields = {...}`      -> printed
	//   log3 = `{ msg = "pineapple mango bar", fields = {...}`  -> omitted
	//
	OmitLogWithMessageStrings []string

	// MaskFieldValuesWithFieldKeys indicates that the logger should mask with asterisks (`*`)
	// any field value where the key matches one of the given keys.
	//
	// Example:
	//
	//   MaskFieldValuesWithFieldKeys = `['foo', 'baz']`
	//
	//   log1 = `{ msg = "...", fields = { 'foo', '***', 'bar', '...' }`   -> masked value
	//   log2 = `{ msg = "...", fields = { 'bar', '...' }`                 -> as-is value
	//   log3 = `{ msg = "...", fields = { 'baz`', '***', 'boo', '...' }`  -> masked value
	//
	MaskFieldValuesWithFieldKeys []string

	// MaskMessageRegexes indicates that the logger should replace, within
	// a log message, the portion matching one of the given *regexp.Regexp.
	//
	// Example:
	//
	//   MaskMessageRegexes = `[regexp.MustCompile("(foo|bar)")]`
	//
	//   log1 = `{ msg = "banana apple ***", fields = {...}`     -> masked portion
	//   log2 = `{ msg = "pineapple mango", fields = {...}`      -> as-is
	//   log3 = `{ msg = "pineapple mango ***", fields = {...}`  -> masked portion
	//
	MaskMessageRegexes []*regexp.Regexp

	// MaskMessageStrings indicates that the logger should replace, within
	// a log message, the portion matching one of the given strings.
	//
	// Example:
	//
	//   MaskMessageStrings = `['foo', 'bar']`
	//
	//   log1 = `{ msg = "banana apple ***", fields = {...}`     -> masked portion
	//   log2 = `{ msg = "pineapple mango", fields = {...}`      -> as-is
	//   log3 = `{ msg = "pineapple mango ***", fields = {...}`  -> masked portion
	//
	MaskMessageStrings []string
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

// WithField sets the provided key/value pair, onto the LoggerOpts.Fields field.
//
// Behind the scene, fields are stored in a map[string]interface{}:
// this means that in case the same key is used multiple times (key collision),
// the last one set is the one that gets persisted and then outputted with the logs.
func WithField(key string, value interface{}) Option {
	return func(l LoggerOpts) LoggerOpts {
		// Lazily create this map, on first assignment
		if l.Fields == nil {
			l.Fields = make(map[string]interface{})
		}

		l.Fields[key] = value
		return l
	}
}

// WithFields sets all the provided key/value pairs, onto the LoggerOpts.Fields field.
//
// Behind the scene, fields are stored in a map[string]interface{}:
// this means that in case the same key is used multiple times (key collision),
// the last one set is the one that gets persisted and then outputted with the logs.
func WithFields(fields map[string]interface{}) Option {
	return func(l LoggerOpts) LoggerOpts {
		// Lazily create this map, on first assignment
		if l.Fields == nil {
			l.Fields = make(map[string]interface{})
		}

		for k, v := range fields {
			l.Fields[k] = v
		}

		return l
	}
}

// WithRootFields enables the copying of root logger fields to a new subsystem
// logger during creation.
func WithRootFields() Option {
	return func(l LoggerOpts) LoggerOpts {
		l.IncludeRootFields = true
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

// WithOmitLogWithFieldKeys appends keys to the LoggerOpts.OmitLogWithFieldKeys field.
func WithOmitLogWithFieldKeys(keys ...string) Option {
	return func(l LoggerOpts) LoggerOpts {
		l.OmitLogWithFieldKeys = append(l.OmitLogWithFieldKeys, keys...)
		return l
	}
}

// WithOmitLogWithMessageRegexes appends *regexp.Regexp to the LoggerOpts.OmitLogWithMessageRegexes field.
func WithOmitLogWithMessageRegexes(expressions ...*regexp.Regexp) Option {
	return func(l LoggerOpts) LoggerOpts {
		l.OmitLogWithMessageRegexes = append(l.OmitLogWithMessageRegexes, expressions...)
		return l
	}
}

// WithOmitLogWithMessageStrings appends string to the LoggerOpts.OmitLogWithMessageStrings field.
func WithOmitLogWithMessageStrings(matchingStrings ...string) Option {
	return func(l LoggerOpts) LoggerOpts {
		l.OmitLogWithMessageStrings = append(l.OmitLogWithMessageStrings, matchingStrings...)
		return l
	}
}

// WithMaskFieldValuesWithFieldKeys appends keys to the LoggerOpts.MaskFieldValuesWithFieldKeys field.
func WithMaskFieldValuesWithFieldKeys(keys ...string) Option {
	return func(l LoggerOpts) LoggerOpts {
		l.MaskFieldValuesWithFieldKeys = append(l.MaskFieldValuesWithFieldKeys, keys...)
		return l
	}
}

// WithMaskMessageRegexes appends *regexp.Regexp to the LoggerOpts.MaskMessageRegexes field.
func WithMaskMessageRegexes(expressions ...*regexp.Regexp) Option {
	return func(l LoggerOpts) LoggerOpts {
		l.MaskMessageRegexes = append(l.MaskMessageRegexes, expressions...)
		return l
	}
}

// WithMaskMessageStrings appends string to the LoggerOpts.MaskMessageStrings field.
func WithMaskMessageStrings(matchingStrings ...string) Option {
	return func(l LoggerOpts) LoggerOpts {
		l.MaskMessageStrings = append(l.MaskMessageStrings, matchingStrings...)
		return l
	}
}
