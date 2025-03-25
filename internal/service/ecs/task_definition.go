// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_task_definition", name="Task Definition")
// @Tags(identifierAttribute="arn")
func resourceTaskDefinition() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceTaskDefinitionCreate,
		ReadWithoutTimeout:   resourceTaskDefinitionRead,
		UpdateWithoutTimeout: resourceTaskDefinitionUpdate,
		DeleteWithoutTimeout: resourceTaskDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				d.Set(names.AttrARN, d.Id())

				idErr := fmt.Errorf("Expected ID in format of arn:PARTITION:ecs:REGION:ACCOUNTID:task-definition/FAMILY:REVISION and provided: %s", d.Id())
				resARN, err := arn.Parse(d.Id())
				if err != nil {
					return nil, idErr
				}
				familyRevision := strings.TrimPrefix(resARN.Resource, "task-definition/")
				familyRevisionParts := strings.Split(familyRevision, ":")
				if len(familyRevisionParts) != 2 {
					return nil, idErr
				}
				d.SetId(familyRevisionParts[0])

				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaVersion: 1,
		MigrateState:  resourceTaskDefinitionMigrateState,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn_without_revision": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_definitions": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v any) string {
					// Sort the lists of environment variables as they are serialized to state, so we won't get
					// spurious reorderings in plans (diff is suppressed if the environment variables haven't changed,
					// but they still show in the plan if some other property changes).
					orderedCDs, err := expandContainerDefinitions(v.(string))
					if err != nil {
						// e.g. The value is unknown ("74D93920-ED26-11E3-AC10-0800200C9A66").
						// Mimic the pre-v5.59.0 behavior.
						return "[]"
					}
					containerDefinitions(orderedCDs).orderContainers()
					containerDefinitions(orderedCDs).orderEnvironmentVariables()
					containerDefinitions(orderedCDs).orderSecrets()
					containerDefinitions(orderedCDs).compactArrays()
					unnormalizedJson, _ := flattenContainerDefinitions(orderedCDs)
					json, _ := structure.NormalizeJsonString(unnormalizedJson)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					networkMode, ok := d.GetOk("network_mode")
					isAWSVPC := ok && networkMode.(string) == string(awstypes.NetworkModeAwsvpc)
					equal, _ := containerDefinitionsAreEquivalent(old, new, isAWSVPC)
					return equal
				},
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validTaskDefinitionContainerDefinitions,
			},
			"cpu": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"enable_fault_injection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"ephemeral_storage": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_in_gib": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(21, 200),
						},
					},
				},
			},
			names.AttrExecutionRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrFamily: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile("^[0-9A-Za-z_-]+$"), "see https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_TaskDefinition.html"),
				),
			},
			"inference_accelerator": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"device_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"ipc_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IpcMode](),
			},
			"memory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.NetworkMode](),
			},
			"pid_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PidMode](),
			},
			"placement_constraints": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrExpression: {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							ForceNew:         true,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TaskDefinitionPlacementConstraintType](),
						},
					},
				},
			},
			"proxy_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrProperties: {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							ForceNew: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Default:          awstypes.ProxyConfigurationTypeAppmesh,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ProxyConfigurationType](),
						},
					},
				},
			},
			"requires_compatibilities": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"EC2",
						"FARGATE",
						"EXTERNAL",
					}, false),
				},
			},
			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"runtime_platform": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_architecture": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.CPUArchitecture](),
						},
						"operating_system_family": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OSFamily](),
						},
					},
				},
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"task_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"track_latest": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"volume": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configure_at_launch": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"docker_volume_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"autoprovision": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
										Default:  false,
									},
									"driver": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"driver_opts": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
										ForceNew: true,
									},
									"labels": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
										ForceNew: true,
									},
									names.AttrScope: {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.Scope](),
									},
								},
							},
						},
						"efs_volume_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authorization_config": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_point_id": {
													Type:     schema.TypeString,
													ForceNew: true,
													Optional: true,
												},
												"iam": {
													Type:             schema.TypeString,
													ForceNew:         true,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.EFSAuthorizationConfigIAM](),
												},
											},
										},
									},
									names.AttrFileSystemID: {
										Type:     schema.TypeString,
										ForceNew: true,
										Required: true,
									},
									"root_directory": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										Default:  "/",
									},
									"transit_encryption": {
										Type:             schema.TypeString,
										ForceNew:         true,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EFSTransitEncryption](),
									},
									"transit_encryption_port": {
										Type:         schema.TypeInt,
										ForceNew:     true,
										Optional:     true,
										ValidateFunc: validation.IsPortNumberOrZero,
										Default:      0,
									},
								},
							},
						},
						"fsx_windows_file_server_volume_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authorization_config": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"credentials_parameter": {
													Type:         schema.TypeString,
													ForceNew:     true,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
												names.AttrDomain: {
													Type:     schema.TypeString,
													ForceNew: true,
													Required: true,
												},
											},
										},
									},
									names.AttrFileSystemID: {
										Type:     schema.TypeString,
										ForceNew: true,
										Required: true,
									},
									"root_directory": {
										Type:     schema.TypeString,
										ForceNew: true,
										Required: true,
									},
								},
							},
						},
						"host_path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
				Set: func(v any) int {
					var str strings.Builder
					tfMap := v.(map[string]any)

					if v, ok := tfMap["configure_at_launch"].(bool); ok {
						str.WriteString(strconv.FormatBool(v))
					}
					if v, ok := tfMap["docker_volume_configuration"]; ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
						tfMap := v.([]any)[0].(map[string]any)

						if v, ok := tfMap["autoprovision"].(bool); ok {
							str.WriteString(strconv.FormatBool(v))
						}
						if v, ok := tfMap["driver"].(string); ok {
							if v == "" {
								v = "local"
							}
							str.WriteString(v)
						}
						if v, ok := tfMap["driver_opts"].(map[string]any); ok && len(v) > 0 {
							str.WriteString(strconv.Itoa(sdkv2.HashStringValueMap(flex.ExpandStringValueMap(v))))
						}
						if v, ok := tfMap["labels"].(map[string]any); ok && len(v) > 0 {
							str.WriteString(strconv.Itoa(sdkv2.HashStringValueMap(flex.ExpandStringValueMap(v))))
						}
						if v, ok := tfMap[names.AttrScope].(string); ok {
							if v == "" {
								v = "task"
							}
							str.WriteString(v)
						}
					}
					if v, ok := tfMap["efs_volume_configuration"]; ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
						tfMap := v.([]any)[0].(map[string]any)

						if v, ok := tfMap["authorization_config"]; ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
							tfMap := v.([]any)[0].(map[string]any)

							if v, ok := tfMap["access_point_id"].(string); ok && v != "" {
								str.WriteString(v)
							}
							if v, ok := tfMap["iam"].(string); ok && v != "" {
								str.WriteString(v)
							}
						}
						if v, ok := tfMap[names.AttrFileSystemID].(string); ok && v != "" {
							str.WriteString(v)
						}
						if v, ok := tfMap["root_directory"].(string); ok && v != "" {
							str.WriteString(v)
						}
						if v, ok := tfMap["transit_encryption"].(string); ok && v != "" {
							str.WriteString(v)
						}
						if v, ok := tfMap["transit_encryption_port"].(int); ok && v != 0 {
							str.WriteString(strconv.Itoa(v))
						}
					}
					if v, ok := tfMap["fsx_windows_file_server_volume_configuration"]; ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
						tfMap := v.([]any)[0].(map[string]any)

						if v, ok := tfMap["authorization_config"]; ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
							tfMap := v.([]any)[0].(map[string]any)

							if v, ok := tfMap["credentials_parameter"].(string); ok && v != "" {
								str.WriteString(v)
							}
							if v, ok := tfMap[names.AttrDomain].(string); ok && v != "" {
								str.WriteString(v)
							}
						}
						if v, ok := tfMap[names.AttrFileSystemID].(string); ok && v != "" {
							str.WriteString(v)
						}
						if v, ok := tfMap["root_directory"].(string); ok && v != "" {
							str.WriteString(v)
						}
					}
					str.WriteString(tfMap["host_path"].(string))
					str.WriteString(tfMap[names.AttrName].(string))

					return create.StringHashcode(str.String())
				},
			},
		},
	}
}

func resourceTaskDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition(ctx)

	definitions, err := expandContainerDefinitions(d.Get("container_definitions").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: definitions,
		Family:               aws.String(d.Get(names.AttrFamily).(string)),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("cpu"); ok {
		input.Cpu = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enable_fault_injection"); ok {
		input.EnableFaultInjection = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]any)) > 0 {
		input.EphemeralStorage = expandEphemeralStorage(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrExecutionRoleARN); ok {
		input.ExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("inference_accelerator"); ok {
		input.InferenceAccelerators = expandInferenceAccelerators(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipc_mode"); ok {
		input.IpcMode = awstypes.IpcMode(v.(string))
	}

	if v, ok := d.GetOk("memory"); ok {
		input.Memory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_mode"); ok {
		input.NetworkMode = awstypes.NetworkMode(v.(string))
	}

	if v, ok := d.GetOk("pid_mode"); ok {
		input.PidMode = awstypes.PidMode(v.(string))
	}

	if v := d.Get("placement_constraints").(*schema.Set).List(); len(v) > 0 {
		apiObject, err := expandTaskDefinitionPlacementConstraints(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.PlacementConstraints = apiObject
	}

	if proxyConfigs := d.Get("proxy_configuration").([]any); len(proxyConfigs) > 0 {
		input.ProxyConfiguration = expandProxyConfiguration(proxyConfigs)
	}

	if v, ok := d.GetOk("requires_compatibilities"); ok && v.(*schema.Set).Len() > 0 {
		input.RequiresCompatibilities = flex.ExpandStringyValueSet[awstypes.Compatibility](v.(*schema.Set))
	}

	if runtimePlatformConfigs := d.Get("runtime_platform").([]any); len(runtimePlatformConfigs) > 0 && runtimePlatformConfigs[0] != nil {
		input.RuntimePlatform = expandRuntimePlatform(runtimePlatformConfigs)
	}

	if v, ok := d.GetOk("task_role_arn"); ok {
		input.TaskRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("volume"); ok {
		input.Volumes = expandVolumes(v.(*schema.Set).List())
	}

	output, err := conn.RegisterTaskDefinition(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.RegisterTaskDefinition(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Task Definition (%s): %s", d.Get(names.AttrFamily).(string), err)
	}

	taskDefinition := output.TaskDefinition

	d.SetId(aws.ToString(taskDefinition.Family))
	d.Set(names.AttrARN, taskDefinition.TaskDefinitionArn)

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Get(names.AttrARN).(string), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]any)) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceTaskDefinitionRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECS Task Definition (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTaskDefinitionRead(ctx, d, meta)...)
}

func resourceTaskDefinitionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	familyOrARN := d.Get(names.AttrARN).(string)
	if _, ok := d.GetOk("track_latest"); ok {
		familyOrARN = d.Get(names.AttrFamily).(string)
	}
	taskDefinition, tags, err := findTaskDefinitionByFamilyOrARN(ctx, conn, familyOrARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Task Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition (%s): %s", familyOrARN, err)
	}

	d.SetId(aws.ToString(taskDefinition.Family))
	arn := aws.ToString(taskDefinition.TaskDefinitionArn)
	d.Set(names.AttrARN, arn)
	d.Set("arn_without_revision", taskDefinitionARNStripRevision(arn))
	d.Set("cpu", taskDefinition.Cpu)
	d.Set("enable_fault_injection", taskDefinition.EnableFaultInjection)
	if err := d.Set("ephemeral_storage", flattenEphemeralStorage(taskDefinition.EphemeralStorage)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ephemeral_storage: %s", err)
	}
	d.Set(names.AttrExecutionRoleARN, taskDefinition.ExecutionRoleArn)
	d.Set(names.AttrFamily, taskDefinition.Family)
	if err := d.Set("inference_accelerator", flattenInferenceAccelerators(taskDefinition.InferenceAccelerators)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting inference accelerators: %s", err)
	}
	d.Set("ipc_mode", taskDefinition.IpcMode)
	d.Set("memory", taskDefinition.Memory)
	d.Set("network_mode", taskDefinition.NetworkMode)
	d.Set("pid_mode", taskDefinition.PidMode)
	if err := d.Set("placement_constraints", flattenTaskDefinitionPlacementConstraints(taskDefinition.PlacementConstraints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting placement_constraints: %s", err)
	}
	if err := d.Set("proxy_configuration", flattenProxyConfiguration(taskDefinition.ProxyConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting proxy_configuration: %s", err)
	}
	d.Set("requires_compatibilities", taskDefinition.RequiresCompatibilities)
	d.Set("revision", taskDefinition.Revision)
	if err := d.Set("runtime_platform", flattenRuntimePlatform(taskDefinition.RuntimePlatform)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting runtime_platform: %s", err)
	}
	d.Set("task_role_arn", taskDefinition.TaskRoleArn)
	d.Set("track_latest", d.Get("track_latest"))
	if err := d.Set("volume", flattenVolumes(taskDefinition.Volumes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting volume: %s", err)
	}

	// Sort the lists of environment variables as they come in, so we won't get spurious reorderings in plans
	// (diff is suppressed if the environment variables haven't changed, but they still show in the plan if
	// some other property changes).
	containerDefinitions(taskDefinition.ContainerDefinitions).orderContainers()
	containerDefinitions(taskDefinition.ContainerDefinitions).orderEnvironmentVariables()
	containerDefinitions(taskDefinition.ContainerDefinitions).orderSecrets()

	defs, err := flattenContainerDefinitions(taskDefinition.ContainerDefinitions)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("container_definitions", defs)

	setTagsOut(ctx, tags)

	return diags
}

func resourceTaskDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceTaskDefinitionRead(ctx, d, meta)...)
}

func resourceTaskDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining ECS Task Definition Revision: %s", d.Id())
		return diags
	}

	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	log.Printf("[DEBUG] Deleting ECS Task Definition: %s", d.Id())
	_, err := conn.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get(names.AttrARN).(string)),
	})

	if tfawserr.ErrMessageContains(err, "ClientException", "in the process of being deleted") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Task Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func findTaskDefinition(ctx context.Context, conn *ecs.Client, input *ecs.DescribeTaskDefinitionInput) (*awstypes.TaskDefinition, []awstypes.Tag, error) {
	output, err := conn.DescribeTaskDefinition(ctx, input)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusBadRequest) {
		return nil, nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, nil, err
	}

	if output == nil || output.TaskDefinition == nil {
		return nil, nil, tfresource.NewEmptyResultError(input)
	}

	return output.TaskDefinition, output.Tags, nil
}

