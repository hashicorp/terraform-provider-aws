// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_batch_job_definition", name="Job Definition")
// @Tags
// @Testing(tagsIdentifierAttribute="arn")
func newJobDefinitionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &jobDefinitionDataSource{}, nil
}

type jobDefinitionDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *jobDefinitionDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_batch_job_definition"
}

func (d *jobDefinitionDataSource) SchemaEKSContainer(ctx context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"args": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"command": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"image": schema.StringAttribute{
				Computed: true,
			},
			"image_pull_policy": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"env": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[keyValuePairModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
						names.AttrValue: schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			names.AttrResources: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eksContainerResourceRequirementsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"limits": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
						"requests": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"security_context": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eksContainerSecurityContextModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"privileged": schema.BoolAttribute{
							Computed: true,
						},
						"run_as_user": schema.Int64Attribute{
							Computed: true,
						},
						"read_only_root_filesystem": schema.BoolAttribute{
							Computed: true,
						},
						"run_as_non_root": schema.BoolAttribute{
							Computed: true,
						},
						"run_as_group": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
			"volume_mounts": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eksContainerVolumeMountModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"mount_path": schema.StringAttribute{
							Computed: true,
						},
						"read_only": schema.BoolAttribute{
							Computed: true,
						},
						"sub_path": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *jobDefinitionDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.ARNType,
			},
			"arn_prefix": schema.StringAttribute{
				Computed: true,
			},
			"container_orchestration_type": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"memory": schema.Int32Attribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
			},
			"node_properties": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[nodePropertiesModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"main_node":             types.Int64Type,
						"node_range_properties": fwtypes.NewListNestedObjectTypeOf[nodeRangePropertyModel](ctx),
						"num_nodes":             types.Int64Type,
					},
				},
			},
			"retry_strategy": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[retryStrategyModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attempts":         types.Int32Type,
						"evaluate_on_exit": fwtypes.NewListNestedObjectTypeOf[evaluateOnExitModel](ctx),
					},
				},
			},
			"repository_credentials": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[repositoryCredentialsModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"credentials_parameter": types.StringType,
					},
				},
			},
			"revision": schema.Int64Attribute{
				Computed: true,
				Optional: true,
			},
			"scheduling_priority": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(jobDefinitionStatus_Values()...),
				},
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			names.AttrTimeout: schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[jobTimeoutModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attempt_duration_seconds": types.Int64Type,
					},
				},
			},
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
			"vcpus": schema.Int32Attribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"eks_properties": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eksPropertiesModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"pod_properties": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[eksPodPropertiesModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"dns_policy": schema.StringAttribute{
										Computed: true,
									},
									"host_network": schema.BoolAttribute{
										Computed: true,
									},
									"service_account_name": schema.StringAttribute{
										Computed: true,
									},
									"share_process_namespace": schema.BoolAttribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"containers": schema.ListNestedBlock{
										CustomType:   fwtypes.NewListNestedObjectTypeOf[eksContainerModel](ctx),
										NestedObject: d.SchemaEKSContainer(ctx),
									},
									"image_pull_secrets": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[eksImagePullSecrets](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrName: schema.StringAttribute{
													Computed: true,
												},
											},
										},
									},
									"init_containers": schema.ListNestedBlock{
										CustomType:   fwtypes.NewListNestedObjectTypeOf[eksContainerModel](ctx),
										NestedObject: d.SchemaEKSContainer(ctx),
									},
									"metadata": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[eksMetadataModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"labels": schema.MapAttribute{
													Computed: true,

													ElementType: types.StringType,
												},
											},
										},
									},
									"volumes": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[eksVolumeModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrName: schema.StringAttribute{
													Computed: true,
												},
											},
											Blocks: map[string]schema.Block{
												"empty_dir": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[eksEmptyDirModel](ctx),
												},
												"host_path": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[eksHostPathModel](ctx),
												},
												"secret": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[eksSecretModel](ctx),
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
	}
}

