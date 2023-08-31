// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugin

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/go-plugin"
)

// ReattachConfig holds the information Terraform needs to be able to attach
// itself to a provider process, so it can drive the process.
type ReattachConfig struct {
	Protocol        string
	ProtocolVersion int
	Pid             int
	Test            bool
	Addr            ReattachConfigAddr
}

// ReattachConfigAddr is a JSON-encoding friendly version of net.Addr.
type ReattachConfigAddr struct {
	Network string
	String  string
}

// DebugServe starts a plugin server in debug mode; this should only be used
// when the provider will manage its own lifecycle. It is not recommended for
// normal usage; Serve is the correct function for that.
func DebugServe(ctx context.Context, opts *ServeOpts) (ReattachConfig, <-chan struct{}, error) {
	reattachCh := make(chan *plugin.ReattachConfig)
	closeCh := make(chan struct{})

	if opts == nil {
		return ReattachConfig{}, closeCh, errors.New("ServeOpts must be passed in with at least GRPCProviderFunc, GRPCProviderV6Func, or ProviderFunc")
	}

	opts.TestConfig = &plugin.ServeTestConfig{
		Context:          ctx,
		ReattachConfigCh: reattachCh,
		CloseCh:          closeCh,
	}

	go Serve(opts)

	var config *plugin.ReattachConfig
	select {
	case config = <-reattachCh:
	case <-time.After(2 * time.Second):
		return ReattachConfig{}, closeCh, errors.New("timeout waiting on reattach config")
	}

	if config == nil {
		return ReattachConfig{}, closeCh, errors.New("nil reattach config received")
	}

	return ReattachConfig{
		Protocol:        string(config.Protocol),
		ProtocolVersion: config.ProtocolVersion,
		Pid:             config.Pid,
		Test:            config.Test,
		Addr: ReattachConfigAddr{
			Network: config.Addr.Network(),
			String:  config.Addr.String(),
		},
	}, closeCh, nil
}

// Debug starts a debug server and controls its lifecycle, printing the
// information needed for Terraform to connect to the provider to stdout.
// os.Interrupt will be captured and used to stop the server.
//
// Deprecated: Use the Serve function with the ServeOpts Debug field instead.
func Debug(ctx context.Context, providerAddr string, opts *ServeOpts) error {
	if opts == nil {
		return errors.New("ServeOpts must be passed in with at least GRPCProviderFunc, GRPCProviderV6Func, or ProviderFunc")
	}

	opts.Debug = true

	if opts.ProviderAddr == "" {
		opts.ProviderAddr = providerAddr
	}

	Serve(opts)

	return nil
}
