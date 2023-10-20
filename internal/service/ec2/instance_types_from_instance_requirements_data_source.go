// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Instance Types From Instance Requirements")
func newDataSourceInstanceTypesFromInstanceRequirements(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceInstanceTypesFromInstanceRequirements{}, nil
}

const (
	DSNameInstanceTypesFromInstanceRequirements = "Instance Types From Instance Requirements Data Source"
)

type dataSourceInstanceTypesFromInstanceRequirements struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceInstanceTypesFromInstanceRequirements) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_ec2_instance_types_from_instance_requirements"
}

func (d *dataSourceInstanceTypesFromInstanceRequirements) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"architecture_types": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"id": framework.IDAttribute(),
			"instance_types": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"virtualization_types": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"instance_requirements": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"accelerator_manufacturers": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"accelerator_names": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"accelerator_types": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"allowed_instance_types": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"bare_metal": schema.StringAttribute{
						Optional: true,
					},
					"burstable_performance": schema.StringAttribute{
						Optional: true,
					},
					"cpu_manufacturers": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"excluded_instance_types": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"instance_generations": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"local_storage": schema.StringAttribute{
						Optional: true,
					},
					"local_storage_types": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"on_demand_max_price_percentage_over_lowest_price": schema.Int64Attribute{
						Optional: true,
					},
					"require_hibernate_support": schema.BoolAttribute{
						Optional: true,
					},
					"spot_max_price_percentage_over_lowest_price": schema.Int64Attribute{
						Optional: true,
					},
				},
				Blocks: map[string]schema.Block{
					"memory_mib": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min": schema.Int64Attribute{
								Required: true,
							},
							"max": schema.Int64Attribute{
								Optional: true,
							},
						},
					},
					"vcpu_count": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min": schema.Int64Attribute{
								Required: true,
							},
							"max": schema.Int64Attribute{
								Optional: true,
							},
						},
					},
					"accelerator_count": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min": schema.Int64Attribute{
								Optional: true,
							},
							"max": schema.Int64Attribute{
								Optional: true,
							},
						},
					},
					"accelerator_total_memory_mib": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min": schema.Int64Attribute{
								Optional: true,
							},
							"max": schema.Int64Attribute{
								Optional: true,
							},
						},
					},
					"baseline_ebs_bandwidth_mbps": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min": schema.Int64Attribute{
								Optional: true,
							},
							"max": schema.Int64Attribute{
								Optional: true,
							},
						},
					},
					"memory_gib_per_vcpu": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min": schema.Float64Attribute{
								Optional: true,
							},
							"max": schema.Float64Attribute{
								Optional: true,
							},
						},
					},
					"network_bandwidth_gbps": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min": schema.Float64Attribute{
								Optional: true,
							},
							"max": schema.Float64Attribute{
								Optional: true,
							},
						},
					},
					"network_interface_count": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min": schema.Int64Attribute{
								Optional: true,
							},
							"max": schema.Int64Attribute{
								Optional: true,
							},
						},
					},
					"total_local_storage_gb": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min": schema.Float64Attribute{
								Optional: true,
							},
							"max": schema.Float64Attribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceInstanceTypesFromInstanceRequirements) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EC2Conn(ctx)

	var data dataSourceInstanceTypesFromInstanceRequirementsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &ec2.GetInstanceTypesFromInstanceRequirementsInput{
		ArchitectureTypes:    flex.ExpandFrameworkStringList(ctx, data.ArchitectureTypes),
		VirtualizationTypes:  flex.ExpandFrameworkStringList(ctx, data.VirtualizationTypes),
		InstanceRequirements: expandInstanceRequirementsRequestOptions(ctx, data.InstanceRequirements, &resp.Diagnostics),
	}

	out, err := findInstanceTypesFromInstanceRequirements(ctx, conn, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, DSNameInstanceTypesFromInstanceRequirements, d.Meta().Region, err),
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(d.Meta().Region)
	data.InstanceTypes = flex.FlattenFrameworkStringList(ctx, out)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func expandInstanceRequirementsRequestOptions(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.InstanceRequirementsRequest {
	var options instanceRequirementsData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx, diags)
}