func (d *jobDefinitionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data jobDefinitionDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BatchClient(ctx)

	var jd *awstypes.JobDefinition

	if !data.JobDefinitionARN.IsNull() {
		arn := data.JobDefinitionARN.ValueString()
		input := &batch.DescribeJobDefinitionsInput{
			JobDefinitions: []string{arn},
		}

		output, err := findJobDefinition(ctx, conn, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading Batch Job Definition (%s)", arn), err.Error())

			return
		}

		jd = output
	} else if !data.JobDefinitionName.IsNull() {
		name := data.JobDefinitionName.ValueString()
		status := jobDefinitionStatusActive
		if !data.Status.IsNull() {
			status = data.Status.ValueString()
		}
		input := &batch.DescribeJobDefinitionsInput{
			JobDefinitionName: aws.String(name),
			Status:            aws.String(status),
		}

		output, err := findJobDefinitions(ctx, conn, input)

		if len(output) == 0 {
			err = tfresource.NewEmptyResultError(input)
		}

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading Batch Job Definitions (%s/%s)", name, status), err.Error())

			return
		}

		if data.Revision.IsNull() {
			// Sort in descending revision order.
			slices.SortFunc(output, func(a, b awstypes.JobDefinition) int {
				return int(aws.ToInt32(b.Revision) - aws.ToInt32(a.Revision))
			})

			jd = &output[0]
		} else {
			revision := int32(data.Revision.ValueInt64())
			i := slices.IndexFunc(output, func(v awstypes.JobDefinition) bool {
				return aws.ToInt32(v.Revision) == revision
			})

			if i == -1 {
				response.Diagnostics.AddError(fmt.Sprintf("reading Batch Job Definition (%s/%s) revision (%d)", name, status, revision), tfresource.NewEmptyResultError(input).Error())

				return
			}

			jd = &output[i]
		}
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, jd, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	arnPrefix := strings.TrimSuffix(aws.ToString(jd.JobDefinitionArn), fmt.Sprintf(":%d", aws.ToInt32(jd.Revision)))
	data.ARNPrefix = types.StringValue(arnPrefix)

	setTagsOut(ctx, jd.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (d *jobDefinitionDataSource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrARN),
			path.MatchRoot(names.AttrName),
		),
	}
}

type jobDefinitionDataSourceModel struct {
	ARNPrefix                  types.String                                                `tfsdk:"arn_prefix"`
	ContainerOrchestrationType types.String                                                `tfsdk:"container_orchestration_type"`
	EKSProperties              fwtypes.ListNestedObjectValueOf[eksPropertiesModel]         `tfsdk:"eks_properties"`
	ID                         types.String                                                `tfsdk:"id"`
	JobDefinitionARN           fwtypes.ARN                                                 `tfsdk:"arn"`
	JobDefinitionName          types.String                                                `tfsdk:"name"`
	Memory                     types.Int32                                                 `tfsdk:"memory"`
	NodeProperties             fwtypes.ListNestedObjectValueOf[nodePropertiesModel]        `tfsdk:"node_properties"`
	RetryStrategy              fwtypes.ListNestedObjectValueOf[retryStrategyModel]         `tfsdk:"retry_strategy"`
	RepositoryCredentials      fwtypes.ListNestedObjectValueOf[repositoryCredentialsModel] `tfsdk:"repository_credentials"`
	Revision                   types.Int64                                                 `tfsdk:"revision"`
	SchedulingPriority         types.Int64                                                 `tfsdk:"scheduling_priority"`
	Status                     types.String                                                `tfsdk:"status"`
	Tags                       tftags.Map                                                  `tfsdk:"tags"`
	Timeout                    fwtypes.ListNestedObjectValueOf[jobTimeoutModel]            `tfsdk:"timeout"`
	Type                       types.String                                                `tfsdk:"type"`
	VCPUs                      types.Int32                                                 `tfsdk:"vcpus"`
}

type eksPropertiesModel struct {
	PodProperties fwtypes.ListNestedObjectValueOf[eksPodPropertiesModel] `tfsdk:"pod_properties"`
}

type eksPodPropertiesModel struct {
	Containers            fwtypes.ListNestedObjectValueOf[eksContainerModel]   `tfsdk:"containers"`
	DNSPolicy             types.String                                         `tfsdk:"dns_policy"`
	HostNetwork           types.Bool                                           `tfsdk:"host_network"`
	ImagePullSecrets      fwtypes.ListNestedObjectValueOf[eksImagePullSecrets] `tfsdk:"image_pull_secrets"`
	InitContainers        fwtypes.ListNestedObjectValueOf[eksContainerModel]   `tfsdk:"init_containers"`
	Metadata              fwtypes.ListNestedObjectValueOf[eksMetadataModel]    `tfsdk:"metadata"`
	ServiceAccountName    types.String                                         `tfsdk:"service_account_name"`
	ShareProcessNamespace types.Bool                                           `tfsdk:"share_process_namespace"`
	Volumes               fwtypes.ListNestedObjectValueOf[eksVolumeModel]      `tfsdk:"volumes"`
}

