// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
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

// @FrameworkDataSource(name="Supported Instance Types")
func newDataSourceSupportedInstanceTypes(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceSupportedInstanceTypes{}, nil
}

const (
	DSNameSupportedInstanceTypes = "Supported Instance Types Data Source"
)

type dataSourceSupportedInstanceTypes struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceSupportedInstanceTypes) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_emr_supported_instance_types"
}

func (d *dataSourceSupportedInstanceTypes) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"release_label": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"supported_instance_types": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"architecture": schema.StringAttribute{
							Computed: true,
						},
						"ebs_optimized_available": schema.BoolAttribute{
							Computed: true,
						},
						"ebs_optimized_by_default": schema.BoolAttribute{
							Computed: true,
						},
						"ebs_storage_only": schema.BoolAttribute{
							Computed: true,
						},
						"instance_family_id": schema.StringAttribute{
							Computed: true,
						},
						"is_64_bits_only": schema.BoolAttribute{
							Computed: true,
						},
						"memory_gb": schema.Float64Attribute{
							Computed: true,
						},
						"number_of_disks": schema.Int64Attribute{
							Computed: true,
						},
						"storage_gb": schema.Int64Attribute{
							Computed: true,
						},
						names.AttrType: schema.StringAttribute{
							Computed: true,
						},
						"vcpu": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceSupportedInstanceTypes) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EMRClient(ctx)

	var data dataSourceSupportedInstanceTypesData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(data.ReleaseLabel.ValueString())

	input := &emr.ListSupportedInstanceTypesInput{
		ReleaseLabel: aws.String(data.ReleaseLabel.ValueString()),
	}

	var results []awstypes.SupportedInstanceType
	paginator := emr.NewListSupportedInstanceTypesPaginator(conn, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EMR, create.ErrActionReading, DSNameSupportedInstanceTypes, data.ID.String(), err),
				err.Error(),
			)
			return
		}
		results = append(results, output.SupportedInstanceTypes...)
	}

	supportedInstanceTypes, diag := flattenSupportedInstanceTypes(ctx, results)
	resp.Diagnostics.Append(diag...)
	data.SupportedInstanceTypes = supportedInstanceTypes

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceSupportedInstanceTypesData struct {
	ID                     types.String `tfsdk:"id"`
	ReleaseLabel           types.String `tfsdk:"release_label"`
	SupportedInstanceTypes types.List   `tfsdk:"supported_instance_types"`
}

var supportedInstanceTypeAttrTypes = map[string]attr.Type{
	"architecture":             types.StringType,
	"ebs_optimized_available":  types.BoolType,
	"ebs_optimized_by_default": types.BoolType,
	"ebs_storage_only":         types.BoolType,
	"instance_family_id":       types.StringType,
	"is_64_bits_only":          types.BoolType,
	"memory_gb":                types.Float64Type,
	"number_of_disks":          types.Int64Type,
	"storage_gb":               types.Int64Type,
	names.AttrType:             types.StringType,
	"vcpu":                     types.Int64Type,
}

func flattenSupportedInstanceTypes(ctx context.Context, apiObjects []awstypes.SupportedInstanceType) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: supportedInstanceTypeAttrTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"architecture":             flex.StringToFramework(ctx, apiObject.Architecture),
			"ebs_optimized_available":  flex.BoolToFramework(ctx, apiObject.EbsOptimizedAvailable),
			"ebs_optimized_by_default": flex.BoolToFramework(ctx, apiObject.EbsOptimizedByDefault),
			"ebs_storage_only":         flex.BoolToFramework(ctx, apiObject.EbsStorageOnly),
			"instance_family_id":       flex.StringToFramework(ctx, apiObject.InstanceFamilyId),
			"is_64_bits_only":          flex.BoolToFramework(ctx, apiObject.Is64BitsOnly),
			"memory_gb":                flex.Float32ToFramework(ctx, apiObject.MemoryGB),
			"number_of_disks":          flex.Int32ToFramework(ctx, apiObject.NumberOfDisks),
			"storage_gb":               flex.Int32ToFramework(ctx, apiObject.StorageGB),
			names.AttrType:             flex.StringToFramework(ctx, apiObject.Type),
			"vcpu":                     flex.Int32ToFramework(ctx, apiObject.VCPU),
		}
		objVal, d := types.ObjectValue(supportedInstanceTypeAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}