func findInstanceTypesFromInstanceRequirements(ctx context.Context, conn *ec2.EC2, input *ec2.GetInstanceTypesFromInstanceRequirementsInput) ([]*string, error) {
	var output []*string

	err := conn.GetInstanceTypesFromInstanceRequirementsPagesWithContext(ctx, input, func(page *ec2.GetInstanceTypesFromInstanceRequirementsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceTypes {
			if v != nil {
				output = append(output, v.InstanceType)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

type dataSourceInstanceTypesFromInstanceRequirementsData struct {
	ArchitectureTypes    types.List   `tfsdk:"architecture_types"`
	InstanceRequirements types.Object `tfsdk:"instance_requirements"`
	InstanceTypes        types.List   `tfsdk:"instance_types"`
	ID                   types.String `tfsdk:"id"`
	VirtualizationTypes  types.List   `tfsdk:"virtualization_types"`
}

type instanceRequirementsData struct {
	MemoryMiB                                 types.Object `tfsdk:"memory_mib"`
	VCpuCount                                 types.Object `tfsdk:"vcpu_count"`
	AcceleratorCount                          types.Object `tfsdk:"accelerator_count"`
	AcceleratorManufacturers                  types.List   `tfsdk:"accelerator_manufacturers"`
	AcceleratorNames                          types.List   `tfsdk:"accelerator_names"`
	AcceleratorTotalMemoryMiB                 types.Object `tfsdk:"accelerator_total_memory_mib"`
	AcceleratorTypes                          types.List   `tfsdk:"accelerator_types"`
	AllowedInstanceTypes                      types.List   `tfsdk:"allowed_instance_types"`
	BareMetal                                 types.String `tfsdk:"bare_metal"`
	BaselineEbsBandwidthMbps                  types.Object `tfsdk:"baseline_ebs_bandwidth_mbps"`
	BurstablePerformance                      types.String `tfsdk:"burstable_performance"`
	CpuManufacturers                          types.List   `tfsdk:"cpu_manufacturers"`
	ExcludedInstanceTypes                     types.List   `tfsdk:"excluded_instance_types"`
	InstanceGenerations                       types.List   `tfsdk:"instance_generations"`
	LocalStorage                              types.String `tfsdk:"local_storage"`
	LocalStorageTypes                         types.List   `tfsdk:"local_storage_types"`
	MemoryGiBPerVCpu                          types.Object `tfsdk:"memory_gib_per_vcpu"`
	NetworkBandwidthGbps                      types.Object `tfsdk:"network_bandwidth_gbps"`
	NetworkInterfaceCount                     types.Object `tfsdk:"network_interface_count"`
	OnDemandMaxPricePercentageOverLowestPrice types.Int64  `tfsdk:"on_demand_max_price_percentage_over_lowest_price"`
	RequireHibernateSupport                   types.Bool   `tfsdk:"require_hibernate_support"`
	SpotMaxPricePercentageOverLowestPrice     types.Int64  `tfsdk:"spot_max_price_percentage_over_lowest_price"`
	TotalLocalStorageGB                       types.Object `tfsdk:"total_local_storage_gb"`
}

func (ir *instanceRequirementsData) expand(ctx context.Context, diags *diag.Diagnostics) *ec2.InstanceRequirementsRequest {
	if ir == nil {
		return nil
	}

	result := &ec2.InstanceRequirementsRequest{
		MemoryMiB:                expandMemoryMiBData(ctx, ir.MemoryMiB, diags),
		VCpuCount:                expandVcpuCountData(ctx, ir.VCpuCount, diags),
		AcceleratorManufacturers: flex.ExpandFrameworkStringList(ctx, ir.AcceleratorManufacturers),
		AcceleratorNames:         flex.ExpandFrameworkStringList(ctx, ir.AcceleratorNames),
		AcceleratorTypes:         flex.ExpandFrameworkStringList(ctx, ir.AcceleratorTypes),
		AllowedInstanceTypes:     flex.ExpandFrameworkStringList(ctx, ir.AllowedInstanceTypes),
		BareMetal:                flex.StringFromFramework(ctx, ir.BareMetal),
		BurstablePerformance:     flex.StringFromFramework(ctx, ir.BurstablePerformance),
		CpuManufacturers:         flex.ExpandFrameworkStringList(ctx, ir.CpuManufacturers),
		ExcludedInstanceTypes:    flex.ExpandFrameworkStringList(ctx, ir.ExcludedInstanceTypes),
		InstanceGenerations:      flex.ExpandFrameworkStringList(ctx, ir.InstanceGenerations),
		LocalStorage:             flex.StringFromFramework(ctx, ir.LocalStorage),
		LocalStorageTypes:        flex.ExpandFrameworkStringList(ctx, ir.LocalStorageTypes),
		OnDemandMaxPricePercentageOverLowestPrice: flex.Int64FromFramework(ctx, ir.OnDemandMaxPricePercentageOverLowestPrice),
		RequireHibernateSupport:                   flex.BoolFromFramework(ctx, ir.RequireHibernateSupport),
		SpotMaxPricePercentageOverLowestPrice:     flex.Int64FromFramework(ctx, ir.SpotMaxPricePercentageOverLowestPrice),
	}

	if !ir.AcceleratorCount.IsNull() {
		result.AcceleratorCount = expandAcceleratorCountData(ctx, ir.AcceleratorCount, diags)
	}

	if !ir.AcceleratorTotalMemoryMiB.IsNull() {
		result.AcceleratorTotalMemoryMiB = expandAcceleratorTotalMemoryMiBData(ctx, ir.AcceleratorTotalMemoryMiB, diags)
	}

	if !ir.BaselineEbsBandwidthMbps.IsNull() {
		result.BaselineEbsBandwidthMbps = expandBaselineEBSBandwidthMbpsData(ctx, ir.BaselineEbsBandwidthMbps, diags)
	}

	if !ir.MemoryGiBPerVCpu.IsNull() {
		result.MemoryGiBPerVCpu = expandMemoryGiBPerVCpuData(ctx, ir.MemoryGiBPerVCpu, diags)
	}

	if !ir.NetworkBandwidthGbps.IsNull() {
		result.NetworkBandwidthGbps = expandNetworkBandwidthGbpsData(ctx, ir.NetworkBandwidthGbps, diags)
	}

	if !ir.NetworkInterfaceCount.IsNull() {
		result.NetworkInterfaceCount = expandNetworkInterfaceCountData(ctx, ir.NetworkInterfaceCount, diags)
	}

	if !ir.TotalLocalStorageGB.IsNull() {
		result.TotalLocalStorageGB = expandTotalLocalStorageGBData(ctx, ir.TotalLocalStorageGB, diags)
	}

	return result
}

type memoryMiBData struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}

func (m *memoryMiBData) expand(ctx context.Context) *ec2.MemoryMiBRequest {
	if m == nil {
		return nil
	}
	return &ec2.MemoryMiBRequest{
		Min: flex.Int64FromFramework(ctx, m.Min),
		Max: flex.Int64FromFramework(ctx, m.Max),
	}
}

func expandMemoryMiBData(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.MemoryMiBRequest {
	var options memoryMiBData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx)
}

type vcpuCountData struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}

func (v *vcpuCountData) expand(ctx context.Context) *ec2.VCpuCountRangeRequest {
	if v == nil {
		return nil
	}
	return &ec2.VCpuCountRangeRequest{
		Min: flex.Int64FromFramework(ctx, v.Min),
		Max: flex.Int64FromFramework(ctx, v.Max),
	}
}

func expandVcpuCountData(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.VCpuCountRangeRequest {
	var options vcpuCountData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx)
}

type acceleratorCountData struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}

func (a *acceleratorCountData) expand(ctx context.Context) *ec2.AcceleratorCountRequest {
	if a == nil {
		return nil
	}
	return &ec2.AcceleratorCountRequest{
		Min: flex.Int64FromFramework(ctx, a.Min),
		Max: flex.Int64FromFramework(ctx, a.Max),
	}
}

func expandAcceleratorCountData(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.AcceleratorCountRequest {
	var options acceleratorCountData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx)
}

type acceleratorTotalMemoryMiBData struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}

func (a *acceleratorTotalMemoryMiBData) expand(ctx context.Context) *ec2.AcceleratorTotalMemoryMiBRequest {
	if a == nil {
		return nil
	}
	return &ec2.AcceleratorTotalMemoryMiBRequest{
		Min: flex.Int64FromFramework(ctx, a.Min),
		Max: flex.Int64FromFramework(ctx, a.Max),
	}
}

func expandAcceleratorTotalMemoryMiBData(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.AcceleratorTotalMemoryMiBRequest {
	var options acceleratorTotalMemoryMiBData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx)
}

type baselineEbsBandwidthMbpsData struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}

