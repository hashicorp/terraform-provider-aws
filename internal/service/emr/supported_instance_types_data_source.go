// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_emr_supported_instance_types", name="Supported Instance Types")
func newSupportedInstanceTypesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &supportedInstanceTypesDataSource{}, nil
}

type supportedInstanceTypesDataSource struct {
	framework.DataSourceWithModel[supportedInstanceTypesDataSourceModel]
}

func (d *supportedInstanceTypesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"release_label": schema.StringAttribute{
				Required: true,
			},
			"supported_instance_types": framework.ResourceComputedListOfObjectsAttribute[supportedInstanceTypeModel](ctx),
		},
	}
}
func (d *supportedInstanceTypesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data supportedInstanceTypesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EMRClient(ctx)

	input := emr.ListSupportedInstanceTypesInput{
		ReleaseLabel: fwflex.StringFromFramework(ctx, data.ReleaseLabel),
	}
	output, err := findSupportedInstanceTypes(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EMR Supported Instance Types (%s)", data.ReleaseLabel.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.SupportedInstanceTypes)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = data.ReleaseLabel

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findSupportedInstanceTypes(ctx context.Context, conn *emr.Client, input *emr.ListSupportedInstanceTypesInput) ([]awstypes.SupportedInstanceType, error) {
	var output []awstypes.SupportedInstanceType

	pages := emr.NewListSupportedInstanceTypesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.SupportedInstanceTypes...)
	}

	return output, nil
}

type supportedInstanceTypesDataSourceModel struct {
	framework.WithRegionModel
	ID                     types.String                                                `tfsdk:"id"`
	ReleaseLabel           types.String                                                `tfsdk:"release_label"`
	SupportedInstanceTypes fwtypes.ListNestedObjectValueOf[supportedInstanceTypeModel] `tfsdk:"supported_instance_types"`
}

type supportedInstanceTypeModel struct {
	Architecture          types.String  `tfsdk:"architecture"`
	EBSOptimizedAvailable types.Bool    `tfsdk:"ebs_optimized_available"`
	EBSOptimizedByDefault types.Bool    `tfsdk:"ebs_optimized_by_default"`
	EBSStorageOnly        types.Bool    `tfsdk:"ebs_storage_only"`
	InstanceFamilyID      types.String  `tfsdk:"instance_family_id"`
	Is64BitsOnly          types.Bool    `tfsdk:"is_64_bits_only"`
	MemoryGB              types.Float64 `tfsdk:"memory_gb"`
	NumberOfDisks         types.Int64   `tfsdk:"number_of_disks"`
	StorageGB             types.Int64   `tfsdk:"storage_gb"`
	Type                  types.String  `tfsdk:"type"`
	VCPU                  types.Int64   `tfsdk:"vcpu"`
}
