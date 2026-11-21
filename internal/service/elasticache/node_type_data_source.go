// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache

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

// @FrameworkDataSource("aws_elasticache_node_type", name="Node Type")
func newNodeTypeDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &nodeTypeDataSource{}, nil
}

const (
	DSNameNodeType = "Node Type Data Source"
)

type nodeTypeDataSource struct {
	framework.DataSourceWithModel[nodeTypeDataSourceModel]
}

func (d *nodeTypeDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"burstable_performance_supported": schema.BoolAttribute{
				Computed: true,
			},
			"cache_node_type": schema.StringAttribute{
				Required: true,
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

func (d *nodeTypeDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data nodeTypeDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	cacheNodeType := data.CacheNodeType.ValueString()
	ec2InstanceType, ok := strings.CutPrefix(cacheNodeType, "cache.")
	if !ok || ec2InstanceType == "" {
		response.Diagnostics.AddAttributeError(
			path.Root("cache_node_type"),
			"Invalid Attribute Value",
			fmt.Sprintf(`cache_node_type must be an ElastiCache node type of the form "cache.<type>" (e.g. "cache.t3.medium"), got: %q`, cacheNodeType),
		)
		return
	}

	conn := d.Meta().EC2Client(ctx)

	v, err := findInstanceTypeByElastiCacheNodeType(ctx, conn, ec2InstanceType)
	if err != nil {
		err = fmt.Errorf("this ElastiCache node type may not have a directly corresponding EC2 instance type (%s): %w", ec2InstanceType, err)
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ElastiCache, create.ErrActionReading, DSNameNodeType, cacheNodeType, err),
			err.Error(),
		)
		return
	}

	archs := make([]string, len(v.ProcessorInfo.SupportedArchitectures))
	for i, a := range v.ProcessorInfo.SupportedArchitectures {
		archs[i] = string(a)
	}

	data.ID = types.StringValue(cacheNodeType)
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

type nodeTypeDataSourceModel struct {
	framework.WithRegionModel
	BurstablePerformanceSupported types.Bool           `tfsdk:"burstable_performance_supported"`
	CacheNodeType                 types.String         `tfsdk:"cache_node_type"`
	CurrentGeneration             types.Bool           `tfsdk:"current_generation"`
	DefaultCores                  types.Int64          `tfsdk:"default_cores"`
	DefaultThreadsPerCore         types.Int64          `tfsdk:"default_threads_per_core"`
	DefaultVCPUs                  types.Int64          `tfsdk:"default_vcpus"`
	EC2InstanceType               types.String         `tfsdk:"ec2_instance_type"`
	FreeTierEligible              types.Bool           `tfsdk:"free_tier_eligible"`
	ID                            types.String         `tfsdk:"id"`
	MemorySize                    types.Int64          `tfsdk:"memory_size"`
	SupportedArchitectures        fwtypes.ListOfString `tfsdk:"supported_architectures"`
}

func findInstanceTypeByElastiCacheNodeType(ctx context.Context, conn *ec2.Client, ec2InstanceType string) (*awstypes.InstanceTypeInfo, error) {
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
