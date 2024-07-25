// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
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
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
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
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

		CustomizeDiff: verify.SetTagsDiff,

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
				StateFunc: func(v interface{}) string {
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
				ValidateFunc: validTaskDefinitionContainerDefinitions,
			},
			"cpu": {
				Type:     schema.TypeString,
				Optional: true,
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
										ForceNew: true,
										Optional: true,
									},
									"driver_opts": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										ForceNew: true,
										Optional: true,
									},
									"labels": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										ForceNew: true,
										Optional: true,
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
				Set: resourceTaskDefinitionVolumeHash,
			},
		},
	}
}

func resourceTaskDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

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

	if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]interface{})) > 0 {
		input.EphemeralStorage = expandTaskDefinitionEphemeralStorage(v.([]interface{}))
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

	if proxyConfigs := d.Get("proxy_configuration").([]interface{}); len(proxyConfigs) > 0 {
		input.ProxyConfiguration = expandTaskDefinitionProxyConfiguration(proxyConfigs)
	}

	if v, ok := d.GetOk("requires_compatibilities"); ok && v.(*schema.Set).Len() > 0 {
		input.RequiresCompatibilities = flex.ExpandStringyValueSet[awstypes.Compatibility](v.(*schema.Set))
	}

	if runtimePlatformConfigs := d.Get("runtime_platform").([]interface{}); len(runtimePlatformConfigs) > 0 && runtimePlatformConfigs[0] != nil {
		input.RuntimePlatform = expandTaskDefinitionRuntimePlatformConfiguration(runtimePlatformConfigs)
	}

	if v, ok := d.GetOk("task_role_arn"); ok {
		input.TaskRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("volume"); ok {
		volumes := expandVolumes(v.(*schema.Set).List())
		input.Volumes = volumes
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
		err := createTags(ctx, conn, aws.ToString(taskDefinition.TaskDefinitionArn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceTaskDefinitionRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECS Task Definition (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTaskDefinitionRead(ctx, d, meta)...)
}

func resourceTaskDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	if err := d.Set("ephemeral_storage", flattenTaskDefinitionEphemeralStorage(taskDefinition.EphemeralStorage)); err != nil {
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
	if err := d.Set("placement_constraints", flattenPlacementConstraints(taskDefinition.PlacementConstraints)); err != nil {
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

func resourceTaskDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceTaskDefinitionRead(ctx, d, meta)...)
}

func resourceTaskDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining ECS Task Definition Revision: %s", d.Id())
		return diags
	}

	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	_, err := conn.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get(names.AttrARN).(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Task Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func findTaskDefinition(ctx context.Context, conn *ecs.Client, input *ecs.DescribeTaskDefinitionInput) (*awstypes.TaskDefinition, []awstypes.Tag, error) {
	output, err := conn.DescribeTaskDefinition(ctx, input)

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

	if status := taskDefinition.Status; status == awstypes.TaskDefinitionStatusInactive {
		return nil, nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return taskDefinition, tags, nil
}

func validTaskDefinitionContainerDefinitions(v interface{}, k string) (ws []string, errors []error) {
	_, err := expandContainerDefinitions(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("ECS Task Definition container_definitions is invalid: %s", err))
	}
	return
}

func resourceTaskDefinitionVolumeHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m[names.AttrName].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["host_path"].(string)))

	if v, ok := m["efs_volume_configuration"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		m := v.([]interface{})[0].(map[string]interface{})

		if v, ok := m[names.AttrFileSystemID]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}

		if v, ok := m["root_directory"]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}

		if v, ok := m["transit_encryption"]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}
		if v, ok := m["transit_encryption_port"]; ok && v.(int) > 0 {
			buf.WriteString(fmt.Sprintf("%d-", v.(int)))
		}
		if v, ok := m["authorization_config"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			m := v.([]interface{})[0].(map[string]interface{})
			if v, ok := m["access_point_id"]; ok && v.(string) != "" {
				buf.WriteString(fmt.Sprintf("%s-", v.(string)))
			}
			if v, ok := m["iam"]; ok && v.(string) != "" {
				buf.WriteString(fmt.Sprintf("%s-", v.(string)))
			}
		}
	}

	if v, ok := m["fsx_windows_file_server_volume_configuration"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		m := v.([]interface{})[0].(map[string]interface{})

		if v, ok := m[names.AttrFileSystemID]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}

		if v, ok := m["root_directory"]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}

		if v, ok := m["authorization_config"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			m := v.([]interface{})[0].(map[string]interface{})
			if v, ok := m["credentials_parameter"]; ok && v.(string) != "" {
				buf.WriteString(fmt.Sprintf("%s-", v.(string)))
			}
			if v, ok := m[names.AttrDomain]; ok && v.(string) != "" {
				buf.WriteString(fmt.Sprintf("%s-", v.(string)))
			}
		}
	}

	return create.StringHashcode(buf.String())
}