func (b *baselineEbsBandwidthMbpsData) expand(ctx context.Context) *ec2.BaselineEbsBandwidthMbpsRequest {
	if b == nil {
		return nil
	}
	return &ec2.BaselineEbsBandwidthMbpsRequest{
		Min: flex.Int64FromFramework(ctx, b.Min),
		Max: flex.Int64FromFramework(ctx, b.Max),
	}
}

func expandBaselineEBSBandwidthMbpsData(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.BaselineEbsBandwidthMbpsRequest {
	var options baselineEbsBandwidthMbpsData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx)
}

type memoryGiBPerVCpuData struct {
	Min types.Float64 `tfsdk:"min"`
	Max types.Float64 `tfsdk:"max"`
}

func (m *memoryGiBPerVCpuData) expand(ctx context.Context) *ec2.MemoryGiBPerVCpuRequest {
	if m == nil {
		return nil
	}
	return &ec2.MemoryGiBPerVCpuRequest{
		Min: flex.Float64FromFramework(ctx, m.Min),
		Max: flex.Float64FromFramework(ctx, m.Max),
	}
}

func expandMemoryGiBPerVCpuData(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.MemoryGiBPerVCpuRequest {
	var options memoryGiBPerVCpuData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx)
}

type networkBandwidthGbpsData struct {
	Min types.Float64 `tfsdk:"min"`
	Max types.Float64 `tfsdk:"max"`
}

