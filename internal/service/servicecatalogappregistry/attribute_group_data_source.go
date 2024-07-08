// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_servicecatalogappregistry_attribute_group", name="Attribute Group")
// @Tags(identifierAttribute="arn")
func newDataSourceAttributeGroup(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAttributeGroup{}, nil
}

const (
	DSNameAttributeGroup = "Attribute Group Data Source"
)

type dataSourceAttributeGroup struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceAttributeGroup) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_servicecatalogappregistry_attribute_group"
}

func (d *dataSourceAttributeGroup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAttributes: schema.StringAttribute{
				Computed:   true,
				CustomType: jsontypes.NormalizedType{},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot(names.AttrName), path.MatchRoot(names.AttrARN)),
				},
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *dataSourceAttributeGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ServiceCatalogAppRegistryClient(ctx)
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig

	var data dataSourceAttributeGroupData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var id string

	if !data.ID.IsNull() {
		id = data.ID.ValueString()
	} else if !data.Name.IsNull() {
		id = data.Name.ValueString()
	} else if !data.ARN.IsNull() {
		id = data.ARN.ValueString()
	}

	out, err := findAttributeGroupByID(ctx, conn, id)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionReading, DSNameAttributeGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Transparent tagging doesn't work for DataSource yet
	data.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceAttributeGroupData struct {
	ARN         types.String         `tfsdk:"arn"`
	Attributes  jsontypes.Normalized `tfsdk:"attributes"`
	Description types.String         `tfsdk:"description"`
	ID          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Tags        types.Map            `tfsdk:"tags"`
}