func flattenPlacementConstraints(pcs []awstypes.TaskDefinitionPlacementConstraint) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c[names.AttrType] = string(pc.Type)
		c[names.AttrExpression] = aws.ToString(pc.Expression)
		results = append(results, c)
	}
	return results
}

func flattenRuntimePlatform(rp *awstypes.RuntimePlatform) []map[string]interface{} {
	if rp == nil {
		return nil
	}

	os := string(rp.OperatingSystemFamily)
	cpu := string(rp.CpuArchitecture)

	if os == "" && cpu == "" {
		return nil
	}

	config := make(map[string]interface{})

	if os != "" {
		config["operating_system_family"] = os
	}
	if cpu != "" {
		config["cpu_architecture"] = cpu
	}

	return []map[string]interface{}{
		config,
	}
}

func flattenProxyConfiguration(pc *awstypes.ProxyConfiguration) []map[string]interface{} {
	if pc == nil {
		return nil
	}

	meshProperties := make(map[string]string)
	if pc.Properties != nil {
		for _, prop := range pc.Properties {
			meshProperties[aws.ToString(prop.Name)] = aws.ToString(prop.Value)
		}
	}

	config := make(map[string]interface{})
	config["container_name"] = aws.ToString(pc.ContainerName)
	config[names.AttrType] = string(pc.Type)
	config[names.AttrProperties] = meshProperties

	return []map[string]interface{}{
		config,
	}
}

func flattenInferenceAccelerators(list []awstypes.InferenceAccelerator) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, iAcc := range list {
		l := map[string]interface{}{
			names.AttrDeviceName: aws.ToString(iAcc.DeviceName),
			"device_type":        aws.ToString(iAcc.DeviceType),
		}

		result = append(result, l)
	}
	return result
}

func expandInferenceAccelerators(configured []interface{}) []awstypes.InferenceAccelerator {
	iAccs := make([]awstypes.InferenceAccelerator, 0, len(configured))
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})
		l := awstypes.InferenceAccelerator{
			DeviceName: aws.String(data[names.AttrDeviceName].(string)),
			DeviceType: aws.String(data["device_type"].(string)),
		}
		iAccs = append(iAccs, l)
	}

	return iAccs
}

func expandTaskDefinitionPlacementConstraints(constraints []interface{}) ([]awstypes.TaskDefinitionPlacementConstraint, error) {
	var pc []awstypes.TaskDefinitionPlacementConstraint
	for _, raw := range constraints {
		p := raw.(map[string]interface{})
		t := p[names.AttrType].(string)
		e := p[names.AttrExpression].(string)
		if err := validPlacementConstraint(t, e); err != nil {
			return nil, err
		}
		pc = append(pc, awstypes.TaskDefinitionPlacementConstraint{
			Type:       awstypes.TaskDefinitionPlacementConstraintType(t),
			Expression: aws.String(e),
		})
	}

	return pc, nil
}

