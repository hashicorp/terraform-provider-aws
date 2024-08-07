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
// @Testing(tagsTest=true)
func newJobDefinitionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &jobDefinitionDataSource{}, nil
}

type jobDefinitionDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *jobDefinitionDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_batch_job_definition"
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
			"eks_properties": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[jobDefinitionEKSPropertiesModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"pod_properties": fwtypes.NewListNestedObjectTypeOf[jobDefinitionEKSPodPropertiesModel](ctx),
					},
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Optional: true,
			},
			"node_properties": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[jobDefinitionNodePropertiesModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"main_node":             types.Int64Type,
						"node_range_properties": fwtypes.NewListNestedObjectTypeOf[jobDefinitionNodeRangePropertyModel](ctx),
						"num_nodes":             types.Int64Type,
					},
				},
			},
			"retry_strategy": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[jobDefinitionRetryStrategyModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attempts":         types.Int64Type,
						"evaluate_on_exit": fwtypes.NewListNestedObjectTypeOf[jobDefinitionEvaluateOnExitModel](ctx),
					},
				},
			},
			"revision": schema.Int64Attribute{
				Optional: true,
			},
			"scheduling_priority": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(jobDefinitionStatus_Values()...),
				},
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			names.AttrTimeout: schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[jobDefinitionJobTimeoutModel](ctx),
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

		output, err := findJobDefinitionV2(ctx, conn, input)

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

		output, err := findJobDefinitionsV2(ctx, conn, input)

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
	data.Tags = fwflex.FlattenFrameworkStringValueMap(ctx, jd.Tags)

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

func findJobDefinitionV2(ctx context.Context, conn *batch.Client, input *batch.DescribeJobDefinitionsInput) (*awstypes.JobDefinition, error) {
	output, err := findJobDefinitionsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findJobDefinitionsV2(ctx context.Context, conn *batch.Client, input *batch.DescribeJobDefinitionsInput) ([]awstypes.JobDefinition, error) {
	var output []awstypes.JobDefinition

	pages := batch.NewDescribeJobDefinitionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.JobDefinitions...)
	}

	return output, nil
}

type jobDefinitionDataSourceModel struct {
	ARNPrefix                  types.String                                                      `tfsdk:"arn_prefix"`
	ContainerOrchestrationType types.String                                                      `tfsdk:"container_orchestration_type"`
	EKSProperties              fwtypes.ListNestedObjectValueOf[jobDefinitionEKSPropertiesModel]  `tfsdk:"eks_properties"`
	ID                         types.String                                                      `tfsdk:"id"`
	JobDefinitionARN           fwtypes.ARN                                                       `tfsdk:"arn"`
	JobDefinitionName          types.String                                                      `tfsdk:"name"`
	NodeProperties             fwtypes.ListNestedObjectValueOf[jobDefinitionNodePropertiesModel] `tfsdk:"node_properties"`
	RetryStrategy              fwtypes.ListNestedObjectValueOf[jobDefinitionRetryStrategyModel]  `tfsdk:"retry_strategy"`
	Revision                   types.Int64                                                       `tfsdk:"revision"`
	SchedulingPriority         types.Int64                                                       `tfsdk:"scheduling_priority"`
	Status                     types.String                                                      `tfsdk:"status"`
	Tags                       types.Map                                                         `tfsdk:"tags"`
	Timeout                    fwtypes.ListNestedObjectValueOf[jobDefinitionJobTimeoutModel]     `tfsdk:"timeout"`
	Type                       types.String                                                      `tfsdk:"type"`
}

type jobDefinitionEKSPropertiesModel struct {
	PodProperties fwtypes.ListNestedObjectValueOf[jobDefinitionEKSPodPropertiesModel] `tfsdk:"pod_properties"`
}

