package sagemaker

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func newResourceCluster(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCluster{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameCluster = "Cluster"
)

type resourceCluster struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceCluster) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	var vpcConfigModelAttributeSchema = map[string]schema.Attribute{
		"security_group_ids": schema.SetAttribute{
			CustomType: fwtypes.SetOfStringType,
			Required:   true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
			Validators: []validator.Set{
				setvalidator.SizeBetween(1, 5),
				setvalidator.ValueStringsAre(
					stringvalidator.RegexMatches(regexache.MustCompile(`[-0-9a-zA-Z]+`), `must match [-0-9a-zA-Z]+`),
					stringvalidator.LengthAtMost(32),
				),
			},
		},
		"subnets": schema.SetAttribute{
			CustomType: fwtypes.SetOfStringType,
			Required:   true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
			Validators: []validator.Set{
				setvalidator.SizeBetween(1, 16),
				setvalidator.ValueStringsAre(
					stringvalidator.RegexMatches(regexache.MustCompile(`[-0-9a-zA-Z]+`), `must match [-0-9a-zA-Z]+`),
					stringvalidator.LengthAtMost(32),
				),
			},
		},
	}
	var capacitySizeConfigModelAttributeSchema = map[string]schema.Attribute{
		"type": schema.StringAttribute{
			CustomType: fwtypes.StringEnumType[awstypes.NodeUnavailabilityType](),
			Required:   true,
		},
		"value": schema.Int32Attribute{
			Required: true,
			Validators: []validator.Int32{
				int32validator.AtLeast(1),
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"cluster_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "must match ^[a-zA-Z0-9](-*[a-zA-Z0-9])*$"),
				},
			},
			"cluster_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ClusterStatus](),
				Computed:   true,
			},
			"node_recovery": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ClusterNodeRecovery](),
				Optional:   true,
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"instance_group": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[clusterInstanceGroupSpecificationModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeBetween(1, 100),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"execution_role": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(20, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws[a-z\-]*:iam::\d{12}:role/?[a-zA-Z_0-9+=,.@\-_/]+$`), `must match ^arn:aws[a-z\-]*:iam::\d{12}:role/?[a-zA-Z_0-9+=,.@\-_/]+$`),
							},
						},
						"instance_count": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.Between(0, 6758),
							},
						},
						"instance_group_name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 63),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "must match ^[a-zA-Z0-9](-*[a-zA-Z0-9])*$"),
							},
						},
						"instance_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ClusterInstanceType](),
							Required:   true,
						},
						"nodes": schema.ListAttribute{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[clusterNodeDetailsModel](ctx),
							Computed:    true,
							ElementType: fwtypes.NewObjectTypeOf[clusterNodeDetailsModel](ctx),
						},
						"on_start_deep_healthchecks": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringEnumType[awstypes.DeepHealthCheckType](),
							Optional:   true,
							Validators: []validator.List{
								listvalidator.AlsoRequires(
									path.MatchRoot("orchestrator").AtAnyListIndex().AtName("eks"),
								),
							},
						},
						"threads_per_core": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Int32{
								int32validator.OneOf(1, 2),
							},
						},
						"training_plan_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(20, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:training-plan/.*`), `must match arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:training-plan/.*`),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"instance_storage_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[clusterInstanceStorageConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"ebs_volume_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[clusterEbsVolumeConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"volume_size_in_gb": schema.Int32Attribute{
													Required: true,
													Validators: []validator.Int32{
														int32validator.Between(1, 16384),
													},
												},
											},
										},
									},
								},
							},
						},
						"lifecycle_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[clusterLifeCycleConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"on_create": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[\S\s]+$`), `must match ^[\S\s]+$`),
										},
									},
									"source_s3_uri": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(1024),
											stringvalidator.RegexMatches(regexache.MustCompile(`^(https|s3)://([^/]+)/?(.*)$`), `must match ^(https|s3)://([^/]+)/?(.*)$`),
										},
									},
								},
							},
						},
						"override_vpc_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: vpcConfigModelAttributeSchema,
							},
						},
						"scheduled_update_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[scheduledUpdateConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"schedule_expression": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 256),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"deployment_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[deploymentConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"wait_interval": schema.Int32Attribute{
													Optional: true,
													Validators: []validator.Int32{
														int32validator.Between(0, 3600),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"auto_rollback_configuration": schema.SetNestedBlock{
													CustomType: fwtypes.NewSetNestedObjectTypeOf[alarmDetailsModel](ctx),
													Validators: []validator.Set{
														setvalidator.SizeBetween(1, 10),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"alarm_name": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 255),
																	stringvalidator.RegexMatches(regexache.MustCompile(`\S`), `must not contain whitespace`),
																},
															},
														},
													},
												},
												"rolling_update_policy": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[rollingDeploymentPolicyModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"maximum_batch_size": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[capacitySizeConfigModel](ctx),
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: capacitySizeConfigModelAttributeSchema,
																},
															},
															"rollback_maximum_batch_size": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[capacitySizeConfigModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: capacitySizeConfigModelAttributeSchema,
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
						},
					},
				},
			},
			"orchestrator": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[clusterOrchestratorModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"eks": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[clusterOrchestratorEksConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"cluster_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(20, 2048),
											stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws[a-z\-]*:eks:[a-z0-9\-]*:[0-9]{12}:cluster\/[0-9A-Za-z][A-Za-z0-9\-_]{0,99}$`), `must match ^arn:aws[a-z\-]*:eks:[a-z0-9\-]*:[0-9]{12}:cluster\/[0-9A-Za-z][A-Za-z0-9\-_]{0,99}$`),
										},
									},
								},
							},
						},
					},
				},
			},
			"vpc_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: vpcConfigModelAttributeSchema,
				},
			},
		},
	}
}

func (r *resourceCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *resourceCluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *resourceCluster) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceCluster) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *resourceCluster) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}

func waitClusterCreated(ctx context.Context, conn *sagemaker.Client, name string, timeout time.Duration) (*sagemaker.DescribeClusterOutput, error) {
	return nil, nil
}

func waitClusterUpdated(ctx context.Context, conn *sagemaker.Client, name string, timeout time.Duration) (*sagemaker.DescribeClusterOutput, error) {
	return nil, nil
}

func waitClusterDeleted(ctx context.Context, conn *sagemaker.Client, name string, timeout time.Duration) (*sagemaker.DescribeClusterOutput, error) {
	return nil, nil
}

func statusCluster(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return nil
}

func findClusterByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeClusterOutput, error) {
	return nil, nil
}

type resourceClusterModel struct {
	Arn            types.String                                                           `tfsdk:"arn"`
	ClusterName    types.String                                                           `tfsdk:"cluster_name"`
	ClusterStatus  fwtypes.StringEnum[awstypes.ClusterStatus]                             `tfsdk:"cluster_status"`
	InstanceGroups fwtypes.SetNestedObjectValueOf[clusterInstanceGroupSpecificationModel] `tfsdk:"instance_group"`
	NodeRecovery   fwtypes.StringEnum[awstypes.ClusterNodeRecovery]                       `tfsdk:"node_recovery"`
	Orchestrator   fwtypes.ListNestedObjectValueOf[clusterOrchestratorModel]              `tfsdk:"orchestrator"`
	Tags           tftags.Map                                                             `tfsdk:"tags"`
	TagsAll        tftags.Map                                                             `tfsdk:"tags_all"`
	Timeouts       timeouts.Value                                                         `tfsdk:"timeouts"`
	VpcConfig      fwtypes.ListNestedObjectValueOf[vpcConfigModel]                        `tfsdk:"vpc_config"`
}

type clusterInstanceGroupSpecificationModel struct {
	ExecutionRole           types.String                                                       `tfsdk:"execution_role"`
	InstanceCount           types.Int32                                                        `tfsdk:"instance_count"`
	InstanceGroupName       types.String                                                       `tfsdk:"instance_group_name"`
	InstanceType            fwtypes.StringEnum[awstypes.ClusterInstanceType]                   `tfsdk:"instance_type"`
	LifeCycleConfig         fwtypes.ListNestedObjectValueOf[clusterLifeCycleConfigModel]       `tfsdk:"lifecycle_config"`
	InstanceStorageConfigs  fwtypes.ListNestedObjectValueOf[clusterInstanceStorageConfigModel] `tfsdk:"instance_storage_config"`
	Nodes                   fwtypes.ListNestedObjectValueOf[clusterNodeDetailsModel]           `tfsdk:"nodes"`
	OnStartDeepHealthChecks fwtypes.ListValueOf[types.String]                                  `tfsdk:"on_start_deep_healthchecks"`
	OverrideVpcConfig       fwtypes.ListNestedObjectValueOf[vpcConfigModel]                    `tfsdk:"override_vpc_config"`
	ScheduledUpdateConfig   fwtypes.ListNestedObjectValueOf[scheduledUpdateConfigModel]        `tfsdk:"scheduled_update_config"`
	ThreadsPerCore          types.Int32                                                        `tfsdk:"threads_per_core"`
	TrainingPlanArn         fwtypes.ARN                                                        `tfsdk:"training_plan_arn"`
}

type clusterLifeCycleConfigModel struct {
	OnCreate    types.String `tfsdk:"on_create"`
	SourceS3Uri types.String `tfsdk:"source_s3_uri"`
}

type clusterInstanceStorageConfigModel struct {
	EbsVolumeConfig fwtypes.ListNestedObjectValueOf[clusterEbsVolumeConfigModel] `tfsdk:"ebs_volume_config"`
}

type clusterEbsVolumeConfigModel struct {
	VolumeSizeInGB types.Int32 `tfsdk:"volume_size_in_gb"`
}

type clusterNodeDetailsModel struct {
	InstanceGroupName      types.String                                                       `tfsdk:"instance_group_name"`
	InstanceID             types.String                                                       `tfsdk:"instance_id"`
	InstanceStatus         fwtypes.ListNestedObjectValueOf[clusterInstanceStatusDetailsModel] `tfsdk:"instance_status"`
	InstanceStorageConfigs fwtypes.ListNestedObjectValueOf[clusterInstanceStorageConfigModel] `tfsdk:"instance_storage_configs"`
	InstanceType           fwtypes.StringEnum[awstypes.ClusterInstanceType]                   `tfsdk:"instance_type"`
	LastSoftwareUpdateTime timetypes.RFC3339                                                  `tfsdk:"last_software_update_time"`
	LaunchTime             timetypes.RFC3339                                                  `tfsdk:"launch_time"`
	LifeCycleConfig        fwtypes.ListNestedObjectValueOf[clusterLifeCycleConfigModel]       `tfsdk:"lifecycle_config"`
	OverrideVpcConfig      fwtypes.ListNestedObjectValueOf[vpcConfigModel]                    `tfsdk:"override_vpc_config"`
	Placement              fwtypes.ListNestedObjectValueOf[clusterInstancePlacementModel]     `tfsdk:"placement"`
	PrivateDnsHostname     types.String                                                       `tfsdk:"private_dns_hostname"`
	PrivatePrimaryIP       types.String                                                       `tfsdk:"private_primary_ip"`
	PrivatePrimaryIpv6     types.String                                                       `tfsdk:"private_primary_ipv6"`
	ThreadsPerCore         types.Int64                                                        `tfsdk:"threads_per_core"`
}

type clusterInstanceStatusDetailsModel struct {
	Message types.String                                       `tfsdk:"message"`
	Status  fwtypes.StringEnum[awstypes.ClusterInstanceStatus] `tfsdk:"status"`
}

type clusterInstancePlacementModel struct {
	AvailabilityZone   types.String `tfsdk:"availability_zone"`
	AvailabilityZoneID types.String `tfsdk:"availability_zone_id"`
}

type scheduledUpdateConfigModel struct {
	ScheduleExpression types.String                                                  `tfsdk:"schedule_expression"`
	DeploymentConfig   fwtypes.ListNestedObjectValueOf[deploymentConfigurationModel] `tfsdk:"deployment_config"`
}

type deploymentConfigurationModel struct {
	AutoRollbackConfiguration fwtypes.SetNestedObjectValueOf[alarmDetailsModel]             `tfsdk:"auto_rollback_configuration"`
	RollingUpdatePolicy       fwtypes.ListNestedObjectValueOf[rollingDeploymentPolicyModel] `tfsdk:"rolling_update_policy"`
	WaitInterval              types.Int32                                                   `tfsdk:"wait_interval"`
}

type alarmDetailsModel struct {
	AlarmName types.String `tfsdk:"alarm_name"`
}

type rollingDeploymentPolicyModel struct {
	MaximumBatchSize         fwtypes.ListNestedObjectValueOf[capacitySizeConfigModel] `tfsdk:"maximum_batch_size"`
	RollbackMaximumBatchSize fwtypes.ListNestedObjectValueOf[capacitySizeConfigModel] `tfsdk:"rollback_maximum_batch_size"`
}

type capacitySizeConfigModel struct {
	Type  fwtypes.StringEnum[awstypes.NodeUnavailabilityType] `tfsdk:"type"`
	Value types.Int32                                         `tfsdk:"value"`
}

type clusterOrchestratorModel struct {
	Eks fwtypes.ListNestedObjectValueOf[clusterOrchestratorEksConfigModel] `tfsdk:"eks"`
}

type clusterOrchestratorEksConfigModel struct {
	ClusterArn fwtypes.ARN `tfsdk:"cluster_arn"`
}

type vpcConfigModel struct {
	SecurityGroupIds fwtypes.SetOfString `tfsdk:"security_group_ids"`
	Subnets          fwtypes.SetOfString `tfsdk:"subnets"`
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	return nil, nil
}
