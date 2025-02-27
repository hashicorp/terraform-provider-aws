// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_default_tags", name="Default Tags")
func newDefaultTagsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &defaultTagsDataSource{}

	return d, nil
}

type defaultTagsDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *defaultTagsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *defaultTagsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data defaultTagsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	defaultTagsConfig := d.Meta().DefaultTagsConfig(ctx)
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)
	tags := defaultTagsConfig.GetTags()

	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, d.Meta().Partition(ctx))
	data.Tags = tftags.FlattenStringValueMap(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type defaultTagsDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Tags tftags.Map   `tfsdk:"tags"`
}