func findTaskDefinitionByFamilyOrARN(ctx context.Context, conn *ecs.Client, familyOrARN string) (*awstypes.TaskDefinition, []awstypes.Tag, error) {
	input := &ecs.DescribeTaskDefinitionInput{
		Include:        []awstypes.TaskDefinitionField{awstypes.TaskDefinitionFieldTags},
		TaskDefinition: aws.String(familyOrARN),
	}

	taskDefinition, tags, err := findTaskDefinition(ctx, conn, input)

	// Some partitions (i.e., ISO) may not support tagging, giving error.
	if errs.IsUnsupportedOperationInPartitionError(partitionFromConn(conn), err) {
		input.Include = nil

		taskDefinition, tags, err = findTaskDefinition(ctx, conn, input)
	}

	if err != nil {
		return nil, nil, err
	}

	if status := taskDefinition.Status; status == awstypes.TaskDefinitionStatusInactive || status == awstypes.TaskDefinitionStatusDeleteInProgress {
		return nil, nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return taskDefinition, tags, nil
}

func validTaskDefinitionContainerDefinitions(v any, k string) (ws []string, errors []error) {
	_, err := expandContainerDefinitions(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("ECS Task Definition container_definitions is invalid: %s", err))
	}
	return
}

func flattenTaskDefinitionPlacementConstraints(apiObjects []awstypes.TaskDefinitionPlacementConstraint) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
		tfMap[names.AttrType] = apiObject.Type

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRuntimePlatform(apiObject *awstypes.RuntimePlatform) []any {
	if apiObject == nil {
		return nil
	}

	os, cpu := apiObject.OperatingSystemFamily, apiObject.CpuArchitecture

	if os == "" && cpu == "" {
		return nil
	}

	tfMap := make(map[string]any)

	if cpu != "" {
		tfMap["cpu_architecture"] = cpu
	}
	if os != "" {
		tfMap["operating_system_family"] = os
	}

	return []any{
		tfMap,
	}
}

