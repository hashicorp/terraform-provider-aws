package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	// "github.com/hashicorp/terraform-provider-aws/internal/tf5provider"
)

// ProtoV5ProviderServerFactory returns a terraform-plugin-go protocol v5 provider factory function.
func ProtoV5ProviderServerFactory(ctx context.Context) (func() tfprotov5.ProviderServer, error) {
	providers := []func() tfprotov5.ProviderServer{
		// tf5provider.Provider().GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer, nil
}