type jobDefinitionEKSPodPropertiesModel struct {
	Containers         fwtypes.ListNestedObjectValueOf[jobDefinitionEKSContainerModel] `tfsdk:"containers"`
	DNSPolicy          types.String                                                    `tfsdk:"dns_policy"`
	HostNetwork        types.Bool                                                      `tfsdk:"host_network"`
	Metadata           fwtypes.ListNestedObjectValueOf[jobDefinitionEKSMetadataModel]  `tfsdk:"metadata"`
	ServiceAccountName types.Bool                                                      `tfsdk:"service_account_name"`
	Volumes            fwtypes.ListNestedObjectValueOf[jobDefinitionEKSVolumeModel]    `tfsdk:"volumes"`
}

type jobDefinitionEKSContainerModel struct {
	Args            fwtypes.ListValueOf[types.String]                                                   `tfsdk:"args"`
	Command         fwtypes.ListValueOf[types.String]                                                   `tfsdk:"command"`
	Env             fwtypes.ListNestedObjectValueOf[jobDefinitionEKSContainerEnvironmentVariableModel]  `tfsdk:"env"`
	Image           types.String                                                                        `tfsdk:"image"`
	ImagePullPolicy types.String                                                                        `tfsdk:"image_pull_policy"`
	Name            types.String                                                                        `tfsdk:"name"`
	Resources       fwtypes.ListNestedObjectValueOf[jobDefinitionEKSContainerResourceRequirementsModel] `tfsdk:"resources"`
	SecurityContext fwtypes.ListNestedObjectValueOf[jobDefinitionEKSContainerSecurityContextModel]      `tfsdk:"security_context"`
	VolumeMounts    fwtypes.ListNestedObjectValueOf[jobDefinitionEKSContainerVolumeMountModel]          `tfsdk:"volume_mounts"`
}

type jobDefinitionEKSContainerEnvironmentVariableModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type jobDefinitionEKSContainerResourceRequirementsModel struct {
	Limits   fwtypes.MapValueOf[types.String] `tfsdk:"limits"`
	Requests fwtypes.MapValueOf[types.String] `tfsdk:"requests"`
}

type jobDefinitionEKSContainerSecurityContextModel struct {
	Privileged             types.Bool  `tfsdk:"privileged"`
	ReadOnlyRootFilesystem types.Bool  `tfsdk:"read_only_root_file_system"`
	RunAsGroup             types.Int64 `tfsdk:"run_as_group"`
	RunAsNonRoot           types.Bool  `tfsdk:"run_as_non_root"`
	RunAsUser              types.Int64 `tfsdk:"run_as_user"`
}

type jobDefinitionEKSContainerVolumeMountModel struct {
	MountPath types.String `tfsdk:"mount_path"`
	Name      types.String `tfsdk:"name"`
	ReadOnly  types.Bool   `tfsdk:"read_only"`
}

type jobDefinitionEKSMetadataModel struct {
	Labels fwtypes.MapValueOf[types.String] `tfsdk:"labels"`
}

type jobDefinitionEKSVolumeModel struct {
	EmptyDir fwtypes.ListNestedObjectValueOf[jobDefinitionEKSEmptyDirModel] `tfsdk:"empty_dir"`
	HostPath fwtypes.ListNestedObjectValueOf[jobDefinitionEKSHostPathModel] `tfsdk:"host_path"`
	Name     types.String                                                   `tfsdk:"name"`
	Secret   fwtypes.ListNestedObjectValueOf[jobDefinitionEKSSecretModel]   `tfsdk:"secret"`
}

type jobDefinitionEKSEmptyDirModel struct {
	Medium    types.String `tfsdk:"medium"`
	SizeLimit types.String `tfsdk:"size_limit"`
}

type jobDefinitionEKSHostPathModel struct {
	Path types.String `tfsdk:"path"`
}

type jobDefinitionEKSSecretModel struct {
	Optional   types.Bool   `tfsdk:"optional"`
	SecretName types.String `tfsdk:"secret_name"`
}

type jobDefinitionNodePropertiesModel struct {
	MainNode            types.Int64                                                          `tfsdk:"main_node"`
	NodeRangeProperties fwtypes.ListNestedObjectValueOf[jobDefinitionNodeRangePropertyModel] `tfsdk:"node_range_properties"`
	NumNodes            types.Int64                                                          `tfsdk:"num_nodes"`
}

