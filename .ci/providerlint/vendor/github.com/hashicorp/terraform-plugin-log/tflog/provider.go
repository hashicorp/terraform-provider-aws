package tflog

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-log/internal/logging"
)

// SetField returns a new context.Context that has a modified logger in it which
// will include key and value as fields in all its log output.
//
// In case of the same key is used multiple times (i.e. key collision),
// the last one set is the one that gets persisted and then outputted with the logs.
func SetField(ctx context.Context, key string, value interface{}) context.Context {
	lOpts := logging.GetProviderRootTFLoggerOpts(ctx)

	lOpts = logging.WithField(key, value)(lOpts)

	return logging.SetProviderRootTFLoggerOpts(ctx, lOpts)
}

// Trace logs `msg` at the trace level to the logger in `ctx`, with optional
// `additionalFields` structured key-value fields in the log output. Fields are
// shallow merged with any defined on the logger, e.g. by the `SetField()` function,
// and across multiple maps.
func Trace(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}

	additionalArgs, shouldOmit := logging.OmitOrMask(logging.GetProviderRootTFLoggerOpts(ctx), &msg, additionalFields)
	if shouldOmit {
		return
	}

	logger.Trace(msg, additionalArgs...)
}

// Debug logs `msg` at the debug level to the logger in `ctx`, with optional
// `additionalFields` structured key-value fields in the log output. Fields are
// shallow merged with any defined on the logger, e.g. by the `SetField()` function,
// and across multiple maps.
func Debug(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}

	additionalArgs, shouldOmit := logging.OmitOrMask(logging.GetProviderRootTFLoggerOpts(ctx), &msg, additionalFields)
	if shouldOmit {
		return
	}

	logger.Debug(msg, additionalArgs...)
}

// Info logs `msg` at the info level to the logger in `ctx`, with optional
// `additionalFields` structured key-value fields in the log output. Fields are
// shallow merged with any defined on the logger, e.g. by the `SetField()` function,
// and across multiple maps.
func Info(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}

	additionalArgs, shouldOmit := logging.OmitOrMask(logging.GetProviderRootTFLoggerOpts(ctx), &msg, additionalFields)
	if shouldOmit {
		return
	}

	logger.Info(msg, additionalArgs...)
}

// Warn logs `msg` at the warn level to the logger in `ctx`, with optional
// `additionalFields` structured key-value fields in the log output. Fields are
// shallow merged with any defined on the logger, e.g. by the `SetField()` function,
// and across multiple maps.
func Warn(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}

	additionalArgs, shouldOmit := logging.OmitOrMask(logging.GetProviderRootTFLoggerOpts(ctx), &msg, additionalFields)
	if shouldOmit {
		return
	}

	logger.Warn(msg, additionalArgs...)
}

// Error logs `msg` at the error level to the logger in `ctx`, with optional
// `additionalFields` structured key-value fields in the log output. Fields are
// shallow merged with any defined on the logger, e.g. by the `SetField()` function,
// and across multiple maps.
func Error(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	logger := logging.GetProviderRootLogger(ctx)
	if logger == nil {
		// this essentially should never happen in production
		// the root logger for provider code should be injected
		// by whatever SDK the provider developer is using, so
		// really this is only likely in unit tests, at most
		// so just making this a no-op is fine
		return
	}

	additionalArgs, shouldOmit := logging.OmitOrMask(logging.GetProviderRootTFLoggerOpts(ctx), &msg, additionalFields)
	if shouldOmit {
		return
	}

	logger.Error(msg, additionalArgs...)
}

// OmitLogWithFieldKeys returns a new context.Context that has a modified logger
// that will omit to write any log when any of the given keys is found
// within its fields.
//
// Each call to this function is additive:
// the keys to omit by are added to the existing configuration.
//
// Example:
//
//   configuration = `['foo', 'baz']`
//
//   log1 = `{ msg = "...", fields = { 'foo': '...', 'bar': '...' }`  -> omitted
//   log2 = `{ msg = "...", fields = { 'bar': '...' }`                -> printed
//   log3 = `{ msg = "...", fields = { 'baz': '...', 'boo': '...' }`  -> omitted
//
func OmitLogWithFieldKeys(ctx context.Context, keys ...string) context.Context {
	lOpts := logging.GetProviderRootTFLoggerOpts(ctx)

	lOpts = logging.WithOmitLogWithFieldKeys(keys...)(lOpts)

	return logging.SetProviderRootTFLoggerOpts(ctx, lOpts)
}

