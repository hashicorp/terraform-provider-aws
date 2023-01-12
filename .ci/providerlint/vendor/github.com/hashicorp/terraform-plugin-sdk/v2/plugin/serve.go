package plugin

import (
	"errors"
	"log"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// The constants below are the names of the plugins that can be dispensed
	// from the plugin server.
	//
	// Deprecated: This is no longer used, but left for backwards compatibility
	// since it is exported. It will be removed in the next major version.
	ProviderPluginName = "provider"
)

// Handshake is the HandshakeConfig used to configure clients and servers.
//
// Deprecated: This is no longer used, but left for backwards compatibility
// since it is exported. It will be removed in the next major version.
var Handshake = plugin.HandshakeConfig{
	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

type ProviderFunc func() *schema.Provider
type GRPCProviderFunc func() tfprotov5.ProviderServer
type GRPCProviderV6Func func() tfprotov6.ProviderServer

// ServeOpts are the configurations to serve a plugin.
type ServeOpts struct {
	ProviderFunc ProviderFunc

	// Wrapped versions of the above plugins will automatically shimmed and
	// added to the GRPC functions when possible.
	GRPCProviderFunc GRPCProviderFunc

	GRPCProviderV6Func GRPCProviderV6Func

	// Logger is the logger that go-plugin will use.
	Logger hclog.Logger

	// Debug starts a debug server and controls its lifecycle, printing the
	// information needed for Terraform to connect to the provider to stdout.
	// os.Interrupt will be captured and used to stop the server.
	//
	// Ensure the ProviderAddr field is correctly set when this is enabled,
	// otherwise the TF_REATTACH_PROVIDERS environment variable will not
	// correctly point Terraform to the running provider binary.
	//
	// This option cannot be combined with TestConfig.
	Debug bool

	// TestConfig should only be set when the provider is being tested; it
	// will opt out of go-plugin's lifecycle management and other features,
	// and will use the supplied configuration options to control the
	// plugin's lifecycle and communicate connection information. See the
	// go-plugin GoDoc for more information.
	//
	// This option cannot be combined with Debug.
	TestConfig *plugin.ServeTestConfig

	// Set NoLogOutputOverride to not override the log output with an hclog
	// adapter. This should only be used when running the plugin in
	// acceptance tests.
	NoLogOutputOverride bool

	// UseTFLogSink is the testing.T for a test function that will turn on
	// the terraform-plugin-log logging sink.
	UseTFLogSink testing.T

	// ProviderAddr is the address of the provider under test or debugging,
	// such as registry.terraform.io/hashicorp/random. This value is used in
	// the TF_REATTACH_PROVIDERS environment variable during debugging so
	// Terraform can correctly match the provider address in the Terraform
	// configuration to the running provider binary.
	ProviderAddr string
}

// Serve serves a plugin. This function never returns and should be the final
// function called in the main function of the plugin.
func Serve(opts *ServeOpts) {
	if opts.Debug && opts.TestConfig != nil {
		log.Printf("[ERROR] Error starting provider: cannot set both Debug and TestConfig")
		return
	}

	if !opts.NoLogOutputOverride {
		// In order to allow go-plugin to correctly pass log-levels through to
		// terraform, we need to use an hclog.Logger with JSON output. We can
		// inject this into the std `log` package here, so existing providers will
		// make use of it automatically.
		logger := hclog.New(&hclog.LoggerOptions{
			// We send all output to terraform. Go-plugin will take the output and
			// pass it through another hclog.Logger on the client side where it can
			// be filtered.
			Level:      hclog.Trace,
			JSONFormat: true,
		})
		log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	}

	if opts.ProviderAddr == "" {
		opts.ProviderAddr = "provider"
	}

	var err error

	switch {
	case opts.ProviderFunc != nil && opts.GRPCProviderFunc == nil:
		opts.GRPCProviderFunc = func() tfprotov5.ProviderServer {
			return schema.NewGRPCProviderServer(opts.ProviderFunc())
		}
		err = tf5serverServe(opts)
	case opts.GRPCProviderFunc != nil:
		err = tf5serverServe(opts)
	case opts.GRPCProviderV6Func != nil:
		err = tf6serverServe(opts)
	default:
		err = errors.New("no provider server defined in ServeOpts")
	}

	if err != nil {
		log.Printf("[ERROR] Error starting provider: %s", err)
	}
}

func tf5serverServe(opts *ServeOpts) error {
	var tf5serveOpts []tf5server.ServeOpt

	if opts.Debug {
		tf5serveOpts = append(tf5serveOpts, tf5server.WithManagedDebug())
	}

	if opts.Logger != nil {
		tf5serveOpts = append(tf5serveOpts, tf5server.WithGoPluginLogger(opts.Logger))
	}

	if opts.TestConfig != nil {
		// Convert send-only channels to bi-directional channels to appease
		// the compiler. WithDebug is errantly defined to require
		// bi-directional when send-only is actually needed, which may be
		// fixed in the future so the opts.TestConfig channels can be passed
		// through directly.
		closeCh := make(chan struct{})
		reattachConfigCh := make(chan *plugin.ReattachConfig)

		go func() {
			// Always forward close channel receive, since its signaling that
			// the channel is closed.
			val := <-closeCh
			opts.TestConfig.CloseCh <- val
		}()

		go func() {
			val, ok := <-reattachConfigCh

			if ok {
				opts.TestConfig.ReattachConfigCh <- val
			}
		}()

		tf5serveOpts = append(tf5serveOpts, tf5server.WithDebug(
			opts.TestConfig.Context,
			reattachConfigCh,
			closeCh),
		)
	}

	if opts.UseTFLogSink != nil {
		tf5serveOpts = append(tf5serveOpts, tf5server.WithLoggingSink(opts.UseTFLogSink))
	}

	return tf5server.Serve(opts.ProviderAddr, opts.GRPCProviderFunc, tf5serveOpts...)
}

func tf6serverServe(opts *ServeOpts) error {
	var tf6serveOpts []tf6server.ServeOpt

	if opts.Debug {
		tf6serveOpts = append(tf6serveOpts, tf6server.WithManagedDebug())
	}

	if opts.Logger != nil {
		tf6serveOpts = append(tf6serveOpts, tf6server.WithGoPluginLogger(opts.Logger))
	}

	if opts.TestConfig != nil {
		// Convert send-only channels to bi-directional channels to appease
		// the compiler. WithDebug is errantly defined to require
		// bi-directional when send-only is actually needed, which may be
		// fixed in the future so the opts.TestConfig channels can be passed
		// through directly.
		closeCh := make(chan struct{})
		reattachConfigCh := make(chan *plugin.ReattachConfig)

		go func() {
			// Always forward close channel receive, since its signaling that
			// the channel is closed.
			val := <-closeCh
			opts.TestConfig.CloseCh <- val
		}()

		go func() {
			val, ok := <-reattachConfigCh

			if ok {
				opts.TestConfig.ReattachConfigCh <- val
			}
		}()

		tf6serveOpts = append(tf6serveOpts, tf6server.WithDebug(
			opts.TestConfig.Context,
			reattachConfigCh,
			closeCh),
		)
	}

	if opts.UseTFLogSink != nil {
		tf6serveOpts = append(tf6serveOpts, tf6server.WithLoggingSink(opts.UseTFLogSink))
	}

	return tf6server.Serve(opts.ProviderAddr, opts.GRPCProviderV6Func, tf6serveOpts...)
}
