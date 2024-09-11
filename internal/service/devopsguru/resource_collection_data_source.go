// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Resource Collection")
func newDataSourceResourceCollection(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceResourceCollection{}, nil
}

const (
	DSNameResourceCollection = "Resource Collection Data Source"
)

type dataSourceResourceCollection struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceResourceCollection) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_devopsguru_resource_collection"
}

func (d *dataSourceResourceCollection) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrType: schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ResourceCollectionType](),
			},
		},
		Blocks: map[string]schema.Block{
			"cloudformation": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[cloudformationData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"stack_names": schema.ListAttribute{
							Computed:    true,
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
						},
					},
				},
			},
			names.AttrTags: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tagsData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"app_boundary_key": schema.StringAttribute{
							Computed: true,
						},
						"tag_values": schema.ListAttribute{
							Computed:    true,
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceResourceCollection) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().DevOpsGuruClient(ctx)

	var data dataSourceResourceCollectionData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(data.Type.ValueString())

	out, err := findResourceCollectionByID(ctx, conn, data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionReading, DSNameResourceCollection, data.Type.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fields named "Tags" are currently hardcoded to be ignored by AutoFlex. Flattening the Tags
	// struct from the response into state.Tags is a temporary workaround until the AutoFlex
	// options implementation can be merged.
	//
	// Ref: https://github.com/hashicorp/terraform-provider-aws/pull/36437
	resp.Diagnostics.Append(flex.Flatten(ctx, out.Tags, &data.Tags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceResourceCollectionData struct {
	CloudFormation fwtypes.ListNestedObjectValueOf[cloudformationData] `tfsdk:"cloudformation"`
	ID             types.String                                        `tfsdk:"id"`
	Tags           fwtypes.ListNestedObjectValueOf[tagsData]           `tfsdk:"tags"`
	Type           fwtypes.StringEnum[awstypes.ResourceCollectionType] `tfsdk:"type"`
}