func flattenProxyConfiguration(apiObject *awstypes.ProxyConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	meshProperties := make(map[string]string)
	for _, property := range apiObject.Properties {
		meshProperties[aws.ToString(property.Name)] = aws.ToString(property.Value)
	}

	tfMap := make(map[string]any)
	tfMap["container_name"] = aws.ToString(apiObject.ContainerName)
	tfMap[names.AttrProperties] = meshProperties
	tfMap[names.AttrType] = apiObject.Type

	return []any{
		tfMap,
	}
}

func flattenInferenceAccelerators(apiObjects []awstypes.InferenceAccelerator) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrDeviceName: aws.ToString(apiObject.DeviceName),
			"device_type":        aws.ToString(apiObject.DeviceType),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandInferenceAccelerators(tfList []any) []awstypes.InferenceAccelerator {
	apiObjects := make([]awstypes.InferenceAccelerator, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.InferenceAccelerator{
			DeviceName: aws.String(tfMap[names.AttrDeviceName].(string)),
			DeviceType: aws.String(tfMap["device_type"].(string)),
		}
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandTaskDefinitionPlacementConstraints(tfList []any) ([]awstypes.TaskDefinitionPlacementConstraint, error) {
	var apiObjects []awstypes.TaskDefinitionPlacementConstraint

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		t := tfMap[names.AttrType].(string)
		e := tfMap[names.AttrExpression].(string)
		if err := validPlacementConstraint(t, e); err != nil {
			return nil, err
		}
		apiObjects = append(apiObjects, awstypes.TaskDefinitionPlacementConstraint{
			Expression: aws.String(e),
			Type:       awstypes.TaskDefinitionPlacementConstraintType(t),
		})
	}

	return apiObjects, nil
}

func expandRuntimePlatform(tfList []any) *awstypes.RuntimePlatform {
	tfMapRaw := tfList[0]
	tfMap := tfMapRaw.(map[string]any)
	apiObject := &awstypes.RuntimePlatform{}

	if v := tfMap["cpu_architecture"].(string); v != "" {
		apiObject.CpuArchitecture = awstypes.CPUArchitecture(v)
	}
	if v := tfMap["operating_system_family"].(string); v != "" {
		apiObject.OperatingSystemFamily = awstypes.OSFamily(v)
	}

	return apiObject
}

func expandProxyConfiguration(tfList []any) *awstypes.ProxyConfiguration {
	tfMapRaw := tfList[0]
	tfMap := tfMapRaw.(map[string]any)

	properties := make([]awstypes.KeyValuePair, 0)
	for k, v := range flex.ExpandStringValueMap(tfMap[names.AttrProperties].(map[string]any)) {
		properties = append(properties, awstypes.KeyValuePair{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	apiObject := &awstypes.ProxyConfiguration{
		ContainerName: aws.String(tfMap["container_name"].(string)),
		Properties:    properties,
		Type:          awstypes.ProxyConfigurationType(tfMap[names.AttrType].(string)),
	}

	return apiObject
}

func expandVolumes(tfList []any) []awstypes.Volume {
	apiObjects := make([]awstypes.Volume, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.Volume{
			Name: aws.String(tfMap[names.AttrName].(string)),
		}

		if v, ok := tfMap["configure_at_launch"].(bool); ok {
			apiObject.ConfiguredAtLaunch = aws.Bool(v)
		}

		if v, ok := tfMap["docker_volume_configuration"].([]any); ok && len(v) > 0 {
			apiObject.DockerVolumeConfiguration = expandDockerVolumeConfiguration(v)
		}

		if v, ok := tfMap["efs_volume_configuration"].([]any); ok && len(v) > 0 {
			apiObject.EfsVolumeConfiguration = expandEFSVolumeConfiguration(v)
		}

		if v, ok := tfMap["fsx_windows_file_server_volume_configuration"].([]any); ok && len(v) > 0 {
			apiObject.FsxWindowsFileServerVolumeConfiguration = expandFSxWindowsFileServerVolumeConfiguration(v)
		}

		if v := tfMap["host_path"].(string); v != "" {
			apiObject.Host = &awstypes.HostVolumeProperties{
				SourcePath: aws.String(v),
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDockerVolumeConfiguration(tfList []any) *awstypes.DockerVolumeConfiguration {
	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.DockerVolumeConfiguration{}

	if v, ok := tfMap[names.AttrScope].(string); ok && v != "" {
		apiObject.Scope = awstypes.Scope(v)
	}

	if v, ok := tfMap["autoprovision"]; ok && v != "" {
		if apiObject.Scope != awstypes.ScopeTask || v.(bool) {
			apiObject.Autoprovision = aws.Bool(v.(bool))
		}
	}

	if v, ok := tfMap["driver"].(string); ok && v != "" {
		apiObject.Driver = aws.String(v)
	}

	if v, ok := tfMap["driver_opts"].(map[string]any); ok && len(v) > 0 {
		apiObject.DriverOpts = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["labels"].(map[string]any); ok && len(v) > 0 {
		apiObject.Labels = flex.ExpandStringValueMap(v)
	}

	return apiObject
}

func expandEFSVolumeConfiguration(tfList []any) *awstypes.EFSVolumeConfiguration {
	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.EFSVolumeConfiguration{}

	if v, ok := tfMap["authorization_config"].([]any); ok && len(v) > 0 {
		apiObject.AuthorizationConfig = expandEFSAuthorizationConfig(v)
	}

	if v, ok := tfMap[names.AttrFileSystemID].(string); ok && v != "" {
		apiObject.FileSystemId = aws.String(v)
	}

	if v, ok := tfMap["root_directory"].(string); ok && v != "" {
		apiObject.RootDirectory = aws.String(v)
	}

	if v, ok := tfMap["transit_encryption"].(string); ok && v != "" {
		apiObject.TransitEncryption = awstypes.EFSTransitEncryption(v)
	}

	if v, ok := tfMap["transit_encryption_port"].(int); ok && v > 0 {
		apiObject.TransitEncryptionPort = aws.Int32(int32(v))
	}

	return apiObject
}

func expandEFSAuthorizationConfig(tfList []any) *awstypes.EFSAuthorizationConfig {
	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.EFSAuthorizationConfig{}

	if v, ok := tfMap["access_point_id"].(string); ok && v != "" {
		apiObject.AccessPointId = aws.String(v)
	}

	if v, ok := tfMap["iam"].(string); ok && v != "" {
		apiObject.Iam = awstypes.EFSAuthorizationConfigIAM(v)
	}

	return apiObject
}

func expandFSxWindowsFileServerVolumeConfiguration(tfList []any) *awstypes.FSxWindowsFileServerVolumeConfiguration {
	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.FSxWindowsFileServerVolumeConfiguration{}

	if v, ok := tfMap["authorization_config"].([]any); ok && len(v) > 0 {
		apiObject.AuthorizationConfig = expandFSxWindowsFileServerAuthorizationConfig(v)
	}

	if v, ok := tfMap[names.AttrFileSystemID].(string); ok && v != "" {
		apiObject.FileSystemId = aws.String(v)
	}

	if v, ok := tfMap["root_directory"].(string); ok && v != "" {
		apiObject.RootDirectory = aws.String(v)
	}

	return apiObject
}

func expandFSxWindowsFileServerAuthorizationConfig(tfList []any) *awstypes.FSxWindowsFileServerAuthorizationConfig {
	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.FSxWindowsFileServerAuthorizationConfig{}

	if v, ok := tfMap["credentials_parameter"].(string); ok && v != "" {
		apiObject.CredentialsParameter = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDomain].(string); ok && v != "" {
		apiObject.Domain = aws.String(v)
	}

	return apiObject
}

func flattenVolumes(apiObjects []awstypes.Volume) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrName: aws.ToString(apiObject.Name),
		}

		if apiObject.ConfiguredAtLaunch != nil {
			tfMap["configure_at_launch"] = aws.ToBool(apiObject.ConfiguredAtLaunch)
		}

		if apiObject.DockerVolumeConfiguration != nil {
			tfMap["docker_volume_configuration"] = flattenDockerVolumeConfiguration(apiObject.DockerVolumeConfiguration)
		}

		if apiObject.EfsVolumeConfiguration != nil {
			tfMap["efs_volume_configuration"] = flattenEFSVolumeConfiguration(apiObject.EfsVolumeConfiguration)
		}

		if apiObject.FsxWindowsFileServerVolumeConfiguration != nil {
			tfMap["fsx_windows_file_server_volume_configuration"] = flattenFSxWindowsFileServerVolumeConfiguration(apiObject.FsxWindowsFileServerVolumeConfiguration)
		}

		if apiObject.Host != nil && apiObject.Host.SourcePath != nil {
			tfMap["host_path"] = aws.ToString(apiObject.Host.SourcePath)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDockerVolumeConfiguration(apiObject *awstypes.DockerVolumeConfiguration) []any {
	var tfList []any
	tfMap := make(map[string]any)

	if v := apiObject.Autoprovision; v != nil {
		tfMap["autoprovision"] = aws.ToBool(v)
	}

	if v := apiObject.Driver; v != nil {
		tfMap["driver"] = aws.ToString(v)
	}

	if v := apiObject.DriverOpts; v != nil {
		tfMap["driver_opts"] = v
	}

	if v := apiObject.Labels; v != nil {
		tfMap["labels"] = v
	}

	tfMap[names.AttrScope] = apiObject.Scope

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenEFSVolumeConfiguration(apiObject *awstypes.EFSVolumeConfiguration) []any {
	var tfList []any
	tfMap := make(map[string]any)

	if apiObject != nil {
		if v := apiObject.AuthorizationConfig; v != nil {
			tfMap["authorization_config"] = flattenEFSAuthorizationConfig(v)
		}

		if v := apiObject.FileSystemId; v != nil {
			tfMap[names.AttrFileSystemID] = aws.ToString(v)
		}

		if v := apiObject.RootDirectory; v != nil {
			tfMap["root_directory"] = aws.ToString(v)
		}

		tfMap["transit_encryption"] = apiObject.TransitEncryption

		if v := apiObject.TransitEncryptionPort; v != nil {
			tfMap["transit_encryption_port"] = aws.ToInt32(v)
		}
	}

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenEFSAuthorizationConfig(apiObject *awstypes.EFSAuthorizationConfig) []any {
	var tfList []any
	tfMap := make(map[string]any)

	if apiObject != nil {
		if v := apiObject.AccessPointId; v != nil {
			tfMap["access_point_id"] = aws.ToString(v)
		}

		tfMap["iam"] = apiObject.Iam
	}

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenFSxWindowsFileServerVolumeConfiguration(apiObject *awstypes.FSxWindowsFileServerVolumeConfiguration) []any {
	var tfList []any
	tfMap := make(map[string]any)

	if apiObject != nil {
		if v := apiObject.AuthorizationConfig; v != nil {
			tfMap["authorization_config"] = flattenFSxWindowsFileServerAuthorizationConfig(v)
		}

		if v := apiObject.FileSystemId; v != nil {
			tfMap[names.AttrFileSystemID] = aws.ToString(v)
		}

		if v := apiObject.RootDirectory; v != nil {
			tfMap["root_directory"] = aws.ToString(v)
		}
	}

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenFSxWindowsFileServerAuthorizationConfig(apiObject *awstypes.FSxWindowsFileServerAuthorizationConfig) []any {
	var tfList []any
	tfMap := make(map[string]any)

	if apiObject != nil {
		if v := apiObject.CredentialsParameter; v != nil {
			tfMap["credentials_parameter"] = aws.ToString(v)
		}

		if v := apiObject.Domain; v != nil {
			tfMap[names.AttrDomain] = aws.ToString(v)
		}
	}

	tfList = append(tfList, tfMap)

	return tfList
}

func expandEphemeralStorage(tfList []any) *awstypes.EphemeralStorage {
	tfMap := tfList[0].(map[string]any)

	apiObject := &awstypes.EphemeralStorage{
		SizeInGiB: int32(tfMap["size_in_gib"].(int)),
	}

	return apiObject
}

func flattenEphemeralStorage(apiObject *awstypes.EphemeralStorage) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)
	tfMap["size_in_gib"] = apiObject.SizeInGiB

	return []any{tfMap}
}

// taskDefinitionARNStripRevision strips the trailing revision number from a task definition ARN
//
// Invalid ARNs will return an empty string. ARNs with an unexpected number of
// separators in the resource section are returned unmodified.
func taskDefinitionARNStripRevision(s string) string {
	tdArn, err := arn.Parse(s)
	if err != nil {
		return ""
	}
	parts := strings.Split(tdArn.Resource, ":")
	if len(parts) == 2 {
		tdArn.Resource = parts[0]
	}
	return tdArn.String()
}
