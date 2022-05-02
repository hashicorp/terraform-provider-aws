package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-mux/tf6to5server"
	"github.com/hashicorp/terraform-provider-aws/internal/tf5provider"
	"github.com/hashicorp/terraform-provider-aws/internal/tf6provider"
)

// ProtoV5ProviderServerFactory returns a terraform-plugin-go protocol v5 provider factory function.
func ProtoV5ProviderServerFactory(ctx context.Context) (func() tfprotov5.ProviderServer, error) {
	downgradedProvider, err := tf6to5server.DowngradeServer(ctx, providerserver.NewProtocol6(tf6provider.New()))

	if err != nil {
		return nil, err
	}

	providers := []func() tfprotov5.ProviderServer{
		func() tfprotov5.ProviderServer { return downgradedProvider },
		tf5provider.Provider().GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer, nil
}
