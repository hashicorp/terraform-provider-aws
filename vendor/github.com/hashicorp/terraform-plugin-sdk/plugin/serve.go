package plugin

import (
	"github.com/hashicorp/go-plugin"
	grpcplugin "github.com/hashicorp/terraform-plugin-sdk/internal/helper/plugin"
	proto "github.com/hashicorp/terraform-plugin-sdk/internal/tfplugin5"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	// The constants below are the names of the plugins that can be dispensed
	// from the plugin server.
	ProviderPluginName = "provider"

	// DefaultProtocolVersion is the protocol version assumed for legacy clients that don't specify
	// a particular version during their handshake. This is the version used when Terraform 0.10
	// and 0.11 launch plugins that were built with support for both versions 4 and 5, and must
	// stay unchanged at 4 until we intentionally build plugins that are not compatible with 0.10 and
	// 0.11.
	DefaultProtocolVersion = 4
)

// Handshake is the HandshakeConfig used to configure clients and servers.
var Handshake = plugin.HandshakeConfig{
	// The ProtocolVersion is the version that must match between TF core
	// and TF plugins. This should be bumped whenever a change happens in
	// one or the other that makes it so that they can't safely communicate.
	// This could be adding a new interface value, it could be how
	// helper/schema computes diffs, etc.
	ProtocolVersion: DefaultProtocolVersion,

	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

type ProviderFunc func() terraform.ResourceProvider
type GRPCProviderFunc func() proto.ProviderServer

// ServeOpts are the configurations to serve a plugin.
type ServeOpts struct {
	ProviderFunc ProviderFunc

	// Wrapped versions of the above plugins will automatically shimmed and
	// added to the GRPC functions when possible.
	GRPCProviderFunc GRPCProviderFunc
}

// Serve serves a plugin. This function never returns and should be the final
// function called in the main function of the plugin.
func Serve(opts *ServeOpts) {
	// since the plugins may not yet be aware of the new protocol, we
	// automatically wrap the plugins in the grpc shims.
	if opts.GRPCProviderFunc == nil && opts.ProviderFunc != nil {
		provider := grpcplugin.NewGRPCProviderServerShim(opts.ProviderFunc())
		// this is almost always going to be a *schema.Provider, but check that
		// we got back a valid provider just in case.
		if provider != nil {
			opts.GRPCProviderFunc = func() proto.ProviderServer {
				return provider
			}
		}
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig:  Handshake,
		VersionedPlugins: pluginSet(opts),
		GRPCServer:       plugin.DefaultGRPCServer,
	})
}

// pluginMap returns the legacy map[string]plugin.Plugin to use for configuring
// a plugin server or client.
func legacyPluginMap(opts *ServeOpts) map[string]plugin.Plugin {
	return map[string]plugin.Plugin{
		"provider": &ResourceProviderPlugin{
			ResourceProvider: opts.ProviderFunc,
		},
	}
}

func pluginSet(opts *ServeOpts) map[int]plugin.PluginSet {
	// Set the legacy netrpc plugins at version 4.
	// The oldest version is returned in when executed by a legacy go-plugin
	// client.
	plugins := map[int]plugin.PluginSet{
		4: legacyPluginMap(opts),
	}

	// add the new protocol versions if they're configured
	if opts.GRPCProviderFunc != nil {
		plugins[5] = plugin.PluginSet{}
		if opts.GRPCProviderFunc != nil {
			plugins[5]["provider"] = &GRPCProviderPlugin{
				GRPCProvider: opts.GRPCProviderFunc,
			}
		}
	}
	return plugins
}
