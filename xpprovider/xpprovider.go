// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package xpprovider exports needed internal types and functions used by Crossplane for instantiating, interacting and
// configuring the underlying Terraform AWS providers.
package xpprovider

import (
	"context"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	internalfwprovider "github.com/hashicorp/terraform-provider-aws/internal/provider/fwprovider"
)

// AWSConfig exports the internal type conns.Config of the Terraform provider
type AWSConfig conns.Config

// AWSClient exports the internal type conns.AWSClient of the Terraform provider
type AWSClient = conns.AWSClient

// GetProvider returns new provider instances for both Terraform Plugin Framework provider of type provider.Provider
// and Terraform Plugin SDKv2 provider of type *schema.Provider
// provider
func GetProvider(ctx context.Context) (fwprovider.Provider, *schema.Provider, error) {
	p, err := provider.New(ctx)
	if err != nil {
		return nil, nil, err
	}
	fwProvider := internalfwprovider.New(p)
	return fwProvider, p, err
}

// GetProviderSchema returns the Terraform Plugin SDKv2 provider schema of the provider
func GetProviderSchema(ctx context.Context) (*schema.Provider, error) {
	return provider.New(ctx)
}

// GetFrameworkProviderSchema returns the Terraform Plugin Framework provider schema of the provider
func GetFrameworkProviderSchema(ctx context.Context) (fwschema.Schema, error) {
	fwProvider, _, err := GetProvider(ctx)
	if err != nil {
		return fwschema.Schema{}, err
	}
	schemaReq := fwprovider.SchemaRequest{}
	schemaResp := fwprovider.SchemaResponse{}
	fwProvider.Schema(ctx, schemaReq, &schemaResp)
	return schemaResp.Schema, nil
}

// GetFrameworkProviderWithMeta returns a new Terraform Plugin Framework-style provider instance with the given
// provider meta (AWS client). Supplied meta can be any type implementing Meta(), that returns a configured AWS Client
// of type *conns.AWSClient
// Can be used to create provider instances with arbitrary AWS clients.
func GetFrameworkProviderWithMeta(primary interface{ Meta() interface{} }) fwprovider.Provider {
	return internalfwprovider.New(primary)
}

// GetClient configures the supplied provider meta (in the *AWSClient). It is a wrapper function that exports
// the internal type conns.Config's ConfigureProvider() func, over the exported type AWSConfig
// supplied *AWSClient in the arguments, is an in-out argument and is returned back as internal type *conns.AWSClient
func (ac *AWSConfig) GetClient(ctx context.Context, client *AWSClient) (*conns.AWSClient, diag.Diagnostics) {
	return (*conns.Config)(ac).ConfigureProvider(ctx, (*conns.AWSClient)(client))
}
