package tf6server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-go/internal/logging"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/fromproto"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tf6serverlogging"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/toproto"
	"google.golang.org/grpc"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
	testing "github.com/mitchellh/go-testing-interface"
)

const (
	// protocolVersionMajor represents the major version number of the protocol
	// being served. This is used during the plugin handshake to validate the
	// server and client are compatible.
	//
	// In the future, it may be possible to include this information directly
	// in the protocol buffers rather than recreating a constant here.
	protocolVersionMajor uint = 6

	// protocolVersionMinor represents the minor version number of the protocol
	// being served. Backwards compatible additions are possible in the
	// protocol definitions, which is when this may be increased. While it is
	// not used in plugin negotiation, it can be helpful to include this value
	// for debugging, such as in logs.
	//
	// In the future, it may be possible to include this information directly
	// in the protocol buffers rather than recreating a constant here.
	protocolVersionMinor uint = 3
)

// protocolVersion represents the combined major and minor version numbers of
// the protocol being served.
var protocolVersion string = fmt.Sprintf("%d.%d", protocolVersionMajor, protocolVersionMinor)

const (
	// envTfReattachProviders is the environment variable used by Terraform CLI
	// to directly connect to already running provider processes, such as those
	// being inspected by debugging processes. When connecting to providers in
	// this manner, Terraform CLI disables certain plugin handshake checks and
	// will not stop the provider process.
	envTfReattachProviders = "TF_REATTACH_PROVIDERS"
)

const (
	// grpcMaxMessageSize is the maximum gRPC send and receive message sizes
	// for the server.
	//
	// This 256MB value is arbitrarily raised from the default message sizes of
	// 4MB to account for advanced use cases, but arbitrarily lowered from
	// MaxInt32 (or similar) to prevent incorrect server implementations from
	// exhausting resources in common execution environments. Receiving a gRPC
	// message size error is preferable for troubleshooting over determining
	// why an execution environment may have terminated the process via its
	// memory management processes, such as oom-killer on Linux.
	//
	// This value is kept as constant over allowing server configurability
	// since there are many factors that influence message size, such as
	// Terraform configuration and state data. If larger message size use
	// cases appear, other gRPC options should be explored, such as
	// implementing streaming RPCs and messages.
	grpcMaxMessageSize = 256 << 20
)

// ServeOpt is an interface for defining options that can be passed to the
// Serve function. Each implementation modifies the ServeConfig being
// generated. A slice of ServeOpts then, cumulatively applied, render a full
// ServeConfig.
type ServeOpt interface {
	ApplyServeOpt(*ServeConfig) error
}

// ServeConfig contains the configured options for how a provider should be
// served.
type ServeConfig struct {
	logger       hclog.Logger
	debugCtx     context.Context
	debugCh      chan *plugin.ReattachConfig
	debugCloseCh chan struct{}

	managedDebug                      bool
	managedDebugReattachConfigTimeout time.Duration
	managedDebugStopSignals           []os.Signal

	disableLogInitStderr bool
	disableLogLocation   bool
	useLoggingSink       testing.T
	envVar               string
}

type serveConfigFunc func(*ServeConfig) error

func (s serveConfigFunc) ApplyServeOpt(in *ServeConfig) error {
	return s(in)
}

// WithDebug returns a ServeOpt that will set the server into debug mode, using
// the passed options to populate the go-plugin ServeTestConfig.
//
// This is an advanced ServeOpt that assumes the caller will fully manage the
// reattach configuration and server lifecycle. Refer to WithManagedDebug for a
// ServeOpt that handles common use cases, such as implementing provider main
// functions.
func WithDebug(ctx context.Context, config chan *plugin.ReattachConfig, closeCh chan struct{}) ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		if in.managedDebug {
			return errors.New("cannot set both WithDebug and WithManagedDebug")
		}

		in.debugCtx = ctx
		in.debugCh = config
		in.debugCloseCh = closeCh
		return nil
	})
}