type eksImagePullSecrets struct {
	Name types.String `tfsdk:"name"`
}

type eksContainerModel struct {
	Args            fwtypes.ListValueOf[types.String]                                      `tfsdk:"args"`
	Command         fwtypes.ListValueOf[types.String]                                      `tfsdk:"command"`
	Env             fwtypes.ListNestedObjectValueOf[eksContainerEnvironmentVariableModel]  `tfsdk:"env"`
	Image           types.String                                                           `tfsdk:"image"`
	ImagePullPolicy types.String                                                           `tfsdk:"image_pull_policy"`
	Name            types.String                                                           `tfsdk:"name"`
	Resources       fwtypes.ListNestedObjectValueOf[eksContainerResourceRequirementsModel] `tfsdk:"resources"`
	SecurityContext fwtypes.ListNestedObjectValueOf[eksContainerSecurityContextModel]      `tfsdk:"security_context"`
	VolumeMounts    fwtypes.ListNestedObjectValueOf[eksContainerVolumeMountModel]          `tfsdk:"volume_mounts"`
}

type eksContainerEnvironmentVariableModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type eksContainerResourceRequirementsModel struct {
	Limits   fwtypes.MapValueOf[types.String] `tfsdk:"limits"`
	Requests fwtypes.MapValueOf[types.String] `tfsdk:"requests"`
}

type eksContainerSecurityContextModel struct {
	Privileged             types.Bool  `tfsdk:"privileged"`
	ReadOnlyRootFilesystem types.Bool  `tfsdk:"read_only_root_file_system"`
	RunAsGroup             types.Int64 `tfsdk:"run_as_group"`
	RunAsNonRoot           types.Bool  `tfsdk:"run_as_non_root"`
	RunAsUser              types.Int64 `tfsdk:"run_as_user"`
}

type eksContainerVolumeMountModel struct {
	MountPath types.String `tfsdk:"mount_path"`
	Name      types.String `tfsdk:"name"`
	ReadOnly  types.Bool   `tfsdk:"read_only"`
}

type eksMetadataModel struct {
	Labels fwtypes.MapValueOf[types.String] `tfsdk:"labels"`
}

type eksVolumeModel struct {
	EmptyDir fwtypes.ListNestedObjectValueOf[eksEmptyDirModel] `tfsdk:"empty_dir"`
	HostPath fwtypes.ListNestedObjectValueOf[eksHostPathModel] `tfsdk:"host_path"`
	Name     types.String                                      `tfsdk:"name"`
	Secret   fwtypes.ListNestedObjectValueOf[eksSecretModel]   `tfsdk:"secret"`
}

type eksEmptyDirModel struct {
	Medium    types.String `tfsdk:"medium"`
	SizeLimit types.String `tfsdk:"size_limit"`
}

type eksHostPathModel struct {
	Path types.String `tfsdk:"path"`
}

type eksSecretModel struct {
	Optional   types.Bool   `tfsdk:"optional"`
	SecretName types.String `tfsdk:"secret_name"`
}

type nodePropertiesModel struct {
	MainNode            types.Int64                                             `tfsdk:"main_node"`
	NodeRangeProperties fwtypes.ListNestedObjectValueOf[nodeRangePropertyModel] `tfsdk:"node_range_properties"`
	NumNodes            types.Int64                                             `tfsdk:"num_nodes"`
}

type nodeRangePropertyModel struct {
	Container     fwtypes.ListNestedObjectValueOf[containerPropertiesModel] `tfsdk:"container"`
	EKSProperties fwtypes.ListNestedObjectValueOf[eksPropertiesModel]       `tfsdk:"eks_properties"`
	ECSProperties fwtypes.ListNestedObjectValueOf[ecsPropertiesModel]       `tfsdk:"ecs_properties"`
	TargetNodes   types.String                                              `tfsdk:"target_nodes"`
	InstanceTypes fwtypes.ListValueOf[types.String]                         `tfsdk:"instance_types"`
}

