// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Application Providers")
func newDataSourceApplicationProviders(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceApplicationProviders{}, nil
}

const (
	DSNameApplicationProviders = "Application Providers Data Source"
)

type dataSourceApplicationProviders struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceApplicationProviders) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_ssoadmin_application_providers"
}

func (d *dataSourceApplicationProviders) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"application_providers": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"application_provider_arn": framework.ARNAttributeComputedOnly(),
						"federation_protocol": schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"display_data": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDescription: schema.StringAttribute{
										Computed: true,
									},
									names.AttrDisplayName: schema.StringAttribute{
										Computed: true,
									},
									"icon_url": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceApplicationProviders) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SSOAdminClient(ctx)

	var data dataSourceApplicationProvidersData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(d.Meta().Region)

	paginator := ssoadmin.NewListApplicationProvidersPaginator(conn, &ssoadmin.ListApplicationProvidersInput{})
	var apiObjects []awstypes.ApplicationProvider
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionReading, DSNameApplicationProviders, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		if page != nil {
			apiObjects = append(apiObjects, page.ApplicationProviders...)
		}
	}

	applicationProviders, diag := flattenApplicationProviders(ctx, apiObjects)
	resp.Diagnostics.Append(diag...)
	data.ApplicationProviders = applicationProviders

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceApplicationProvidersData struct {
	ApplicationProviders types.List   `tfsdk:"application_providers"`
	ID                   types.String `tfsdk:"id"`
}

var applicationProviderAttrTypes = map[string]attr.Type{
	"application_provider_arn": types.StringType,
	"display_data":             types.ListType{ElemType: types.ObjectType{AttrTypes: displayDataAttrTypes}},
	"federation_protocol":      types.StringType,
}

var displayDataAttrTypes = map[string]attr.Type{
	names.AttrDescription: types.StringType,
	names.AttrDisplayName: types.StringType,
	"icon_url":            types.StringType,
}

func flattenApplicationProviders(ctx context.Context, apiObjects []awstypes.ApplicationProvider) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: applicationProviderAttrTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		displayData, d := flattenDisplayData(ctx, apiObject.DisplayData)
		diags.Append(d...)

		obj := map[string]attr.Value{
			"application_provider_arn": flex.StringToFramework(ctx, apiObject.ApplicationProviderArn),
			"display_data":             displayData,
			"federation_protocol":      flex.StringValueToFramework(ctx, apiObject.FederationProtocol),
		}
		objVal, d := types.ObjectValue(applicationProviderAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenDisplayData(ctx context.Context, apiObject *awstypes.DisplayData) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: displayDataAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		names.AttrDescription: flex.StringToFramework(ctx, apiObject.Description),
		names.AttrDisplayName: flex.StringToFramework(ctx, apiObject.DisplayName),
		"icon_url":            flex.StringToFramework(ctx, apiObject.IconUrl),
	}
	objVal, d := types.ObjectValue(displayDataAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}
