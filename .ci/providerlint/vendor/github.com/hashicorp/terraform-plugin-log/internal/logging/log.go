package logging

import (
	"io"
	"os"
)

const (
	// DefaultProviderRootLoggerName is the default provider root logger name.
	DefaultProviderRootLoggerName string = "provider"

	// DefaultSDKRootLoggerName is the default SDK root logger name.
	DefaultSDKRootLoggerName string = "sdk"
)

// loggerKey defines context keys for locating loggers in context.Context
// it's a private type to make sure no other packages can override the key
type loggerKey string

const (
	// ProviderRootLoggerKey is the loggerKey that will hold the root
	// logger for writing logs from within provider code.
	ProviderRootLoggerKey loggerKey = "provider"

	// ProviderRootLoggerOptionsKey is the loggerKey that will hold the root
	// logger options when the root provider logger is created. This is to
	// assist creating subsystem loggers, as most options cannot be fetched and
	// a logger does not provide set methods for these options.
	ProviderRootLoggerOptionsKey loggerKey = "provider-options"

	// SDKRootLoggerKey is the loggerKey that will hold the root logger for
	// writing logs from with SDKs.
	SDKRootLoggerKey loggerKey = "sdk"

	// SDKRootLoggerOptionsKey is the loggerKey that will hold the root
	// logger options when the SDK provider logger is created. This is to
	// assist creating subsystem loggers, as most options cannot be fetched and
	// a logger does not provide set methods for these options.
	SDKRootLoggerOptionsKey loggerKey = "sdk-options"

	// SinkKey is the loggerKey that will hold the logging sink used for
	// test frameworks.
	SinkKey loggerKey = ""

	// SinkOptionsKey is the loggerKey that will hold the sink
	// logger options when the SDK provider logger is created. This is to
	// assist creating subsystem loggers, as most options cannot be fetched and
	// a logger does not provide set methods for these options.
	SinkOptionsKey loggerKey = "sink-options"

	// TFLoggerOpts is the loggerKey that will hold the LoggerOpts associated
	// with the provider root logger (at `provider.tf-logger-opts`), and the
	// provider sub-system logger (at `provider.SUBSYSTEM.tf-logger-opts`),
	// in the context.Context.
	// Note that only some LoggerOpts require to be stored this way,
	// while others use the underlying *hclog.LoggerOptions of hclog.Logger.
	TFLoggerOpts loggerKey = "tf-logger-opts"
)

// providerSubsystemLoggerKey is the loggerKey that will hold the subsystem logger
// for writing logs from within a provider subsystem.
func providerSubsystemLoggerKey(subsystem string) loggerKey {
	return ProviderRootLoggerKey + loggerKey("."+subsystem)
}

// providerRootTFLoggerOptsKey is the loggerKey that will hold
// the LoggerOpts of the provider.
func providerRootTFLoggerOptsKey() loggerKey {
	return ProviderRootLoggerKey + "." + TFLoggerOpts
}

// providerRootTFLoggerOptsKey is the loggerKey that will hold
// the LoggerOpts of a provider subsystem.
func providerSubsystemTFLoggerOptsKey(subsystem string) loggerKey {
	return providerSubsystemLoggerKey(subsystem) + "." + TFLoggerOpts
}

// providerSubsystemLoggerKey is the loggerKey that will hold the subsystem logger
// for writing logs from within an SDK subsystem.
func sdkSubsystemLoggerKey(subsystem string) loggerKey {
	return SDKRootLoggerKey + loggerKey("."+subsystem)
}

// sdkRootTFLoggerOptsKey is the loggerKey that will hold
// the LoggerOpts of the SDK.
func sdkRootTFLoggerOptsKey() loggerKey {
	return SDKRootLoggerKey + "." + TFLoggerOpts
}

// sdkSubsystemTFLoggerOptsKey is the loggerKey that will hold
// the LoggerOpts of an SDK subsystem.
func sdkSubsystemTFLoggerOptsKey(subsystem string) loggerKey {
	return sdkSubsystemLoggerKey(subsystem) + "." + TFLoggerOpts
}

var (
	// Stderr caches the original os.Stderr when the process is started.
	//
	// When go-plugin.Serve is called, it overwrites our os.Stderr with a
	// gRPC stream which Terraform ignores. This tends to be before our
	// loggers get set up, as go-plugin has no way to pass in a base
	// context, and our loggers are passed around via contexts. This leaves
	// our loggers writing to an output that is never read by anything,
	// meaning the logs get blackholed. This isn't ideal, for log output,
	// so this is our workaround: we copy stderr on init, before Serve can
	// be called, and offer an option to write to that instead of the
	// os.Stderr available at runtime.
	//
	// Ideally, this is a short-term fix until Terraform starts reading
	// from go-plugin's gRPC-streamed stderr channel, but for the moment it
	// works.
	Stderr io.Writer
)

func init() {
	Stderr = os.Stderr
}