type jobDefinitionNodeRangePropertyModel struct {
	Container   fwtypes.ListNestedObjectValueOf[jobDefinitionContainerPropertiesModel] `tfsdk:"container"`
	TargetNodes types.String                                                           `tfsdk:"target_nodes"`
}

type jobDefinitionContainerPropertiesModel struct {
	Command                      fwtypes.ListValueOf[types.String]                                               `tfsdk:"command"`
	Environment                  fwtypes.ListNestedObjectValueOf[jobDefinitionKeyValuePairModel]                 `tfsdk:"environment"`
	EphemeralStorage             fwtypes.ListNestedObjectValueOf[jobDefinitionEphemeralStorageModel]             `tfsdk:"ephemeral_storage"`
	ExecutionRoleARN             types.String                                                                    `tfsdk:"execution_role_arn"`
	FargatePlatformConfiguration fwtypes.ListNestedObjectValueOf[jobDefinitionFargatePlatformConfigurationModel] `tfsdk:"fargate_platform_configuration"`
	Image                        types.String                                                                    `tfsdk:"image"`
	InstanceType                 types.String                                                                    `tfsdk:"instance_type"`
	JobRoleARN                   types.String                                                                    `tfsdk:"job_role_arn"`
	LinuxParameters              fwtypes.ListNestedObjectValueOf[jobDefinitionLinuxParametersModel]              `tfsdk:"linux_parameters"`
	LogConfiguration             fwtypes.ListNestedObjectValueOf[jobDefinitionLogConfigurationModel]             `tfsdk:"log_configuration"`
	MountPoints                  fwtypes.ListNestedObjectValueOf[jobDefinitionMountPointModel]                   `tfsdk:"mount_points"`
	NetworkConfiguration         fwtypes.ListNestedObjectValueOf[jobDefinitionNetworkConfigurationModel]         `tfsdk:"network_configuration"`
	Privileged                   types.Bool                                                                      `tfsdk:"privileged"`
	ReadonlyRootFilesystem       types.Bool                                                                      `tfsdk:"readonly_root_filesystem"`
	ResourceRequirements         fwtypes.ListNestedObjectValueOf[jobDefinitionResourceRequirementModel]          `tfsdk:"resource_requirements"`
	RuntimePlatform              fwtypes.ListNestedObjectValueOf[jobDefinitionRuntimePlatformModel]              `tfsdk:"runtime_platform"`
	Secrets                      fwtypes.ListNestedObjectValueOf[jobDefinitionSecretModel]                       `tfsdk:"secrets"`
	Ulimits                      fwtypes.ListNestedObjectValueOf[jobDefinitionUlimitModel]                       `tfsdk:"ulimits"`
	User                         types.String                                                                    `tfsdk:"user"`
	Volumes                      fwtypes.ListNestedObjectValueOf[jobDefinitionVolumeModel]                       `tfsdk:"volumes"`
}

type jobDefinitionKeyValuePairModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type jobDefinitionEphemeralStorageModel struct {
	SizeInGiB types.Int64 `tfsdk:"size_in_gib"`
}

type jobDefinitionFargatePlatformConfigurationModel struct {
	PlatformVersion types.String `tfsdk:"platform_version"`
}

type jobDefinitionLinuxParametersModel struct {
	Devices            fwtypes.ListNestedObjectValueOf[jobDefinitionDeviceModel] `tfsdk:"devices"`
	InitProcessEnabled types.Bool                                                `tfsdk:"init_process_enabled"`
	MaxSwap            types.Int64                                               `tfsdk:"max_swap"`
	SharedMemorySize   types.Int64                                               `tfsdk:"shared_memory_size"`
	Swappiness         types.Int64                                               `tfsdk:"swappiness"`
	Tmpfs              fwtypes.ListNestedObjectValueOf[jobDefinitionTmpfsModel]  `tfsdk:"tmpfs"`
}