// WithManagedDebug returns a ServeOpt that will start the server in debug
// mode, managing the reattach configuration handling and server lifecycle.
// Reattach configuration is output to stdout with human friendly instructions.
// By default, the server can be stopped with os.Interrupt (SIGINT; ctrl-c).
//
// Refer to the optional WithManagedDebugStopSignals and
// WithManagedDebugReattachConfigTimeout ServeOpt for additional configuration.
//
// The reattach configuration output of this handling is not protected by
// compatibility guarantees. Use the WithDebug ServeOpt for advanced use cases.
func WithManagedDebug() ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		if in.debugCh != nil {
			return errors.New("cannot set both WithDebug and WithManagedDebug")
		}

		in.managedDebug = true
		return nil
	})
}

// WithManagedDebugStopSignals returns a ServeOpt that will set the stop signals for a
// debug managed process (WithManagedDebug). When not configured, os.Interrupt
// (SIGINT; Ctrl-c) will stop the process.
func WithManagedDebugStopSignals(signals []os.Signal) ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		in.managedDebugStopSignals = signals
		return nil
	})
}

// WithManagedDebugReattachConfigTimeout returns a ServeOpt that will set the timeout
// for a debug managed process to start and return its reattach configuration.
// When not configured, 2 seconds is the default.
func WithManagedDebugReattachConfigTimeout(timeout time.Duration) ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		in.managedDebugReattachConfigTimeout = timeout
		return nil
	})
}

// WithGoPluginLogger returns a ServeOpt that will set the logger that
// go-plugin should use to log messages.
func WithGoPluginLogger(logger hclog.Logger) ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		in.logger = logger
		return nil
	})
}

// WithLoggingSink returns a ServeOpt that will enable the logging sink, which
// is used in test frameworks to control where terraform-plugin-log output is
// written and at what levels, mimicking Terraform's logging sink behaviors.
func WithLoggingSink(t testing.T) ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		in.useLoggingSink = t
		return nil
	})
}

// WithoutLogStderrOverride returns a ServeOpt that will disable the
// terraform-plugin-log behavior of logging to the stderr that existed at
// startup, not the stderr that exists when the logging statement is called.
func WithoutLogStderrOverride() ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		in.disableLogInitStderr = true
		return nil
	})
}

// WithoutLogLocation returns a ServeOpt that will exclude file names and line
// numbers from log output for the terraform-plugin-log logs generated by the
// SDKs and provider.
func WithoutLogLocation() ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		in.disableLogLocation = true
		return nil
	})
}

// WithLogEnvVarName sets the name of the provider for the purposes of the
// logging environment variable that controls the provider's log level. It is
// the part following TF_LOG_PROVIDER_ and defaults to the name part of the
// provider's registry address, or disabled if it can't parse the provider's
// registry address. Name must only contain letters, numbers, and hyphens.
func WithLogEnvVarName(name string) ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(name) {
			return errors.New("environment variable names can only contain a-z, A-Z, 0-9, and -")
		}
		in.envVar = name
		return nil
	})
}

