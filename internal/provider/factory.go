package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-mux/tf6to5server"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/fw"
)

// ProtoV5ProviderServerFactory returns a muxed terraform-plugin-go protocol v5 provider factory function.
// This factory function is suitable for use with the terraform-plugin-go Serve function.
func ProtoV5ProviderServerFactory(ctx context.Context) (func() tfprotov5.ProviderServer, error) {
	downgradedFrameworkProvider, err := tf6to5server.DowngradeServer(ctx, providerserver.NewProtocol6(fw.New()))

	if err != nil {
		return nil, err
	}

	providers := []func() tfprotov5.ProviderServer{
		Provider().GRPCProvider,
		func() tfprotov5.ProviderServer { return downgradedFrameworkProvider },
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer, nil
}
