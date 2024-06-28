// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)


// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource(name="Attribute Group")
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
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Computed: true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"type": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"complex_argument": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"nested_required": schema.StringAttribute{
							Computed: true,
						},
						"nested_computed": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceAttributeGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ServiceCatalogAppRegistryClient(ctx)
	
	var data dataSourceAttributeGroupData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	out, err := findAttributeGroupByName(ctx, conn, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionReading, DSNameAttributeGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}
	
	data.ARN = flex.StringToFramework(ctx, out.Arn)
	data.ID = flex.StringToFramework(ctx, out.AttributeGroupId)
	data.Name = flex.StringToFramework(ctx, out.AttributeGroupName)
	data.Type = flex.StringToFramework(ctx, out.AttributeGroupType)
	
	complexArgument, diag := flattenComplexArgument(ctx, out.ComplexArgument)
	resp.Diagnostics.Append(diag...)
	data.ComplexArgument = complexArgument
	
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig
	tags := KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	data.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.Map())
	
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}


type dataSourceAttributeGroupData struct {
	ARN             types.String   `tfsdk:"arn"`
	ComplexArgument types.List     `tfsdk:"complex_argument"`
	Description     types.String   `tfsdk:"description"`
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Tags            types.Map      `tfsdk:"tags"`
	Type            types.String   `tfsdk:"type"`
}

type complexArgumentData struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}