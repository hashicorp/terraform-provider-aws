// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_instance_types_from_instance_requirements", name="Instance Types From Instance Requirements")
func newInstanceTypesFromInstanceRequirementsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &instanceTypesFromInstanceRequirementsDataSource{}

	return d, nil
}

const (
	DSNameInstanceTypesFromInstanceRequirements = "Instance Types From Instance Requirements Data Source"
)

type instanceTypesFromInstanceRequirementsDataSource struct {
	framework.DataSourceWithModel[instanceTypesFromInstanceRequirementsDataSourceModel]
}

func (d *instanceTypesFromInstanceRequirementsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"architecture_types": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						enum.FrameworkValidate[awstypes.ArchitectureType](),
					),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"instance_types": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"virtualization_types": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						enum.FrameworkValidate[awstypes.VirtualizationType](),
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"instance_requirements": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[instanceRequirementsData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"accelerator_manufacturers": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									enum.FrameworkValidate[awstypes.AcceleratorManufacturer](),
								),
							},
						},
						"accelerator_names": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									enum.FrameworkValidate[awstypes.AcceleratorName](),
								),
							},
						},
						"accelerator_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									enum.FrameworkValidate[awstypes.AcceleratorType](),
								),
							},
						},
						"allowed_instance_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
						"bare_metal": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.BareMetal](),
							},
						},
						"burstable_performance": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.BurstablePerformance](),
							},
						},
						"cpu_manufacturers": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									enum.FrameworkValidate[awstypes.CpuManufacturer](),
								),
							},
						},
						"excluded_instance_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
						"instance_generations": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									enum.FrameworkValidate[awstypes.InstanceGeneration](),
								),
							},
						},
						"local_storage": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.LocalStorage](),
							},
						},
						"local_storage_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									enum.FrameworkValidate[awstypes.LocalStorageType](),
								),
							},
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
						"memory_mib":                   minMaxInt64(ctx, false),
						"vcpu_count":                   minMaxInt64(ctx, false),
						"accelerator_count":            minMaxInt64(ctx, true),
						"accelerator_total_memory_mib": minMaxInt64(ctx, true),
						"baseline_ebs_bandwidth_mbps":  minMaxInt64(ctx, true),
						"baseline_performance_factors": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[baselinePerformanceFactorsData](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"cpu": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[cpuPerformanceFactorData](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"reference": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[performanceFactorReferenceData](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"instance_family": schema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"memory_gib_per_vcpu":     minMaxFloat64(ctx),
						"network_bandwidth_gbps":  minMaxFloat64(ctx),
						"network_interface_count": minMaxInt64(ctx, true),
						"total_local_storage_gb":  minMaxFloat64(ctx),
					},
				},
			},
		},
	}
}

func (d *instanceTypesFromInstanceRequirementsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data instanceTypesFromInstanceRequirementsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	var input ec2.GetInstanceTypesFromInstanceRequirementsInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findInstanceTypesFromInstanceRequirements(ctx, conn, &input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, DSNameInstanceTypesFromInstanceRequirements, d.Meta().Region(ctx), err),
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(d.Meta().Region(ctx))
	data.InstanceTypes = fwflex.FlattenFrameworkStringValueListOfString(ctx, tfslices.ApplyToAll(output, func(v awstypes.InstanceTypeInfoFromInstanceRequirements) string {
		return aws.ToString(v.InstanceType)
	}))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type instanceTypesFromInstanceRequirementsDataSourceModel struct {
	framework.WithRegionModel
	ArchitectureTypes    fwtypes.ListOfString                                      `tfsdk:"architecture_types"`
	InstanceRequirements fwtypes.ListNestedObjectValueOf[instanceRequirementsData] `tfsdk:"instance_requirements"`
	InstanceTypes        fwtypes.ListOfString                                      `tfsdk:"instance_types"`
	ID                   types.String                                              `tfsdk:"id"`
	VirtualizationTypes  fwtypes.ListOfString                                      `tfsdk:"virtualization_types"`
}

type instanceRequirementsData struct {
	MemoryMiB                                 fwtypes.ListNestedObjectValueOf[minMax[types.Int64]]            `tfsdk:"memory_mib"`
	VCpuCount                                 fwtypes.ListNestedObjectValueOf[minMax[types.Int64]]            `tfsdk:"vcpu_count"`
	AcceleratorCount                          fwtypes.ListNestedObjectValueOf[minMax[types.Int64]]            `tfsdk:"accelerator_count"`
	AcceleratorManufacturers                  fwtypes.ListOfString                                            `tfsdk:"accelerator_manufacturers"`
	AcceleratorNames                          fwtypes.ListOfString                                            `tfsdk:"accelerator_names"`
	AcceleratorTotalMemoryMiB                 fwtypes.ListNestedObjectValueOf[minMax[types.Int64]]            `tfsdk:"accelerator_total_memory_mib"`
	AcceleratorTypes                          fwtypes.ListOfString                                            `tfsdk:"accelerator_types"`
	AllowedInstanceTypes                      fwtypes.ListOfString                                            `tfsdk:"allowed_instance_types"`
	BareMetal                                 types.String                                                    `tfsdk:"bare_metal"`
	BaselineEbsBandwidthMbps                  fwtypes.ListNestedObjectValueOf[minMax[types.Int64]]            `tfsdk:"baseline_ebs_bandwidth_mbps"`
	BaselinePerformanceFactors                fwtypes.ListNestedObjectValueOf[baselinePerformanceFactorsData] `tfsdk:"baseline_performance_factors"`
	BurstablePerformance                      types.String                                                    `tfsdk:"burstable_performance"`
	CpuManufacturers                          fwtypes.ListOfString                                            `tfsdk:"cpu_manufacturers"`
	ExcludedInstanceTypes                     fwtypes.ListOfString                                            `tfsdk:"excluded_instance_types"`
	InstanceGenerations                       fwtypes.ListOfString                                            `tfsdk:"instance_generations"`
	LocalStorage                              types.String                                                    `tfsdk:"local_storage"`
	LocalStorageTypes                         fwtypes.ListOfString                                            `tfsdk:"local_storage_types"`
	MemoryGiBPerVCpu                          fwtypes.ListNestedObjectValueOf[minMax[types.Float64]]          `tfsdk:"memory_gib_per_vcpu"`
	NetworkBandwidthGbps                      fwtypes.ListNestedObjectValueOf[minMax[types.Float64]]          `tfsdk:"network_bandwidth_gbps"`
	NetworkInterfaceCount                     fwtypes.ListNestedObjectValueOf[minMax[types.Int64]]            `tfsdk:"network_interface_count"`
	OnDemandMaxPricePercentageOverLowestPrice types.Int64                                                     `tfsdk:"on_demand_max_price_percentage_over_lowest_price"`
	RequireHibernateSupport                   types.Bool                                                      `tfsdk:"require_hibernate_support"`
	SpotMaxPricePercentageOverLowestPrice     types.Int64                                                     `tfsdk:"spot_max_price_percentage_over_lowest_price"`
	TotalLocalStorageGB                       fwtypes.ListNestedObjectValueOf[minMax[types.Float64]]          `tfsdk:"total_local_storage_gb"`
}

type baselinePerformanceFactorsData struct {
	Cpu fwtypes.ListNestedObjectValueOf[cpuPerformanceFactorData] `tfsdk:"cpu"`
}

type cpuPerformanceFactorData struct {
	References fwtypes.ListNestedObjectValueOf[performanceFactorReferenceData] `tfsdk:"reference"`
}

type performanceFactorReferenceData struct {
	InstanceFamily types.String `tfsdk:"instance_family"`
}

type minMax[T comparable] struct {
	Min T `tfsdk:"min"`
	Max T `tfsdk:"max"`
}

func minMaxInt64(ctx context.Context, optional bool) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[minMax[types.Int64]](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrMin: schema.Int64Attribute{
					Optional: optional,
					Required: !optional,
				},
				names.AttrMax: schema.Int64Attribute{
					Optional: true,
				},
			},
		},
	}
}

func minMaxFloat64(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[minMax[types.Float64]](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrMin: schema.Float64Attribute{
					Optional: true,
				},
				names.AttrMax: schema.Float64Attribute{
					Optional: true,
				},
			},
		},
	}
}