func expandTaskDefinitionRuntimePlatformConfiguration(runtimePlatformConfig []interface{}) *awstypes.RuntimePlatform {
	config := runtimePlatformConfig[0]

	configMap := config.(map[string]interface{})
	ecsProxyConfig := &awstypes.RuntimePlatform{}

	os := configMap["operating_system_family"].(string)
	if os != "" {
		ecsProxyConfig.OperatingSystemFamily = awstypes.OSFamily(os)
	}

	osFamily := configMap["cpu_architecture"].(string)
	if osFamily != "" {
		ecsProxyConfig.CpuArchitecture = awstypes.CPUArchitecture(osFamily)
	}

	return ecsProxyConfig
}

func expandTaskDefinitionProxyConfiguration(proxyConfigs []interface{}) *awstypes.ProxyConfiguration {
	proxyConfig := proxyConfigs[0]
	configMap := proxyConfig.(map[string]interface{})

	rawProperties := configMap[names.AttrProperties].(map[string]interface{})

	properties := make([]awstypes.KeyValuePair, len(rawProperties))
	i := 0
	for name, value := range rawProperties {
		properties[i] = awstypes.KeyValuePair{
			Name:  aws.String(name),
			Value: aws.String(value.(string)),
		}
		i++
	}

	ecsProxyConfig := &awstypes.ProxyConfiguration{
		ContainerName: aws.String(configMap["container_name"].(string)),
		Type:          awstypes.ProxyConfigurationType(configMap[names.AttrType].(string)),
		Properties:    properties,
	}

	return ecsProxyConfig
}

func expandVolumes(configured []interface{}) []awstypes.Volume {
	volumes := make([]awstypes.Volume, 0, len(configured))

	// Loop over our configured volumes and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := awstypes.Volume{
			Name: aws.String(data[names.AttrName].(string)),
		}

		hostPath := data["host_path"].(string)
		if hostPath != "" {
			l.Host = &awstypes.HostVolumeProperties{
				SourcePath: aws.String(hostPath),
			}
		}

		if v, ok := data["configure_at_launch"].(bool); ok {
			l.ConfiguredAtLaunch = aws.Bool(v)
		}

		if v, ok := data["docker_volume_configuration"].([]interface{}); ok && len(v) > 0 {
			l.DockerVolumeConfiguration = expandVolumesDockerVolume(v)
		}

		if v, ok := data["efs_volume_configuration"].([]interface{}); ok && len(v) > 0 {
			l.EfsVolumeConfiguration = expandVolumesEFSVolume(v)
		}

		if v, ok := data["fsx_windows_file_server_volume_configuration"].([]interface{}); ok && len(v) > 0 {
			l.FsxWindowsFileServerVolumeConfiguration = expandVolumesFSxWinVolume(v)
		}

		volumes = append(volumes, l)
	}

	return volumes
}

func expandVolumesDockerVolume(configList []interface{}) *awstypes.DockerVolumeConfiguration {
	config := configList[0].(map[string]interface{})
	dockerVol := &awstypes.DockerVolumeConfiguration{}

	if v, ok := config[names.AttrScope].(string); ok && v != "" {
		dockerVol.Scope = awstypes.Scope(v)
	}

	if v, ok := config["autoprovision"]; ok && v != "" {
		if dockerVol.Scope != awstypes.ScopeTask || v.(bool) {
			dockerVol.Autoprovision = aws.Bool(v.(bool))
		}
	}

	if v, ok := config["driver"].(string); ok && v != "" {
		dockerVol.Driver = aws.String(v)
	}

	if v, ok := config["driver_opts"].(map[string]interface{}); ok && len(v) > 0 {
		dockerVol.DriverOpts = flex.ExpandStringValueMap(v)
	}

	if v, ok := config["labels"].(map[string]interface{}); ok && len(v) > 0 {
		dockerVol.Labels = flex.ExpandStringValueMap(v)
	}

	return dockerVol
}

