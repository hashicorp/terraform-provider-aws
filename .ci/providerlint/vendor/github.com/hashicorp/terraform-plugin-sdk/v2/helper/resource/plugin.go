package resource

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugintest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	testing "github.com/mitchellh/go-testing-interface"
)

// protov5ProviderFactory is a function which is called to start a protocol
// version 5 provider server.
type protov5ProviderFactory func() (tfprotov5.ProviderServer, error)

// protov5ProviderFactories is a mapping of provider addresses to provider
// factory for protocol version 5 provider servers.
type protov5ProviderFactories map[string]func() (tfprotov5.ProviderServer, error)

// merge combines provider factories.
//
// In case of an overlapping entry, the later entry will overwrite the previous
// value.
func (pf protov5ProviderFactories) merge(otherPfs ...protov5ProviderFactories) protov5ProviderFactories {
	result := make(protov5ProviderFactories)

	for name, providerFactory := range pf {
		result[name] = providerFactory
	}

	for _, otherPf := range otherPfs {
		for name, providerFactory := range otherPf {
			result[name] = providerFactory
		}
	}

	return result
}

// protov6ProviderFactory is a function which is called to start a protocol
// version 6 provider server.
type protov6ProviderFactory func() (tfprotov6.ProviderServer, error)

// protov6ProviderFactories is a mapping of provider addresses to provider
// factory for protocol version 6 provider servers.
type protov6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

// merge combines provider factories.
//
// In case of an overlapping entry, the later entry will overwrite the previous
// value.
func (pf protov6ProviderFactories) merge(otherPfs ...protov6ProviderFactories) protov6ProviderFactories {
	result := make(protov6ProviderFactories)

	for name, providerFactory := range pf {
		result[name] = providerFactory
	}

	for _, otherPf := range otherPfs {
		for name, providerFactory := range otherPf {
			result[name] = providerFactory
		}
	}

	return result
}

// sdkProviderFactory is a function which is called to start a SDK provider
// server.
type sdkProviderFactory func() (*schema.Provider, error)

// protov6ProviderFactories is a mapping of provider addresses to provider
// factory for protocol version 6 provider servers.
type sdkProviderFactories map[string]func() (*schema.Provider, error)

// merge combines provider factories.
//
// In case of an overlapping entry, the later entry will overwrite the previous
// value.
func (pf sdkProviderFactories) merge(otherPfs ...sdkProviderFactories) sdkProviderFactories {
	result := make(sdkProviderFactories)

	for name, providerFactory := range pf {
		result[name] = providerFactory
	}

	for _, otherPf := range otherPfs {
		for name, providerFactory := range otherPf {
			result[name] = providerFactory
		}
	}

	return result
}

type providerFactories struct {
	legacy  sdkProviderFactories
	protov5 protov5ProviderFactories
	protov6 protov6ProviderFactories
}

