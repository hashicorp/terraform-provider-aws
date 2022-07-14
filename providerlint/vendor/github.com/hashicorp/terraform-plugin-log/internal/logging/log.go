package logging

import (
	"io"
	"os"
)

const (
	// Default provider root logger name.
	DefaultProviderRootLoggerName string = "provider"

	// Default SDK root logger name.
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
)

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
