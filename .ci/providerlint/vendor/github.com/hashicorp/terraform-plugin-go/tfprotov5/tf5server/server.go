// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tf5server

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

	"google.golang.org/grpc"

	"github.com/hashicorp/terraform-plugin-go/internal/logging"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/fromproto"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tf5serverlogging"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/toproto"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
	"github.com/mitchellh/go-testing-interface"
)

const (
	// protocolVersionMajor represents the major version number of the protocol
	// being served. This is used during the plugin handshake to validate the
	// server and client are compatible.
	//
	// In the future, it may be possible to include this information directly
	// in the protocol buffers rather than recreating a constant here.
	protocolVersionMajor uint = 5

	// protocolVersionMinor represents the minor version number of the protocol
	// being served. Backwards compatible additions are possible in the
	// protocol definitions, which is when this may be increased. While it is
	// not used in plugin negotiation, it can be helpful to include this value
	// for debugging, such as in logs.
	//
	// In the future, it may be possible to include this information directly
	// in the protocol buffers rather than recreating a constant here.
	protocolVersionMinor uint = 4
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

// Serve starts a tfprotov5.ProviderServer serving, ready for Terraform to
// connect to it. The name passed in should be the fully qualified name that
// users will enter in the source field of the required_providers block, like
// "registry.terraform.io/hashicorp/time".
//
// Zero or more options to configure the server may also be passed. The default
// invocation is sufficient, but if the provider wants to run in debug mode or
// modify the logger that go-plugin is using, ServeOpts can be specified to
// support that.
func Serve(name string, serverFactory func() tfprotov5.ProviderServer, opts ...ServeOpt) error {
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
				Name:         name,
				Opts:         opts,
				GRPCProvider: serverFactory,
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
	downstream tfprotov5.ProviderServer
	tfplugin5.UnimplementedProviderServer

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

// New converts a tfprotov5.ProviderServer into a server capable of handling
// Terraform protocol requests and issuing responses using the gRPC types.
func New(name string, serve tfprotov5.ProviderServer, opts ...ServeOpt) tfplugin5.ProviderServer {
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

func (s *server) GetMetadata(ctx context.Context, protoReq *tfplugin5.GetMetadata_Request) (*tfplugin5.GetMetadata_Response, error) {
	rpc := "GetMetadata"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.GetMetadataRequest(protoReq)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.GetMetadata(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	tf5serverlogging.ServerCapabilities(ctx, resp.ServerCapabilities)

	protoResp := toproto.GetMetadata_Response(resp)

	return protoResp, nil
}

func (s *server) GetSchema(ctx context.Context, protoReq *tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
	rpc := "GetProviderSchema"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.GetProviderSchemaRequest(protoReq)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.GetProviderSchema(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	tf5serverlogging.ServerCapabilities(ctx, resp.ServerCapabilities)

	protoResp := toproto.GetProviderSchema_Response(resp)

	return protoResp, nil
}

func (s *server) PrepareProviderConfig(ctx context.Context, protoReq *tfplugin5.PrepareProviderConfig_Request) (*tfplugin5.PrepareProviderConfig_Response, error) {
	rpc := "PrepareProviderConfig"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.PrepareProviderConfigRequest(protoReq)

	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", req.Config)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.PrepareProviderConfig(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "PreparedConfig", resp.PreparedConfig)

	protoResp := toproto.PrepareProviderConfig_Response(resp)

	return protoResp, nil
}

func (s *server) Configure(ctx context.Context, protoReq *tfplugin5.Configure_Request) (*tfplugin5.Configure_Response, error) {
	rpc := "Configure"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")
	req := fromproto.ConfigureProviderRequest(protoReq)

	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", req.Config)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.ConfigureProvider(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)

	protoResp := toproto.Configure_Response(resp)

	return protoResp, nil
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

func (s *server) Stop(ctx context.Context, protoReq *tfplugin5.Stop_Request) (*tfplugin5.Stop_Response, error) {
	rpc := "Stop"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.StopProviderRequest(protoReq)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.StopProvider(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, nil)
	logging.ProtocolTrace(ctx, "Closing all our contexts")
	s.stop()
	logging.ProtocolTrace(ctx, "Closed all our contexts")

	protoResp := toproto.Stop_Response(resp)

	return protoResp, nil
}

func (s *server) ValidateDataSourceConfig(ctx context.Context, protoReq *tfplugin5.ValidateDataSourceConfig_Request) (*tfplugin5.ValidateDataSourceConfig_Response, error) {
	rpc := "ValidateDataSourceConfig"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.DataSourceContext(ctx, protoReq.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.ValidateDataSourceConfigRequest(protoReq)

	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", req.Config)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.ValidateDataSourceConfig(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)

	protoResp := toproto.ValidateDataSourceConfig_Response(resp)

	return protoResp, nil
}

func (s *server) ReadDataSource(ctx context.Context, protoReq *tfplugin5.ReadDataSource_Request) (*tfplugin5.ReadDataSource_Response, error) {
	rpc := "ReadDataSource"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.DataSourceContext(ctx, protoReq.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.ReadDataSourceRequest(protoReq)

	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", req.Config)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "ProviderMeta", req.ProviderMeta)
	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.ReadDataSource(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "State", resp.State)

	protoResp := toproto.ReadDataSource_Response(resp)

	return protoResp, nil
}

func (s *server) ValidateResourceTypeConfig(ctx context.Context, protoReq *tfplugin5.ValidateResourceTypeConfig_Request) (*tfplugin5.ValidateResourceTypeConfig_Response, error) {
	rpc := "ValidateResourceTypeConfig"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, protoReq.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.ValidateResourceTypeConfigRequest(protoReq)

	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", req.Config)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.ValidateResourceTypeConfig(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)

	protoResp := toproto.ValidateResourceTypeConfig_Response(resp)

	return protoResp, nil
}

func (s *server) UpgradeResourceState(ctx context.Context, protoReq *tfplugin5.UpgradeResourceState_Request) (*tfplugin5.UpgradeResourceState_Response, error) {
	rpc := "UpgradeResourceState"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, protoReq.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.UpgradeResourceStateRequest(protoReq)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.UpgradeResourceState(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "UpgradedState", resp.UpgradedState)

	protoResp := toproto.UpgradeResourceState_Response(resp)

	return protoResp, nil
}

func (s *server) ReadResource(ctx context.Context, protoReq *tfplugin5.ReadResource_Request) (*tfplugin5.ReadResource_Response, error) {
	rpc := "ReadResource"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, protoReq.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.ReadResourceRequest(protoReq)

	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "CurrentState", req.CurrentState)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "ProviderMeta", req.ProviderMeta)
	logging.ProtocolPrivateData(ctx, s.protocolDataDir, rpc, "Request", "Private", req.Private)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.ReadResource(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)

	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "NewState", resp.NewState)
	logging.ProtocolPrivateData(ctx, s.protocolDataDir, rpc, "Response", "Private", resp.Private)

	protoResp := toproto.ReadResource_Response(resp)

	return protoResp, nil
}

func (s *server) PlanResourceChange(ctx context.Context, protoReq *tfplugin5.PlanResourceChange_Request) (*tfplugin5.PlanResourceChange_Response, error) {
	rpc := "PlanResourceChange"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, protoReq.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.PlanResourceChangeRequest(protoReq)

	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", req.Config)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "PriorState", req.PriorState)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "ProposedNewState", req.ProposedNewState)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "ProviderMeta", req.ProviderMeta)
	logging.ProtocolPrivateData(ctx, s.protocolDataDir, rpc, "Request", "PriorPrivate", req.PriorPrivate)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.PlanResourceChange(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "PlannedState", resp.PlannedState)
	logging.ProtocolPrivateData(ctx, s.protocolDataDir, rpc, "Response", "PlannedPrivate", resp.PlannedPrivate)

	protoResp := toproto.PlanResourceChange_Response(resp)

	return protoResp, nil
}

func (s *server) ApplyResourceChange(ctx context.Context, protoReq *tfplugin5.ApplyResourceChange_Request) (*tfplugin5.ApplyResourceChange_Response, error) {
	rpc := "ApplyResourceChange"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, protoReq.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.ApplyResourceChangeRequest(protoReq)

	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "Config", req.Config)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "PlannedState", req.PlannedState)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "PriorState", req.PriorState)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", "ProviderMeta", req.ProviderMeta)
	logging.ProtocolPrivateData(ctx, s.protocolDataDir, rpc, "Request", "PlannedPrivate", req.PlannedPrivate)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.ApplyResourceChange(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "NewState", resp.NewState)
	logging.ProtocolPrivateData(ctx, s.protocolDataDir, rpc, "Response", "Private", resp.Private)

	protoResp := toproto.ApplyResourceChange_Response(resp)

	return protoResp, nil
}

func (s *server) ImportResourceState(ctx context.Context, protoReq *tfplugin5.ImportResourceState_Request) (*tfplugin5.ImportResourceState_Response, error) {
	rpc := "ImportResourceState"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, protoReq.TypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	req := fromproto.ImportResourceStateRequest(protoReq)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	resp, err := s.downstream.ImportResourceState(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)

	for _, importedResource := range resp.ImportedResources {
		logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response_ImportedResource", "State", importedResource.State)
		logging.ProtocolPrivateData(ctx, s.protocolDataDir, rpc, "Response_ImportedResource", "Private", importedResource.Private)
	}

	protoResp := toproto.ImportResourceState_Response(resp)

	return protoResp, nil
}

func (s *server) MoveResourceState(ctx context.Context, protoReq *tfplugin5.MoveResourceState_Request) (*tfplugin5.MoveResourceState_Response, error) {
	rpc := "MoveResourceState"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = logging.ResourceContext(ctx, protoReq.TargetTypeName)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	// Remove this check and error in preference of
	// s.downstream.MoveResourceState below once ResourceServer interface
	// implements the MoveResourceState method.
	// Reference: https://github.com/hashicorp/terraform-plugin-go/issues/363
	// nolint:staticcheck
	resourceServerWMRS, ok := s.downstream.(tfprotov5.ResourceServerWithMoveResourceState)

	if !ok {
		logging.ProtocolError(ctx, "ProviderServer does not implement ResourceServerWithMoveResourceState")

		protoResp := &tfplugin5.MoveResourceState_Response{
			Diagnostics: []*tfplugin5.Diagnostic{
				{
					Severity: tfplugin5.Diagnostic_ERROR,
					Summary:  "Provider Move Resource State Not Implemented",
					Detail: "A MoveResourceState call was received by the provider, however the provider does not implement the call. " +
						"Either upgrade the provider to a version that implements move resource state support or this is a bug in Terraform that should be reported to the Terraform maintainers.",
				},
			},
		}

		return protoResp, nil
	}

	req := fromproto.MoveResourceStateRequest(protoReq)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	// Reference: https://github.com/hashicorp/terraform-plugin-go/issues/363
	// resp, err := s.downstream.MoveResourceState(ctx, req)
	resp, err := resourceServerWMRS.MoveResourceState(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]interface{}{logging.KeyError: err})

		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "TargetState", resp.TargetState)

	protoResp := toproto.MoveResourceState_Response(resp)

	return protoResp, nil
}