type containerPropertiesModel struct {
	Command                      fwtypes.ListValueOf[types.String]                                  `tfsdk:"command"`
	Environment                  fwtypes.ListNestedObjectValueOf[keyValuePairModel]                 `tfsdk:"environment"`
	EphemeralStorage             fwtypes.ListNestedObjectValueOf[ephemeralStorageModel]             `tfsdk:"ephemeral_storage"`
	ExecutionRoleARN             types.String                                                       `tfsdk:"execution_role_arn"`
	FargatePlatformConfiguration fwtypes.ListNestedObjectValueOf[fargatePlatformConfigurationModel] `tfsdk:"fargate_platform_configuration"`
	Memory                       types.Int32                                                        `tfsdk:"memory"`
	Image                        types.String                                                       `tfsdk:"image"`
	InstanceType                 types.String                                                       `tfsdk:"instance_type"`
	JobRoleARN                   types.String                                                       `tfsdk:"job_role_arn"`
	LinuxParameters              fwtypes.ListNestedObjectValueOf[linuxParametersModel]              `tfsdk:"linux_parameters"`
	LogConfiguration             fwtypes.ListNestedObjectValueOf[logConfigurationModel]             `tfsdk:"log_configuration"`
	MountPoints                  fwtypes.ListNestedObjectValueOf[mountPointModel]                   `tfsdk:"mount_points"`
	NetworkConfiguration         fwtypes.ListNestedObjectValueOf[networkConfigurationModel]         `tfsdk:"network_configuration"`
	Privileged                   types.Bool                                                         `tfsdk:"privileged"`
	ReadonlyRootFilesystem       types.Bool                                                         `tfsdk:"readonly_root_filesystem"`
	RepositoryCredential         fwtypes.ListNestedObjectValueOf[repositoryCredentialsModel]        `tfsdk:"repository_credentials"`
	ResourceRequirements         fwtypes.ListNestedObjectValueOf[resourceRequirementModel]          `tfsdk:"resource_requirements"`
	RuntimePlatform              fwtypes.ListNestedObjectValueOf[runtimePlatformModel]              `tfsdk:"runtime_platform"`
	Secrets                      fwtypes.ListNestedObjectValueOf[secretModel]                       `tfsdk:"secrets"`
	Ulimits                      fwtypes.ListNestedObjectValueOf[ulimitModel]                       `tfsdk:"ulimits"`
	User                         types.String                                                       `tfsdk:"user"`
	Vcpus                        types.Int32                                                        `tfsdk:"vcpus"`
	Volumes                      fwtypes.ListNestedObjectValueOf[volumeModel]                       `tfsdk:"volumes"`
}

type taskPropertiesContainerModel struct {
	Image                  types.String                                                  `tfsdk:"image"`
	Command                fwtypes.ListValueOf[types.String]                             `tfsdk:"command"`
	DependsOn              fwtypes.ListNestedObjectValueOf[taskContainerDependencyModel] `tfsdk:"depends_on"`
	Environment            fwtypes.ListNestedObjectValueOf[keyValuePairModel]            `tfsdk:"environment"`
	Essential              types.Bool                                                    `tfsdk:"essential"`
	LinuxParameters        fwtypes.ListNestedObjectValueOf[linuxParametersModel]         `tfsdk:"linux_parameters"`
	LogConfiguration       fwtypes.ListNestedObjectValueOf[logConfigurationModel]        `tfsdk:"log_configuration"`
	MountPoints            fwtypes.ListNestedObjectValueOf[mountPointModel]              `tfsdk:"mount_points"`
	Name                   types.String                                                  `tfsdk:"name"`
	Privileged             types.Bool                                                    `tfsdk:"privileged"`
	ReadonlyRootFilesystem types.Bool                                                    `tfsdk:"readonly_root_filesystem"`
	ResourceRequirements   fwtypes.ListNestedObjectValueOf[resourceRequirementModel]     `tfsdk:"resource_requirements"`
	RepositoryCredentials  fwtypes.ListNestedObjectValueOf[repositoryCredentialsModel]   `tfsdk:"repository_credentials"`
	Secrets                fwtypes.ListNestedObjectValueOf[secretModel]                  `tfsdk:"secrets"`
	Ulimits                fwtypes.ListNestedObjectValueOf[ulimitModel]                  `tfsdk:"ulimits"`
	User                   types.String                                                  `tfsdk:"user"`
}

type taskContainerDependencyModel struct {
	Condition     types.String `tfsdk:"condition"`
	ContainerName types.String `tfsdk:"container_name"`
}

type keyValuePairModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type ephemeralStorageModel struct {
	SizeInGiB types.Int64 `tfsdk:"size_in_gib"`
}

