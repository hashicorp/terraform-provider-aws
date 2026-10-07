// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_rds_instance_type", name="Instance Type")
func newInstanceTypeDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &instanceTypeDataSource{}, nil
}

const (
	DSNameInstanceType = "Instance Type Data Source"
)

type instanceTypeDataSource struct {
	framework.DataSourceWithModel[instanceTypeDataSourceModel]
}

func (d *instanceTypeDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"burstable_performance_supported": schema.BoolAttribute{
				Computed: true,
			},
			"current_generation": schema.BoolAttribute{
				Computed: true,
			},
			"default_cores": schema.Int64Attribute{
				Computed: true,
			},
			"default_threads_per_core": schema.Int64Attribute{
				Computed: true,
			},
			"default_vcpus": schema.Int64Attribute{
				Computed: true,
			},
			"ec2_instance_type": schema.StringAttribute{
				Computed: true,
			},
			"free_tier_eligible": schema.BoolAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"instance_class": schema.StringAttribute{
				Required: true,
			},
			"memory_size": schema.Int64Attribute{
				Computed: true,
			},
			"supported_architectures": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *instanceTypeDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data instanceTypeDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	instanceClass := data.InstanceClass.ValueString()
	ec2InstanceType, ok := strings.CutPrefix(instanceClass, "db.")
	if !ok || ec2InstanceType == "" {
		response.Diagnostics.AddAttributeError(
			path.Root("instance_class"),
			"Invalid Attribute Value",
			fmt.Sprintf(`instance_class must be an RDS DB instance class of the form "db.<type>" (e.g. "db.t3.medium"), got: %q`, instanceClass),
		)
		return
	}

	conn := d.Meta().EC2Client(ctx)

	v, err := findInstanceTypeByRDSInstanceClass(ctx, conn, ec2InstanceType)
	if err != nil {
		err = fmt.Errorf("this RDS instance class may not have a directly corresponding EC2 instance type (%s): %w", ec2InstanceType, err)
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionReading, DSNameInstanceType, instanceClass, err),
			err.Error(),
		)
		return
	}

	archs := make([]string, len(v.ProcessorInfo.SupportedArchitectures))
	for i, a := range v.ProcessorInfo.SupportedArchitectures {
		archs[i] = string(a)
	}

	data.ID = types.StringValue(instanceClass)
	data.BurstablePerformanceSupported = types.BoolValue(aws.ToBool(v.BurstablePerformanceSupported))
	data.CurrentGeneration = types.BoolValue(aws.ToBool(v.CurrentGeneration))
	data.DefaultCores = types.Int64Value(int64(aws.ToInt32(v.VCpuInfo.DefaultCores)))
	data.DefaultThreadsPerCore = types.Int64Value(int64(aws.ToInt32(v.VCpuInfo.DefaultThreadsPerCore)))
	data.DefaultVCPUs = types.Int64Value(int64(aws.ToInt32(v.VCpuInfo.DefaultVCpus)))
	data.EC2InstanceType = types.StringValue(string(v.InstanceType))
	data.FreeTierEligible = types.BoolValue(aws.ToBool(v.FreeTierEligible))
	data.MemorySize = types.Int64Value(aws.ToInt64(v.MemoryInfo.SizeInMiB))
	data.SupportedArchitectures = flex.FlattenFrameworkStringValueListOfString(ctx, archs)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type instanceTypeDataSourceModel struct {
	framework.WithRegionModel
	BurstablePerformanceSupported types.Bool           `tfsdk:"burstable_performance_supported"`
	CurrentGeneration             types.Bool           `tfsdk:"current_generation"`
	DefaultCores                  types.Int64          `tfsdk:"default_cores"`
	DefaultThreadsPerCore         types.Int64          `tfsdk:"default_threads_per_core"`
	DefaultVCPUs                  types.Int64          `tfsdk:"default_vcpus"`
	EC2InstanceType               types.String         `tfsdk:"ec2_instance_type"`
	FreeTierEligible              types.Bool           `tfsdk:"free_tier_eligible"`
	ID                            types.String         `tfsdk:"id"`
	InstanceClass                 types.String         `tfsdk:"instance_class"`
	MemorySize                    types.Int64          `tfsdk:"memory_size"`
	SupportedArchitectures        fwtypes.ListOfString `tfsdk:"supported_architectures"`
}

func findInstanceTypeByRDSInstanceClass(ctx context.Context, conn *ec2.Client, ec2InstanceType string) (*awstypes.InstanceTypeInfo, error) {
	input := ec2.DescribeInstanceTypesInput{
		InstanceTypes: []awstypes.InstanceType{awstypes.InstanceType(ec2InstanceType)},
	}

	output, err := conn.DescribeInstanceTypes(ctx, &input)
	if err != nil {
		return nil, err
	}

	v, err := tfresource.AssertSingleValueResult(output.InstanceTypes)
	if err != nil {
		return nil, err
	}

	if v.MemoryInfo == nil || v.VCpuInfo == nil || v.ProcessorInfo == nil {
		return nil, fmt.Errorf("incomplete instance type information returned for %q", ec2InstanceType)
	}

	return v, nil
}