func (s *server) CallFunction(ctx context.Context, protoReq *tfplugin5.CallFunction_Request) (*tfplugin5.CallFunction_Response, error) {
	rpc := "CallFunction"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	// Remove this check and error in preference of s.downstream.CallFunction
	// below once ProviderServer interface requires FunctionServer.
	// Reference: https://github.com/hashicorp/terraform-plugin-go/issues/353
	functionServer, ok := s.downstream.(tfprotov5.FunctionServer)

	if !ok {
		logging.ProtocolError(ctx, "ProviderServer does not implement FunctionServer")

		text := "Provider Functions Not Implemented: A provider-defined function call was received by the provider, however the provider does not implement functions. " +
			"Either upgrade the provider to a version that implements provider-defined functions or this is a bug in Terraform that should be reported to the Terraform maintainers."

		protoResp := &tfplugin5.CallFunction_Response{
			Error: &tfplugin5.FunctionError{
				Text: text,
			},
		}

		return protoResp, nil
	}

	req := fromproto.CallFunctionRequest(protoReq)

	for position, argument := range req.Arguments {
		logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Request", fmt.Sprintf("Arguments_%d", position), argument)
	}

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	// Reference: https://github.com/hashicorp/terraform-plugin-go/issues/353
	// resp, err := s.downstream.CallFunction(ctx, req)
	resp, err := functionServer.CallFunction(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]any{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponseWithError(ctx, resp.Error)
	logging.ProtocolData(ctx, s.protocolDataDir, rpc, "Response", "Result", resp.Result)

	protoResp := toproto.CallFunction_Response(resp)

	return protoResp, nil
}

func (s *server) GetFunctions(ctx context.Context, protoReq *tfplugin5.GetFunctions_Request) (*tfplugin5.GetFunctions_Response, error) {
	rpc := "GetFunctions"
	ctx = s.loggingContext(ctx)
	ctx = logging.RpcContext(ctx, rpc)
	ctx = s.stoppableContext(ctx)
	logging.ProtocolTrace(ctx, "Received request")
	defer logging.ProtocolTrace(ctx, "Served request")

	// Remove this check and response in preference of s.downstream.GetFunctions
	// below once ProviderServer interface requires FunctionServer.
	// Reference: https://github.com/hashicorp/terraform-plugin-go/issues/353
	functionServer, ok := s.downstream.(tfprotov5.FunctionServer)

	if !ok {
		logging.ProtocolWarn(ctx, "ProviderServer does not implement FunctionServer")

		protoResp := &tfplugin5.GetFunctions_Response{
			Functions: map[string]*tfplugin5.Function{},
		}

		return protoResp, nil
	}

	req := fromproto.GetFunctionsRequest(protoReq)

	ctx = tf5serverlogging.DownstreamRequest(ctx)

	// Reference: https://github.com/hashicorp/terraform-plugin-go/issues/353
	// resp, err := s.downstream.GetFunctions(ctx, req)
	resp, err := functionServer.GetFunctions(ctx, req)

	if err != nil {
		logging.ProtocolError(ctx, "Error from downstream", map[string]any{logging.KeyError: err})
		return nil, err
	}

	tf5serverlogging.DownstreamResponse(ctx, resp.Diagnostics)

	protoResp := toproto.GetFunctions_Response(resp)

	return protoResp, nil
}