type fargatePlatformConfigurationModel struct {
	PlatformVersion types.String `tfsdk:"platform_version"`
}

type linuxParametersModel struct {
	Devices            fwtypes.ListNestedObjectValueOf[deviceModel] `tfsdk:"devices"`
	InitProcessEnabled types.Bool                                   `tfsdk:"init_process_enabled"`
	MaxSwap            types.Int64                                  `tfsdk:"max_swap"`
	SharedMemorySize   types.Int64                                  `tfsdk:"shared_memory_size"`
	Swappiness         types.Int64                                  `tfsdk:"swappiness"`
	Tmpfs              fwtypes.ListNestedObjectValueOf[tmpfsModel]  `tfsdk:"tmpfs"`
}

type deviceModel struct {
	ContainerPath types.String                      `tfsdk:"container_path"`
	HostPath      types.String                      `tfsdk:"host_path"`
	Permissions   fwtypes.ListValueOf[types.String] `tfsdk:"permissions"`
}

type tmpfsModel struct {
	ContainerPath types.String                      `tfsdk:"container_path"`
	MountOptions  fwtypes.ListValueOf[types.String] `tfsdk:"mount_options"`
	Size          types.Int64                       `tfsdk:"size"`
}

type logConfigurationModel struct {
	LogDriver     types.String                                 `tfsdk:"log_driver"`
	Options       fwtypes.MapValueOf[types.String]             `tfsdk:"options"`
	SecretOptions fwtypes.ListNestedObjectValueOf[secretModel] `tfsdk:"secret_options"`
}

type secretModel struct {
	Name      types.String `tfsdk:"name"`
	ValueFrom types.String `tfsdk:"value_from"`
}

type mountPointModel struct {
	ContainerPath types.String `tfsdk:"container_path"`
	ReadOnly      types.Bool   `tfsdk:"read_only"`
	SourceVolume  types.String `tfsdk:"source_volume"`
}

type networkConfigurationModel struct {
	AssignPublicIP types.String `tfsdk:"assign_public_ip"`
}

type resourceRequirementModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type runtimePlatformModel struct {
	CPUArchitecture       types.String `tfsdk:"cpu_architecture"`
	OperatingSystemFamily types.String `tfsdk:"operating_system_family"`
}

type ulimitModel struct {
	HardLimit types.Int64  `tfsdk:"hard_limit"`
	Name      types.String `tfsdk:"name"`
	SoftLimit types.Int64  `tfsdk:"soft_limit"`
}

type volumeModel struct {
	EFSVolumeConfiguration fwtypes.ListNestedObjectValueOf[efsVolumeConfigurationModel] `tfsdk:"efs_volume_configuration"`
	Host                   fwtypes.ListNestedObjectValueOf[hostModel]                   `tfsdk:"host"`
	Name                   types.String                                                 `tfsdk:"name"`
}

type efsVolumeConfigurationModel struct {
	AuthorizationConfig   fwtypes.ListNestedObjectValueOf[efsAuthorizationConfigModel] `tfsdk:"authorization_config"`
	FileSystemID          types.String                                                 `tfsdk:"file_system_id"`
	RootDirectory         types.String                                                 `tfsdk:"root_directory"`
	TransitEncryption     types.String                                                 `tfsdk:"transit_encryption"`
	TransitEncryptionPort types.Int64                                                  `tfsdk:"transit_encryption_port"`
}

type efsAuthorizationConfigModel struct {
	AccessPointID types.String `tfsdk:"access_point_id"`
	IAM           types.String `tfsdk:"iam"`
}

type hostModel struct {
	SourcePath types.String `tfsdk:"source_path"`
}

type retryStrategyModel struct {
	Attempts       types.Int32                                          `tfsdk:"attempts"`
	EvaluateOnExit fwtypes.ListNestedObjectValueOf[evaluateOnExitModel] `tfsdk:"evaluate_on_exit"`
}

type evaluateOnExitModel struct {
	Action         fwtypes.CaseInsensitiveString `tfsdk:"action"`
	OnExitCode     types.String                  `tfsdk:"on_exit_code"`
	OnReason       types.String                  `tfsdk:"on_reason"`
	OnStatusReason types.String                  `tfsdk:"on_status_reason"`
}

type jobTimeoutModel struct {
	AttemptDurationSeconds types.Int64 `tfsdk:"attempt_duration_seconds"`
}