func expandVolumesEFSVolume(efsConfig []interface{}) *awstypes.EFSVolumeConfiguration {
	config := efsConfig[0].(map[string]interface{})
	efsVol := &awstypes.EFSVolumeConfiguration{}

	if v, ok := config[names.AttrFileSystemID].(string); ok && v != "" {
		efsVol.FileSystemId = aws.String(v)
	}

	if v, ok := config["root_directory"].(string); ok && v != "" {
		efsVol.RootDirectory = aws.String(v)
	}
	if v, ok := config["transit_encryption"].(string); ok && v != "" {
		efsVol.TransitEncryption = awstypes.EFSTransitEncryption(v)
	}

	if v, ok := config["transit_encryption_port"].(int); ok && v > 0 {
		efsVol.TransitEncryptionPort = aws.Int32(int32(v))
	}
	if v, ok := config["authorization_config"].([]interface{}); ok && len(v) > 0 {
		efsVol.AuthorizationConfig = expandVolumesEFSVolumeAuthorizationConfig(v)
	}

	return efsVol
}

func expandVolumesEFSVolumeAuthorizationConfig(efsConfig []interface{}) *awstypes.EFSAuthorizationConfig {
	authconfig := efsConfig[0].(map[string]interface{})
	auth := &awstypes.EFSAuthorizationConfig{}

	if v, ok := authconfig["access_point_id"].(string); ok && v != "" {
		auth.AccessPointId = aws.String(v)
	}

	if v, ok := authconfig["iam"].(string); ok && v != "" {
		auth.Iam = awstypes.EFSAuthorizationConfigIAM(v)
	}

	return auth
}

func expandVolumesFSxWinVolume(fsxWinConfig []interface{}) *awstypes.FSxWindowsFileServerVolumeConfiguration {
	config := fsxWinConfig[0].(map[string]interface{})
	fsxVol := &awstypes.FSxWindowsFileServerVolumeConfiguration{}

	if v, ok := config[names.AttrFileSystemID].(string); ok && v != "" {
		fsxVol.FileSystemId = aws.String(v)
	}

	if v, ok := config["root_directory"].(string); ok && v != "" {
		fsxVol.RootDirectory = aws.String(v)
	}

	if v, ok := config["authorization_config"].([]interface{}); ok && len(v) > 0 {
		fsxVol.AuthorizationConfig = expandVolumesFSxWinVolumeAuthorizationConfig(v)
	}

	return fsxVol
}

func expandVolumesFSxWinVolumeAuthorizationConfig(config []interface{}) *awstypes.FSxWindowsFileServerAuthorizationConfig {
	authconfig := config[0].(map[string]interface{})
	auth := &awstypes.FSxWindowsFileServerAuthorizationConfig{}

	if v, ok := authconfig["credentials_parameter"].(string); ok && v != "" {
		auth.CredentialsParameter = aws.String(v)
	}

	if v, ok := authconfig[names.AttrDomain].(string); ok && v != "" {
		auth.Domain = aws.String(v)
	}

	return auth
}

func flattenVolumes(list []awstypes.Volume) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, volume := range list {
		l := map[string]interface{}{
			names.AttrName: aws.ToString(volume.Name),
		}

		if volume.Host != nil && volume.Host.SourcePath != nil {
			l["host_path"] = aws.ToString(volume.Host.SourcePath)
		}

		if volume.ConfiguredAtLaunch != nil {
			l["configure_at_launch"] = aws.ToBool(volume.ConfiguredAtLaunch)
		}

		if volume.DockerVolumeConfiguration != nil {
			l["docker_volume_configuration"] = flattenDockerVolumeConfiguration(volume.DockerVolumeConfiguration)
		}

		if volume.EfsVolumeConfiguration != nil {
			l["efs_volume_configuration"] = flattenEFSVolumeConfiguration(volume.EfsVolumeConfiguration)
		}

		if volume.FsxWindowsFileServerVolumeConfiguration != nil {
			l["fsx_windows_file_server_volume_configuration"] = flattenFSxWinVolumeConfiguration(volume.FsxWindowsFileServerVolumeConfiguration)
		}

		result = append(result, l)
	}
	return result
}

