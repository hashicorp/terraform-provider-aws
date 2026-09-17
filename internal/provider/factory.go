// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2"
)

// ProtoV6ProviderServerFactory returns a muxed terraform-plugin-go protocol v6 provider factory function.
// This factory function is suitable for use with the terraform-plugin-go Serve function.
// The primary (Plugin SDK) provider server is also returned (useful for testing).
func ProtoV6ProviderServerFactory(ctx context.Context) (func() tfprotov6.ProviderServer, *schema.Provider, error) {
	internal.RegisterSmarterrFS()

	primary, err := sdkv2.NewProvider(ctx)

	if err != nil {
		return nil, nil, err
	}

	upgradedSDKServer, err := tf5to6server.UpgradeServer(
		ctx,
		primary.GRPCProvider,
	)

	secondary, err := framework.NewProvider(ctx, primary)

	if err != nil {
		return nil, nil, err
	}

	servers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer {
			return upgradedSDKServer
		},
		providerserver.NewProtocol6(secondary),
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, servers...)

	if err != nil {
		return nil, nil, err
	}

	return muxServer.ProviderServer, primary, nil // TODO: primary is returned for testing - assert tests still work with upgradedSDKServer
}