// OmitLogWithMessageRegexes returns a new context.Context that has a modified logger
// that will omit to write any log that has a message matching any of the
// given *regexp.Regexp.
//
// Each call to this function is additive:
// the regexp to omit by are added to the existing configuration.
//
// Example:
//
//   configuration = `[regexp.MustCompile("(foo|bar)")]`
//
//   log1 = `{ msg = "banana apple foo", fields = {...}`     -> omitted
//   log2 = `{ msg = "pineapple mango", fields = {...}`      -> printed
//   log3 = `{ msg = "pineapple mango bar", fields = {...}`  -> omitted
//
func OmitLogWithMessageRegexes(ctx context.Context, expressions ...*regexp.Regexp) context.Context {
	lOpts := logging.GetProviderRootTFLoggerOpts(ctx)

	lOpts = logging.WithOmitLogWithMessageRegexes(expressions...)(lOpts)

	return logging.SetProviderRootTFLoggerOpts(ctx, lOpts)
}

// OmitLogWithMessageStrings  returns a new context.Context that has a modified logger
// that will omit to write any log that matches any of the given string.
//
// Each call to this function is additive:
// the string to omit by are added to the existing configuration.
//
// Example:
//
//   configuration = `['foo', 'bar']`
//
//   log1 = `{ msg = "banana apple foo", fields = {...}`     -> omitted
//   log2 = `{ msg = "pineapple mango", fields = {...}`      -> printed
//   log3 = `{ msg = "pineapple mango bar", fields = {...}`  -> omitted
//
func OmitLogWithMessageStrings(ctx context.Context, matchingStrings ...string) context.Context {
	lOpts := logging.GetProviderRootTFLoggerOpts(ctx)

	lOpts = logging.WithOmitLogWithMessageStrings(matchingStrings...)(lOpts)

	return logging.SetProviderRootTFLoggerOpts(ctx, lOpts)
}

// MaskFieldValuesWithFieldKeys returns a new context.Context that has a modified logger
// that masks (replaces) with asterisks (`***`) any field value where the
// key matches one of the given keys.
//
// Each call to this function is additive:
// the keys to mask by are added to the existing configuration.
//
// Example:
//
//   configuration = `['foo', 'baz']`
//
//   log1 = `{ msg = "...", fields = { 'foo': '***', 'bar': '...' }`  -> masked value
//   log2 = `{ msg = "...", fields = { 'bar': '...' }`                -> as-is value
//   log3 = `{ msg = "...", fields = { 'baz': '***', 'boo': '...' }`  -> masked value
//
func MaskFieldValuesWithFieldKeys(ctx context.Context, keys ...string) context.Context {
	lOpts := logging.GetProviderRootTFLoggerOpts(ctx)

	lOpts = logging.WithMaskFieldValuesWithFieldKeys(keys...)(lOpts)

	return logging.SetProviderRootTFLoggerOpts(ctx, lOpts)
}

// MaskAllFieldValuesRegexes returns a new context.Context that has a modified logger
// that masks (replaces) with asterisks (`***`) all field value substrings,
// matching one of the given *regexp.Regexp.
//
// Note that the replacement will happen, only for field values that are of type string.
//
// Each call to this function is additive:
// the regexp to mask by are added to the existing configuration.
//
// Example:
//
//   configuration = `[regexp.MustCompile("(foo|bar)")]`
//
//   log1 = `{ msg = "...", fields = { 'k1': '***', 'k2': '***', 'k3': 'baz' }`  -> masked value
//   log2 = `{ msg = "...", fields = { 'k1': 'boo', 'k2': 'far', 'k3': 'baz' }`  -> as-is value
//   log2 = `{ msg = "...", fields = { 'k1': '*** *** baz' }`                    -> masked value
//
func MaskAllFieldValuesRegexes(ctx context.Context, expressions ...*regexp.Regexp) context.Context {
	lOpts := logging.GetProviderRootTFLoggerOpts(ctx)

	lOpts = logging.WithMaskAllFieldValuesRegexes(expressions...)(lOpts)

	return logging.SetProviderRootTFLoggerOpts(ctx, lOpts)
}