func flattenDockerVolumeConfiguration(config *awstypes.DockerVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})

	m[names.AttrScope] = string(config.Scope)

	if v := config.Autoprovision; v != nil {
		m["autoprovision"] = aws.ToBool(v)
	}

	if v := config.Driver; v != nil {
		m["driver"] = aws.ToString(v)
	}

	if config.DriverOpts != nil {
		m["driver_opts"] = flex.FlattenStringValueMap(config.DriverOpts)
	}

	if v := config.Labels; v != nil {
		m["labels"] = flex.FlattenStringValueMap(v)
	}

	items = append(items, m)
	return items
}

func flattenEFSVolumeConfiguration(config *awstypes.EFSVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.FileSystemId; v != nil {
			m[names.AttrFileSystemID] = aws.ToString(v)
		}

		if v := config.RootDirectory; v != nil {
			m["root_directory"] = aws.ToString(v)
		}
		m["transit_encryption"] = string(config.TransitEncryption)

		if v := config.TransitEncryptionPort; v != nil {
			m["transit_encryption_port"] = int(aws.ToInt32(v))
		}

		if v := config.AuthorizationConfig; v != nil {
			m["authorization_config"] = flattenEFSVolumeAuthorizationConfig(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenEFSVolumeAuthorizationConfig(config *awstypes.EFSAuthorizationConfig) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.AccessPointId; v != nil {
			m["access_point_id"] = aws.ToString(v)
		}
		m["iam"] = string(config.Iam)
	}

	items = append(items, m)
	return items
}

func flattenFSxWinVolumeConfiguration(config *awstypes.FSxWindowsFileServerVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.FileSystemId; v != nil {
			m[names.AttrFileSystemID] = aws.ToString(v)
		}

		if v := config.RootDirectory; v != nil {
			m["root_directory"] = aws.ToString(v)
		}

		if v := config.AuthorizationConfig; v != nil {
			m["authorization_config"] = flattenFSxWinVolumeAuthorizationConfig(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenFSxWinVolumeAuthorizationConfig(config *awstypes.FSxWindowsFileServerAuthorizationConfig) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.CredentialsParameter; v != nil {
			m["credentials_parameter"] = aws.ToString(v)
		}
		if v := config.Domain; v != nil {
			m[names.AttrDomain] = aws.ToString(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenContainerDefinitions(apiObjects []awstypes.ContainerDefinition) (string, error) {
	return tfjson.EncodeToString(apiObjects)
}

func expandContainerDefinitions(tfString string) ([]awstypes.ContainerDefinition, error) {
	var apiObjects []awstypes.ContainerDefinition

	if err := tfjson.DecodeFromString(tfString, &apiObjects); err != nil {
		return nil, err
	}

	for i, apiObject := range apiObjects {
		if itypes.IsZero(&apiObject) {
			return nil, fmt.Errorf("invalid container definition supplied at index (%d)", i)
		}
	}

	return apiObjects, nil
}

func expandTaskDefinitionEphemeralStorage(config []interface{}) *awstypes.EphemeralStorage {
	configMap := config[0].(map[string]interface{})

	es := &awstypes.EphemeralStorage{
		SizeInGiB: int32(configMap["size_in_gib"].(int)),
	}

	return es
}

func flattenTaskDefinitionEphemeralStorage(pc *awstypes.EphemeralStorage) []map[string]interface{} {
	if pc == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["size_in_gib"] = pc.SizeInGiB

	return []map[string]interface{}{m}
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