func runProviderCommand(ctx context.Context, t testing.T, f func() error, wd *plugintest.WorkingDir, factories *providerFactories) error {
	// don't point to this as a test failure location
	// point to whatever called it
	t.Helper()

	// This should not happen, but prevent panics just in case.
	if factories == nil {
		err := fmt.Errorf("Provider factories are missing to run Terraform command. Please report this bug in the testing framework.")
		logging.HelperResourceError(ctx, err.Error())
		return err
	}

	// Run the providers in the same process as the test runner using the
	// reattach behavior in Terraform. This ensures we get test coverage
	// and enables the use of delve as a debugger.

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// this is needed so Terraform doesn't default to expecting protocol 4;
	// we're skipping the handshake because Terraform didn't launch the
	// plugins.
	os.Setenv("PLUGIN_PROTOCOL_VERSIONS", "5")

	// Acceptance testing does not need to call checkpoint as the output
	// is not accessible, nor desirable if explicitly using
	// TF_ACC_TERRAFORM_PATH or TF_ACC_TERRAFORM_VERSION environment variables.
	//
	// Avoid calling (tfexec.Terraform).SetEnv() as it will stop copying
	// os.Environ() and prevents TF_VAR_ environment variable usage.
	os.Setenv("CHECKPOINT_DISABLE", "1")

	// Terraform 0.12.X and 0.13.X+ treat namespaceless providers
	// differently in terms of what namespace they default to. So we're
	// going to set both variations, as we don't know which version of
	// Terraform we're talking to. We're also going to allow overriding
	// the host or namespace using environment variables.
	var namespaces []string
	host := "registry.terraform.io"
	if v := os.Getenv(EnvTfAccProviderNamespace); v != "" {
		namespaces = append(namespaces, v)
	} else {
		namespaces = append(namespaces, "-", "hashicorp")
	}
	if v := os.Getenv(EnvTfAccProviderHost); v != "" {
		host = v
	}

	// schema.Provider have a global stop context that is created outside
	// the server context and have their own associated goroutine. Since
	// Terraform does not call the StopProvider RPC to stop the server in
	// reattach mode, ensure that we save these servers to later call that
	// RPC and end those goroutines.
	legacyProviderServers := make([]*schema.GRPCProviderServer, 0, len(factories.legacy))

	// Spin up gRPC servers for every provider factory, start a
	// WaitGroup to listen for all of the close channels.
	var wg sync.WaitGroup
	reattachInfo := map[string]tfexec.ReattachConfig{}
	for providerName, factory := range factories.legacy {
		// providerName may be returned as terraform-provider-foo, and
		// we need just foo. So let's fix that.
		providerName = strings.TrimPrefix(providerName, "terraform-provider-")
		providerAddress := getProviderAddr(providerName)

		logging.HelperResourceDebug(ctx, "Creating sdkv2 provider instance", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		provider, err := factory()
		if err != nil {
			return fmt.Errorf("unable to create provider %q from factory: %w", providerName, err)
		}

		logging.HelperResourceDebug(ctx, "Created sdkv2 provider instance", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		// keep track of the running factory, so we can make sure it's
		// shut down.
		wg.Add(1)

		grpcProviderServer := schema.NewGRPCProviderServer(provider)
		legacyProviderServers = append(legacyProviderServers, grpcProviderServer)

		// Ensure StopProvider is always called when returning early.
		defer grpcProviderServer.StopProvider(ctx, nil) //nolint:errcheck // does not return errors

		// configure the settings our plugin will be served with
		// the GRPCProviderFunc wraps a non-gRPC provider server
		// into a gRPC interface, and the logger just discards logs
		// from go-plugin.
		opts := &plugin.ServeOpts{
			GRPCProviderFunc: func() tfprotov5.ProviderServer {
				return grpcProviderServer
			},
			Logger: hclog.New(&hclog.LoggerOptions{
				Name:   "plugintest",
				Level:  hclog.Trace,
				Output: io.Discard,
			}),
			NoLogOutputOverride: true,
			UseTFLogSink:        t,
			ProviderAddr:        providerAddress,
		}

		logging.HelperResourceDebug(ctx, "Starting sdkv2 provider instance server", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		config, closeCh, err := plugin.DebugServe(ctx, opts)
		if err != nil {
			return fmt.Errorf("unable to serve provider %q: %w", providerName, err)
		}

		logging.HelperResourceDebug(ctx, "Started sdkv2 provider instance server", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		tfexecConfig := tfexec.ReattachConfig{
			Protocol:        config.Protocol,
			ProtocolVersion: config.ProtocolVersion,
			Pid:             config.Pid,
			Test:            config.Test,
			Addr: tfexec.ReattachConfigAddr{
				Network: config.Addr.Network,
				String:  config.Addr.String,
			},
		}

		// when the provider exits, remove one from the waitgroup
		// so we can track when everything is done
		go func(c <-chan struct{}) {
			<-c
			wg.Done()
		}(closeCh)

		// set our provider's reattachinfo in our map, once
		// for every namespace that different Terraform versions
		// may expect.
		for _, ns := range namespaces {
			reattachInfo[strings.TrimSuffix(host, "/")+"/"+
				strings.TrimSuffix(ns, "/")+"/"+
				providerName] = tfexecConfig
		}
	}

	// Now spin up gRPC servers for every protov5 provider factory
	// in the same way.
	for providerName, factory := range factories.protov5 {
		// providerName may be returned as terraform-provider-foo, and
		// we need just foo. So let's fix that.
		providerName = strings.TrimPrefix(providerName, "terraform-provider-")
		providerAddress := getProviderAddr(providerName)

		// If the user has supplied the same provider in both
		// ProviderFactories and ProtoV5ProviderFactories, they made a
		// mistake and we should exit early.
		for _, ns := range namespaces {
			reattachString := strings.TrimSuffix(host, "/") + "/" +
				strings.TrimSuffix(ns, "/") + "/" +
				providerName
			if _, ok := reattachInfo[reattachString]; ok {
				return fmt.Errorf("Provider %s registered in both TestCase.ProviderFactories and TestCase.ProtoV5ProviderFactories: please use one or the other, or supply a muxed provider to TestCase.ProtoV5ProviderFactories.", providerName)
			}
		}

		logging.HelperResourceDebug(ctx, "Creating tfprotov5 provider instance", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		provider, err := factory()
		if err != nil {
			return fmt.Errorf("unable to create provider %q from factory: %w", providerName, err)
		}

		logging.HelperResourceDebug(ctx, "Created tfprotov5 provider instance", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		// keep track of the running factory, so we can make sure it's
		// shut down.
		wg.Add(1)

		// configure the settings our plugin will be served with
		// the GRPCProviderFunc wraps a non-gRPC provider server
		// into a gRPC interface, and the logger just discards logs
		// from go-plugin.
		opts := &plugin.ServeOpts{
			GRPCProviderFunc: func() tfprotov5.ProviderServer {
				return provider
			},
			Logger: hclog.New(&hclog.LoggerOptions{
				Name:   "plugintest",
				Level:  hclog.Trace,
				Output: io.Discard,
			}),
			NoLogOutputOverride: true,
			UseTFLogSink:        t,
			ProviderAddr:        providerAddress,
		}

		logging.HelperResourceDebug(ctx, "Starting tfprotov5 provider instance server", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		config, closeCh, err := plugin.DebugServe(ctx, opts)
		if err != nil {
			return fmt.Errorf("unable to serve provider %q: %w", providerName, err)
		}

		logging.HelperResourceDebug(ctx, "Started tfprotov5 provider instance server", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		tfexecConfig := tfexec.ReattachConfig{
			Protocol:        config.Protocol,
			ProtocolVersion: config.ProtocolVersion,
			Pid:             config.Pid,
			Test:            config.Test,
			Addr: tfexec.ReattachConfigAddr{
				Network: config.Addr.Network,
				String:  config.Addr.String,
			},
		}

		// when the provider exits, remove one from the waitgroup
		// so we can track when everything is done
		go func(c <-chan struct{}) {
			<-c
			wg.Done()
		}(closeCh)

		// set our provider's reattachinfo in our map, once
		// for every namespace that different Terraform versions
		// may expect.
		for _, ns := range namespaces {
			reattachString := strings.TrimSuffix(host, "/") + "/" +
				strings.TrimSuffix(ns, "/") + "/" +
				providerName
			reattachInfo[reattachString] = tfexecConfig
		}
	}

	// Now spin up gRPC servers for every protov6 provider factory
	// in the same way.
	for providerName, factory := range factories.protov6 {
		// providerName may be returned as terraform-provider-foo, and
		// we need just foo. So let's fix that.
		providerName = strings.TrimPrefix(providerName, "terraform-provider-")
		providerAddress := getProviderAddr(providerName)

		// If the user has already registered this provider in
		// ProviderFactories or ProtoV5ProviderFactories, they made a
		// mistake and we should exit early.
		for _, ns := range namespaces {
			reattachString := strings.TrimSuffix(host, "/") + "/" +
				strings.TrimSuffix(ns, "/") + "/" +
				providerName
			if _, ok := reattachInfo[reattachString]; ok {
				return fmt.Errorf("Provider %s registered in both TestCase.ProtoV6ProviderFactories and either TestCase.ProviderFactories or TestCase.ProtoV5ProviderFactories: please use one of the three, or supply a muxed provider to TestCase.ProtoV5ProviderFactories.", providerName)
			}
		}

		logging.HelperResourceDebug(ctx, "Creating tfprotov6 provider instance", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		provider, err := factory()
		if err != nil {
			return fmt.Errorf("unable to create provider %q from factory: %w", providerName, err)
		}

		logging.HelperResourceDebug(ctx, "Created tfprotov6 provider instance", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		// keep track of the running factory, so we can make sure it's
		// shut down.
		wg.Add(1)

		opts := &plugin.ServeOpts{
			GRPCProviderV6Func: func() tfprotov6.ProviderServer {
				return provider
			},
			Logger: hclog.New(&hclog.LoggerOptions{
				Name:   "plugintest",
				Level:  hclog.Trace,
				Output: io.Discard,
			}),
			NoLogOutputOverride: true,
			UseTFLogSink:        t,
			ProviderAddr:        providerAddress,
		}

		logging.HelperResourceDebug(ctx, "Starting tfprotov6 provider instance server", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		config, closeCh, err := plugin.DebugServe(ctx, opts)
		if err != nil {
			return fmt.Errorf("unable to serve provider %q: %w", providerName, err)
		}

		logging.HelperResourceDebug(ctx, "Started tfprotov6 provider instance server", map[string]interface{}{logging.KeyProviderAddress: providerAddress})

		tfexecConfig := tfexec.ReattachConfig{
			Protocol:        config.Protocol,
			ProtocolVersion: config.ProtocolVersion,
			Pid:             config.Pid,
			Test:            config.Test,
			Addr: tfexec.ReattachConfigAddr{
				Network: config.Addr.Network,
				String:  config.Addr.String,
			},
		}

		// when the provider exits, remove one from the waitgroup
		// so we can track when everything is done
		go func(c <-chan struct{}) {
			<-c
			wg.Done()
		}(closeCh)

		// set our provider's reattachinfo in our map, once
		// for every namespace that different Terraform versions
		// may expect.
		for _, ns := range namespaces {
			reattachString := strings.TrimSuffix(host, "/") + "/" +
				strings.TrimSuffix(ns, "/") + "/" +
				providerName
			reattachInfo[reattachString] = tfexecConfig
		}
	}

	// set the working directory reattach info that will tell Terraform how to
	// connect to our various running servers.
	wd.SetReattachInfo(ctx, reattachInfo)

	logging.HelperResourceTrace(ctx, "Calling wrapped Terraform CLI command")

	// ok, let's call whatever Terraform command the test was trying to
	// call, now that we know it'll attach back to those servers we just
	// started.
	err := f()
	if err != nil {
		logging.HelperResourceWarn(ctx, "Error running Terraform CLI command", map[string]interface{}{logging.KeyError: err})
	}

	logging.HelperResourceTrace(ctx, "Called wrapped Terraform CLI command")
	logging.HelperResourceDebug(ctx, "Stopping providers")

	// cancel the servers so they'll return. Otherwise, this closeCh won't
	// get closed, and we'll hang here.
	cancel()

	// For legacy providers, call the StopProvider RPC so the StopContext
	// goroutine is cleaned up properly.
	for _, legacyProviderServer := range legacyProviderServers {
		legacyProviderServer.StopProvider(ctx, nil) //nolint:errcheck // does not return errors
	}

	logging.HelperResourceTrace(ctx, "Waiting for providers to stop")

	// wait for the servers to actually shut down; it may take a moment for
	// them to clean up, or whatever.
	// TODO: add a timeout here?
	// PC: do we need one? The test will time out automatically...
	wg.Wait()

	logging.HelperResourceTrace(ctx, "Providers have successfully stopped")

	// once we've run the Terraform command, let's remove the reattach
	// information from the WorkingDir's environment. The WorkingDir will
	// persist until the next call, but the server in the reattach info
	// doesn't exist anymore at this point, so the reattach info is no
	// longer valid. In theory it should be overwritten in the next call,
	// but just to avoid any confusing bug reports, let's just unset the
	// environment variable altogether.
	wd.UnsetReattachInfo()

	// return any error returned from the orchestration code running
	// Terraform commands
	return err
}

func getProviderAddr(name string) string {
	host := "registry.terraform.io"
	namespace := "hashicorp"
	if v := os.Getenv(EnvTfAccProviderNamespace); v != "" {
		namespace = v
	}
	if v := os.Getenv(EnvTfAccProviderHost); v != "" {
		host = v
	}
	return strings.TrimSuffix(host, "/") + "/" +
		strings.TrimSuffix(namespace, "/") + "/" +
		name
}