// Serve starts a tfprotov6.ProviderServer serving, ready for Terraform to
// connect to it. The name passed in should be the fully qualified name that
// users will enter in the source field of the required_providers block, like
// "registry.terraform.io/hashicorp/time".
//
// Zero or more options to configure the server may also be passed. The default
// invocation is sufficient, but if the provider wants to run in debug mode or
// modify the logger that go-plugin is using, ServeOpts can be specified to
// support that.
func Serve(name string, serverFactory func() tfprotov6.ProviderServer, opts ...ServeOpt) error {
	// Defaults
	conf := ServeConfig{
		managedDebugReattachConfigTimeout: 2 * time.Second,
		managedDebugStopSignals:           []os.Signal{os.Interrupt},
	}

	for _, opt := range opts {
		err := opt.ApplyServeOpt(&conf)
		if err != nil {
			return err
		}
	}

	serveConfig := &plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  protocolVersionMajor,
			MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
			MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
		},
		Plugins: plugin.PluginSet{
			"provider": &GRPCProviderPlugin{
				GRPCProvider: serverFactory,
				Opts:         opts,
				Name:         name,
			},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			opts = append(opts, grpc.MaxRecvMsgSize(grpcMaxMessageSize))
			opts = append(opts, grpc.MaxSendMsgSize(grpcMaxMessageSize))

			return grpc.NewServer(opts...)
		},
	}

	if conf.logger != nil {
		serveConfig.Logger = conf.logger
	}

	if conf.managedDebug {
		ctx, cancel := context.WithCancel(context.Background())
		signalCh := make(chan os.Signal, len(conf.managedDebugStopSignals))

		signal.Notify(signalCh, conf.managedDebugStopSignals...)

		defer func() {
			signal.Stop(signalCh)
			cancel()
		}()

		go func() {
			select {
			case <-signalCh:
				cancel()
			case <-ctx.Done():
			}
		}()

		conf.debugCh = make(chan *plugin.ReattachConfig)
		conf.debugCloseCh = make(chan struct{})
		conf.debugCtx = ctx
	}

	if conf.debugCh != nil {
		serveConfig.Test = &plugin.ServeTestConfig{
			Context:          conf.debugCtx,
			ReattachConfigCh: conf.debugCh,
			CloseCh:          conf.debugCloseCh,
		}
	}

	if !conf.managedDebug {
		plugin.Serve(serveConfig)
		return nil
	}

	go plugin.Serve(serveConfig)

	var pluginReattachConfig *plugin.ReattachConfig

	select {
	case pluginReattachConfig = <-conf.debugCh:
	case <-time.After(conf.managedDebugReattachConfigTimeout):
		return errors.New("timeout waiting on reattach configuration")
	}

	if pluginReattachConfig == nil {
		return errors.New("nil reattach configuration received")
	}

	// Duplicate implementation is required because the go-plugin
	// ReattachConfig.Addr implementation is not friendly for JSON encoding
	// and to avoid importing terraform-exec.
	type reattachConfigAddr struct {
		Network string
		String  string
	}

	type reattachConfig struct {
		Protocol        string
		ProtocolVersion int
		Pid             int
		Test            bool
		Addr            reattachConfigAddr
	}

	reattachBytes, err := json.Marshal(map[string]reattachConfig{
		name: {
			Protocol:        string(pluginReattachConfig.Protocol),
			ProtocolVersion: pluginReattachConfig.ProtocolVersion,
			Pid:             pluginReattachConfig.Pid,
			Test:            pluginReattachConfig.Test,
			Addr: reattachConfigAddr{
				Network: pluginReattachConfig.Addr.Network(),
				String:  pluginReattachConfig.Addr.String(),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("Error building reattach string: %w", err)
	}

	reattachStr := string(reattachBytes)

	// This is currently intended to be executed via provider main function and
	// human friendly, so output directly to stdout.
	fmt.Printf("Provider started. To attach Terraform CLI, set the %s environment variable with the following:\n\n", envTfReattachProviders)

	switch runtime.GOOS {
	case "windows":
		fmt.Printf("\tCommand Prompt:\tset \"%s=%s\"\n", envTfReattachProviders, reattachStr)
		fmt.Printf("\tPowerShell:\t$env:%s='%s'\n", envTfReattachProviders, strings.ReplaceAll(reattachStr, `'`, `''`))
	default:
		fmt.Printf("\t%s='%s'\n", envTfReattachProviders, strings.ReplaceAll(reattachStr, `'`, `'"'"'`))
	}

	fmt.Println("")

	// Wait for the server to be done.
	<-conf.debugCloseCh

	return nil
}

type server struct {
	downstream tfprotov6.ProviderServer
	tfplugin6.UnimplementedProviderServer

	stopMu sync.Mutex
	stopCh chan struct{}

	tflogSDKOpts tfsdklog.Options
	tflogOpts    tflog.Options
	useTFLogSink bool
	testHandle   testing.T
	name         string

	// protocolDataDir is a directory to store raw protocol data files for
	// debugging purposes.
	protocolDataDir string

	// protocolVersion is the protocol version for the server.
	protocolVersion string
}

func mergeStop(ctx context.Context, cancel context.CancelFunc, stopCh chan struct{}) {
	select {
	case <-ctx.Done():
		return
	case <-stopCh:
		cancel()
	}
}

// stoppableContext returns a context that wraps `ctx` but will be canceled
// when the server's stopCh is closed.
//
// This is used to cancel all in-flight contexts when the Stop method of the
// server is called.
func (s *server) stoppableContext(ctx context.Context) context.Context {
	s.stopMu.Lock()
	defer s.stopMu.Unlock()

	stoppable, cancel := context.WithCancel(ctx)
	go mergeStop(stoppable, cancel, s.stopCh)
	return stoppable
}

// loggingContext returns a context that wraps `ctx` and has
// terraform-plugin-log loggers injected.
func (s *server) loggingContext(ctx context.Context) context.Context {
	if s.useTFLogSink {
		ctx = tfsdklog.RegisterTestSink(ctx, s.testHandle)
	}

	ctx = logging.InitContext(ctx, s.tflogSDKOpts, s.tflogOpts)
	ctx = logging.RequestIdContext(ctx)
	ctx = logging.ProviderAddressContext(ctx, s.name)
	ctx = logging.ProtocolVersionContext(ctx, s.protocolVersion)

	return ctx
}

// New converts a tfprotov6.ProviderServer into a server capable of handling
// Terraform protocol requests and issuing responses using the gRPC types.
func New(name string, serve tfprotov6.ProviderServer, opts ...ServeOpt) tfplugin6.ProviderServer {
	var conf ServeConfig
	for _, opt := range opts {
		err := opt.ApplyServeOpt(&conf)
		if err != nil {
			// this should never happen, we already executed all
			// this code as part of Serve
			panic(err)
		}
	}
	var sdkOptions tfsdklog.Options
	var options tflog.Options
	if !conf.disableLogInitStderr {
		sdkOptions = append(sdkOptions, tfsdklog.WithStderrFromInit())
		options = append(options, tfsdklog.WithStderrFromInit())
	}
	if conf.disableLogLocation {
		sdkOptions = append(sdkOptions, tfsdklog.WithoutLocation())
		options = append(options, tflog.WithoutLocation())
	}
	envVar := conf.envVar
	if envVar == "" {
		envVar = logging.ProviderLoggerName(name)
	}
	if envVar != "" {
		options = append(options, tfsdklog.WithLogName(envVar), tflog.WithLevelFromEnv(logging.EnvTfLogProvider, envVar))
	}
	return &server{
		downstream:      serve,
		stopCh:          make(chan struct{}),
		tflogOpts:       options,
		tflogSDKOpts:    sdkOptions,
		name:            name,
		useTFLogSink:    conf.useLoggingSink != nil,
		testHandle:      conf.useLoggingSink,
		protocolDataDir: os.Getenv(logging.EnvTfLogSdkProtoDataDir),
		protocolVersion: protocolVersion,
	}
}

func (s *server) GetProviderSchema(ctx context.Context, req *tfplugin6.GetProviderSchema_Request) (*tfplugin6.GetProviderSchema_Response, error) {
	rpc := "GetProviderSchema"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.GetProviderSchemaRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.GetProviderSchema(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	ret, err := toproto.GetProviderSchema_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) ConfigureProvider(ctx context.Context, req *tfplugin6.ConfigureProvider_Request) (*tfplugin6.ConfigureProvider_Response, error) {
	rpc := "ConfigureProvider"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.ConfigureProviderRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", r.Config)
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.ConfigureProvider(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	ret, err := toproto.Configure_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) ValidateProviderConfig(ctx context.Context, req *tfplugin6.ValidateProviderConfig_Request) (*tfplugin6.ValidateProviderConfig_Response, error) {
	rpc := "ValidateProviderConfig"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.ValidateProviderConfigRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", r.Config)
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.ValidateProviderConfig(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	ret, err := toproto.ValidateProviderConfig_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

// stop closes the stopCh associated with the server and replaces it with a new
// one.
//
// This causes all in-flight requests for the server to have their contexts
// canceled.
func (s *server) stop() {
	s.stopMu.Lock()
	defer s.stopMu.Unlock()

	close(s.stopCh)
	s.stopCh = make(chan struct{})
}

func (s *server) Stop(ctx context.Context, req *tfplugin6.StopProvider_Request) (*tfplugin6.StopProvider_Response, error) {
	rpc := "StopProvider"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.StopProviderRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.StopProvider(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, nil)
	logging.ProtocolTrace(ctx, "Closing all our contexts")
	s.stop()
	logging.ProtocolTrace(ctx, "Closed all our contexts")
	ret, err := toproto.Stop_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) ValidateDataResourceConfig(ctx context.Context, req *tfplugin6.ValidateDataResourceConfig_Request) (*tfplugin6.ValidateDataResourceConfig_Response, error) {
	rpc := "ValidateDataResourceConfig"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.DataSourceContext(ctx, req.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.ValidateDataResourceConfigRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", r.Config)
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.ValidateDataResourceConfig(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	ret, err := toproto.ValidateDataResourceConfig_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) ReadDataSource(ctx context.Context, req *tfplugin6.ReadDataSource_Request) (*tfplugin6.ReadDataSource_Response, error) {
	rpc := "ReadDataSource"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.DataSourceContext(ctx, req.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.ReadDataSourceRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", r.Config)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "ProviderMeta", r.ProviderMeta)
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.ReadDataSource(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "State", resp.State)
	ret, err := toproto.ReadDataSource_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) ValidateResourceConfig(ctx context.Context, req *tfplugin6.ValidateResourceConfig_Request) (*tfplugin6.ValidateResourceConfig_Response, error) {
	rpc := "ValidateResourceConfig"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, req.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.ValidateResourceConfigRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", r.Config)
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.ValidateResourceConfig(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	ret, err := toproto.ValidateResourceConfig_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) UpgradeResourceState(ctx context.Context, req *tfplugin6.UpgradeResourceState_Request) (*tfplugin6.UpgradeResourceState_Response, error) {
	rpc := "UpgradeResourceState"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, req.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.UpgradeResourceStateRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.UpgradeResourceState(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "UpgradedState", resp.UpgradedState)
	ret, err := toproto.UpgradeResourceState_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) ReadResource(ctx context.Context, req *tfplugin6.ReadResource_Request) (*tfplugin6.ReadResource_Response, error) {
	rpc := "ReadResource"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, req.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.ReadResourceRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "CurrentState", r.CurrentState)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "ProviderMeta", r.ProviderMeta)
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.ReadResource(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "NewState", resp.NewState)
	ret, err := toproto.ReadResource_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) PlanResourceChange(ctx context.Context, req *tfplugin6.PlanResourceChange_Request) (*tfplugin6.PlanResourceChange_Response, error) {
	rpc := "PlanResourceChange"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, req.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.PlanResourceChangeRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", r.Config)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "PriorState", r.PriorState)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "ProposedNewState", r.ProposedNewState)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "ProviderMeta", r.ProviderMeta)
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.PlanResourceChange(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "PlannedState", resp.PlannedState)
	ret, err := toproto.PlanResourceChange_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) ApplyResourceChange(ctx context.Context, req *tfplugin6.ApplyResourceChange_Request) (*tfplugin6.ApplyResourceChange_Response, error) {
	rpc := "ApplyResourceChange"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, req.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.ApplyResourceChangeRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", r.Config)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "PlannedState", r.PlannedState)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", r.Config)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", r.Config)
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.ApplyResourceChange(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "NewState", resp.NewState)
	ret, err := toproto.ApplyResourceChange_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}

func (s *server) ImportResourceState(ctx context.Context, req *tfplugin6.ImportResourceState_Request) (*tfplugin6.ImportResourceState_Response, error) {
	rpc := "ImportResourceState"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, req.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	r, err := fromproto.ImportResourceStateRequest(req)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting request from protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	ctx = tf6serverlogging.DownstreamRequest(ctx)
	resp, err := s.downstream.ImportResourceState(ctx, r)
	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	tf6serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	for _, importedResource := range resp.ImportedResources {
		logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response_ImportedResource", "State", importedResource.State)
	}
	ret, err := toproto.ImportResourceState_Response(resp)
	if err != nil {
		logging.ProtocolError(ctx, "Error converting response to protobuf", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}
	return ret, nil
}
