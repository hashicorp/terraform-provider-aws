// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_servicecatalogappregistry_application", name="Application")
// @Tags(identifierAttribute="arn")
func newDataSourceApplication(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceApplication{}, nil
}

const (
	DSNameApplication = "Application Data Source"
)

type dataSourceApplication struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceApplication) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"application_tag": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}
func (d *dataSourceApplication) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ServiceCatalogAppRegistryClient(ctx)
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)

	var data dataSourceApplicationData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionSetting, ResNameApplication, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)

	// Transparent tagging doesn't work for DataSource yet
	data.Tags = tftags.NewMapFromMapValue(flex.FlattenFrameworkStringValueMapLegacy(ctx, KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceApplicationData struct {
	ARN            types.String `tfsdk:"arn"`
	Description    types.String `tfsdk:"description"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ApplicationTag types.Map    `tfsdk:"application_tag"`
	Tags           tftags.Map   `tfsdk:"tags"`
}
