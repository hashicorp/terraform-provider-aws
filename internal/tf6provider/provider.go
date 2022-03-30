package tf6provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct{}

func (p *provider) Configure(ctx context.Context, request tfsdk.ConfigureProviderRequest, response *tfsdk.ConfigureProviderResponse) {
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	var diags diag.Diagnostics
	dataSources := make(map[string]tfsdk.DataSourceType)

	return dataSources, diags
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	var diags diag.Diagnostics
	resources := make(map[string]tfsdk.ResourceType)

	return resources, diags
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	var diags diag.Diagnostics
	schema := tfsdk.Schema{
		Version: 1,
	}

	return schema, diags
}