// MaskAllFieldValuesStrings returns a new context.Context that has a modified logger
// that masks (replaces) with asterisks (`***`) all field value substrings,
// equal to one of the given strings.
//
// Note that the replacement will happen, only for field values that are of type string.
//
// Each call to this function is additive:
// the regexp to mask by are added to the existing configuration.
//
// Example:
//
//   configuration = `[regexp.MustCompile("(foo|bar)")]`
//
//   log1 = `{ msg = "...", fields = { 'k1': '***', 'k2': '***', 'k3': 'baz' }`  -> masked value
//   log2 = `{ msg = "...", fields = { 'k1': 'boo', 'k2': 'far', 'k3': 'baz' }`  -> as-is value
//   log2 = `{ msg = "...", fields = { 'k1': '*** *** baz' }`                    -> masked value
//
func MaskAllFieldValuesStrings(ctx context.Context, matchingStrings ...string) context.Context {
	lOpts := logging.GetProviderRootTFLoggerOpts(ctx)

	lOpts = logging.WithMaskAllFieldValuesStrings(matchingStrings...)(lOpts)

	return logging.SetProviderRootTFLoggerOpts(ctx, lOpts)
}

// MaskMessageRegexes returns a new context.Context that has a modified logger
// that masks (replaces) with asterisks (`***`) all message substrings,
// matching one of the given *regexp.Regexp.
//
// Each call to this function is additive:
// the regexp to mask by are added to the existing configuration.
//
// Example:
//
//   configuration = `[regexp.MustCompile("(foo|bar)")]`
//
//   log1 = `{ msg = "banana apple ***", fields = {...}`     -> masked portion
//   log2 = `{ msg = "pineapple mango", fields = {...}`      -> as-is
//   log3 = `{ msg = "pineapple mango ***", fields = {...}`  -> masked portion
//
func MaskMessageRegexes(ctx context.Context, expressions ...*regexp.Regexp) context.Context {
	lOpts := logging.GetProviderRootTFLoggerOpts(ctx)

	lOpts = logging.WithMaskMessageRegexes(expressions...)(lOpts)

	return logging.SetProviderRootTFLoggerOpts(ctx, lOpts)
}

// MaskMessageStrings returns a new context.Context that has a modified logger
// that masks (replace) with asterisks (`***`) all message substrings,
// equal to one of the given strings.
//
// Each call to this function is additive:
// the string to mask by are added to the existing configuration.
//
// Example:
//
//   configuration = `['foo', 'bar']`
//
//   log1 = `{ msg = "banana apple ***", fields = { 'k1': 'foo, bar, baz' }`  -> masked portion
//   log2 = `{ msg = "pineapple mango", fields = {...}`                       -> as-is
//   log3 = `{ msg = "pineapple mango ***", fields = {...}`                   -> masked portion
//
func MaskMessageStrings(ctx context.Context, matchingStrings ...string) context.Context {
	lOpts := logging.GetProviderRootTFLoggerOpts(ctx)

	lOpts = logging.WithMaskMessageStrings(matchingStrings...)(lOpts)

	return logging.SetProviderRootTFLoggerOpts(ctx, lOpts)
}

// MaskLogRegexes is a shortcut to invoke MaskMessageRegexes and MaskAllFieldValuesRegexes using the same input.
// Refer to those functions for details.
func MaskLogRegexes(ctx context.Context, expressions ...*regexp.Regexp) context.Context {
	return MaskMessageRegexes(MaskAllFieldValuesRegexes(ctx, expressions...), expressions...)
}

// MaskLogStrings is a shortcut to invoke MaskMessageStrings and MaskAllFieldValuesStrings using the same input.
// Refer to those functions for details.
func MaskLogStrings(ctx context.Context, matchingStrings ...string) context.Context {
	return MaskMessageStrings(MaskAllFieldValuesStrings(ctx, matchingStrings...), matchingStrings...)
}