type jobDefinitionDeviceModel struct {
	ContainerPath types.String                      `tfsdk:"container_path"`
	HostPath      types.String                      `tfsdk:"host_path"`
	Permissions   fwtypes.ListValueOf[types.String] `tfsdk:"permissions"`
}

type jobDefinitionTmpfsModel struct {
	ContainerPath types.String                      `tfsdk:"container_path"`
	MountOptions  fwtypes.ListValueOf[types.String] `tfsdk:"mount_options"`
	Size          types.Int64                       `tfsdk:"size"`
}

type jobDefinitionLogConfigurationModel struct {
	LogDriver     types.String                                              `tfsdk:"log_driver"`
	Options       fwtypes.MapValueOf[types.String]                          `tfsdk:"options"`
	SecretOptions fwtypes.ListNestedObjectValueOf[jobDefinitionSecretModel] `tfsdk:"secret_options"`
}

type jobDefinitionSecretModel struct {
	Name      types.String `tfsdk:"name"`
	ValueFrom types.String `tfsdk:"value_from"`
}

type jobDefinitionMountPointModel struct {
	ContainerPath types.String `tfsdk:"container_path"`
	ReadOnly      types.Bool   `tfsdk:"read_only"`
	SourceVolume  types.String `tfsdk:"source_volume"`
}

type jobDefinitionNetworkConfigurationModel struct {
	AssignPublicIP types.Bool `tfsdk:"assign_public_ip"`
}

type jobDefinitionResourceRequirementModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type jobDefinitionRuntimePlatformModel struct {
	CPUArchitecture       types.String `tfsdk:"cpu_architecture"`
	OperatingSystemFamily types.String `tfsdk:"operating_system_family"`
}

type jobDefinitionUlimitModel struct {
	HardLimit types.Int64  `tfsdk:"hard_limit"`
	Name      types.String `tfsdk:"name"`
	SoftLimit types.Int64  `tfsdk:"soft_limit"`
}

type jobDefinitionVolumeModel struct {
	EFSVolumeConfiguration fwtypes.ListNestedObjectValueOf[jobDefinitionEFSVolumeConfigurationModel] `tfsdk:"efs_volume_configuration"`
	Host                   fwtypes.ListNestedObjectValueOf[jobDefinitionHostModel]                   `tfsdk:"host"`
	Name                   types.String                                                              `tfsdk:"name"`
}

type jobDefinitionEFSVolumeConfigurationModel struct {
	AuthorizationConfig   fwtypes.ListNestedObjectValueOf[jobDefinitionEFSAuthorizationConfigModel] `tfsdk:"authorization_config"`
	FileSystemID          types.String                                                              `tfsdk:"file_system_id"`
	RootDirectory         types.String                                                              `tfsdk:"root_directory"`
	TransitEncryption     types.String                                                              `tfsdk:"transit_encryption"`
	TransitEncryptionPort types.Int64                                                               `tfsdk:"transit_encryption_port"`
}

type jobDefinitionEFSAuthorizationConfigModel struct {
	AccessPointID types.String `tfsdk:"access_point_id"`
	IAM           types.String `tfsdk:"iam"`
}

type jobDefinitionHostModel struct {
	SourcePath types.String `tfsdk:"source_path"`
}

type jobDefinitionRetryStrategyModel struct {
	Attempts       types.Int64                                                       `tfsdk:"attempts"`
	EvaluateOnExit fwtypes.ListNestedObjectValueOf[jobDefinitionEvaluateOnExitModel] `tfsdk:"evaluate_on_exit"`
}

type jobDefinitionEvaluateOnExitModel struct {
	Action         types.String `tfsdk:"action"`
	OnExitCode     types.String `tfsdk:"on_exit_code"`
	OnReason       types.String `tfsdk:"on_reason"`
	OnStatusReason types.String `tfsdk:"on_status_reason"`
}

type jobDefinitionJobTimeoutModel struct {
	AttemptDurationSeconds types.Int64 `tfsdk:"attempt_duration_seconds"`
}