func (n *networkBandwidthGbpsData) expand(ctx context.Context) *ec2.NetworkBandwidthGbpsRequest {
	if n == nil {
		return nil
	}
	return &ec2.NetworkBandwidthGbpsRequest{
		Min: flex.Float64FromFramework(ctx, n.Min),
		Max: flex.Float64FromFramework(ctx, n.Max),
	}
}

func expandNetworkBandwidthGbpsData(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.NetworkBandwidthGbpsRequest {
	var options networkBandwidthGbpsData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx)
}

type networkInterfaceCountData struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}

func (n *networkInterfaceCountData) expand(ctx context.Context) *ec2.NetworkInterfaceCountRequest {
	if n == nil {
		return nil
	}
	return &ec2.NetworkInterfaceCountRequest{
		Min: flex.Int64FromFramework(ctx, n.Min),
		Max: flex.Int64FromFramework(ctx, n.Max),
	}
}

func expandNetworkInterfaceCountData(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.NetworkInterfaceCountRequest {
	var options networkInterfaceCountData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx)
}

type totalLocalStorageGBData struct {
	Min types.Float64 `tfsdk:"min"`
	Max types.Float64 `tfsdk:"max"`
}

func (t *totalLocalStorageGBData) expand(ctx context.Context) *ec2.TotalLocalStorageGBRequest {
	if t == nil {
		return nil
	}
	return &ec2.TotalLocalStorageGBRequest{
		Min: flex.Float64FromFramework(ctx, t.Min),
		Max: flex.Float64FromFramework(ctx, t.Max),
	}
}

func expandTotalLocalStorageGBData(ctx context.Context, object types.Object, diags *diag.Diagnostics) *ec2.TotalLocalStorageGBRequest {
	var options totalLocalStorageGBData
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return options.expand(ctx)
}
