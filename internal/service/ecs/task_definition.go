// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_task_definition", name="Task Definition")
// @Tags(identifierAttribute="arn")
func ResourceTaskDefinition() *schema.Resource {
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
					orderedCDs, _ := expandContainerDefinitions(v.(string))
					containerDefinitions(orderedCDs).OrderContainers()
					containerDefinitions(orderedCDs).OrderEnvironmentVariables()
					containerDefinitions(orderedCDs).OrderSecrets()
					unnormalizedJson, _ := flattenContainerDefinitions(orderedCDs)
					json, _ := structure.NormalizeJsonString(unnormalizedJson)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					networkMode, ok := d.GetOk("network_mode")
					isAWSVPC := ok && networkMode.(string) == ecs.NetworkModeAwsvpc
					equal, _ := ContainerDefinitionsAreEquivalent(old, new, isAWSVPC)
					return equal
				},
				ValidateFunc: ValidTaskDefinitionContainerDefinitions,
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ecs.IpcMode_Values(), false),
			},
			"memory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ecs.NetworkMode_Values(), false),
			},
			"pid_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ecs.PidMode_Values(), false),
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
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ecs.TaskDefinitionPlacementConstraintType_Values(), false),
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
							Type:         schema.TypeString,
							Default:      ecs.ProxyConfigurationTypeAppmesh,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ecs.ProxyConfigurationType_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ecs.CPUArchitecture_Values(), false),
						},
						"operating_system_family": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ecs.OSFamily_Values(), false),
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
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(ecs.Scope_Values(), false),
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
													Type:         schema.TypeString,
													ForceNew:     true,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(ecs.EFSAuthorizationConfigIAM_Values(), false),
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
										Type:         schema.TypeString,
										ForceNew:     true,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ecs.EFSTransitEncryption_Values(), false),
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
						"configure_at_launch": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
				Set: resourceTaskDefinitionVolumeHash,
			},
		},
	}
}

func ValidTaskDefinitionContainerDefinitions(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandContainerDefinitions(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("ECS Task Definition container_definitions is invalid: %s", err))
	}
	return
}

func resourceTaskDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	rawDefinitions := d.Get("container_definitions").(string)
	definitions, err := expandContainerDefinitions(rawDefinitions)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Task Definition (%s): %s", d.Get(names.AttrFamily).(string), err)
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
		input.IpcMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("memory"); ok {
		input.Memory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_mode"); ok {
		input.NetworkMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pid_mode"); ok {
		input.PidMode = aws.String(v.(string))
	}

	if constraints := d.Get("placement_constraints").(*schema.Set).List(); len(constraints) > 0 {
		cons, err := expandTaskDefinitionPlacementConstraints(constraints)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating ECS Task Definition (%s): %s", d.Get(names.AttrFamily).(string), err)
		}
		input.PlacementConstraints = cons
	}

	if proxyConfigs := d.Get("proxy_configuration").([]interface{}); len(proxyConfigs) > 0 {
		input.ProxyConfiguration = expandTaskDefinitionProxyConfiguration(proxyConfigs)
	}

	if v, ok := d.GetOk("requires_compatibilities"); ok && v.(*schema.Set).Len() > 0 {
		input.RequiresCompatibilities = flex.ExpandStringSet(v.(*schema.Set))
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

	output, err := conn.RegisterTaskDefinitionWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.RegisterTaskDefinitionWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Task Definition (%s): %s", d.Get(names.AttrFamily).(string), err)
	}

	taskDefinition := *output.TaskDefinition // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment // false positive

	d.SetId(aws.StringValue(taskDefinition.Family))
	d.Set(names.AttrARN, taskDefinition.TaskDefinitionArn)
	d.Set("arn_without_revision", StripRevision(aws.StringValue(taskDefinition.TaskDefinitionArn)))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.StringValue(taskDefinition.TaskDefinitionArn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	trackedTaskDefinition := d.Get(names.AttrARN).(string)
	if _, ok := d.GetOk("track_latest"); ok {
		trackedTaskDefinition = d.Get(names.AttrFamily).(string)
	}

	input := ecs.DescribeTaskDefinitionInput{
		Include:        aws.StringSlice([]string{ecs.TaskDefinitionFieldTags}),
		TaskDefinition: aws.String(trackedTaskDefinition),
	}

	out, err := conn.DescribeTaskDefinitionWithContext(ctx, &input)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		log.Printf("[WARN] ECS tagging failed describing Task Definition (%s) with tags: %s; retrying without tags", d.Id(), err)

		input.Include = nil
		out, err = conn.DescribeTaskDefinitionWithContext(ctx, &input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Received task definition %s, status:%s\n %s", aws.StringValue(out.TaskDefinition.Family),
		aws.StringValue(out.TaskDefinition.Status), out)

	taskDefinition := out.TaskDefinition

	if aws.StringValue(taskDefinition.Status) == ecs.TaskDefinitionStatusInactive {
		log.Printf("[DEBUG] Removing ECS task definition %s because it's INACTIVE", aws.StringValue(out.TaskDefinition.Family))
		d.SetId("")
		return diags
	}

	d.SetId(aws.StringValue(taskDefinition.Family))
	d.Set(names.AttrARN, taskDefinition.TaskDefinitionArn)
	d.Set("arn_without_revision", StripRevision(aws.StringValue(taskDefinition.TaskDefinitionArn)))
	d.Set(names.AttrFamily, taskDefinition.Family)
	d.Set("revision", taskDefinition.Revision)
	d.Set("track_latest", d.Get("track_latest"))

	// Sort the lists of environment variables as they come in, so we won't get spurious reorderings in plans
	// (diff is suppressed if the environment variables haven't changed, but they still show in the plan if
	// some other property changes).
	containerDefinitions(taskDefinition.ContainerDefinitions).OrderContainers()
	containerDefinitions(taskDefinition.ContainerDefinitions).OrderEnvironmentVariables()
	containerDefinitions(taskDefinition.ContainerDefinitions).OrderSecrets()

	defs, err := flattenContainerDefinitions(taskDefinition.ContainerDefinitions)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition (%s): %s", d.Id(), err)
	}
	err = d.Set("container_definitions", defs)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition (%s): %s", d.Id(), err)
	}

	d.Set("task_role_arn", taskDefinition.TaskRoleArn)
	d.Set(names.AttrExecutionRoleARN, taskDefinition.ExecutionRoleArn)
	d.Set("cpu", taskDefinition.Cpu)
	d.Set("memory", taskDefinition.Memory)
	d.Set("network_mode", taskDefinition.NetworkMode)
	d.Set("ipc_mode", taskDefinition.IpcMode)
	d.Set("pid_mode", taskDefinition.PidMode)

	if err := d.Set("volume", flattenVolumes(taskDefinition.Volumes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting volume: %s", err)
	}

	if err := d.Set("inference_accelerator", flattenInferenceAccelerators(taskDefinition.InferenceAccelerators)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting inference accelerators: %s", err)
	}

	if err := d.Set("placement_constraints", flattenPlacementConstraints(taskDefinition.PlacementConstraints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting placement_constraints: %s", err)
	}

	if err := d.Set("requires_compatibilities", flex.FlattenStringList(taskDefinition.RequiresCompatibilities)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting requires_compatibilities: %s", err)
	}

	if err := d.Set("runtime_platform", flattenRuntimePlatform(taskDefinition.RuntimePlatform)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting runtime_platform: %s", err)
	}

	if err := d.Set("proxy_configuration", flattenProxyConfiguration(taskDefinition.ProxyConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting proxy_configuration: %s", err)
	}

	if err := d.Set("ephemeral_storage", flattenTaskDefinitionEphemeralStorage(taskDefinition.EphemeralStorage)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ephemeral_storage: %s", err)
	}

	setTagsOut(ctx, out.Tags)

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
		log.Printf("[DEBUG] Retaining ECS Task Definition Revision %q", d.Id())
		return diags
	}

	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	_, err := conn.DeregisterTaskDefinitionWithContext(ctx, &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get(names.AttrARN).(string)),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Task Definition (%s): %s", d.Id(), err)
	}

	return diags
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

func flattenPlacementConstraints(pcs []*ecs.TaskDefinitionPlacementConstraint) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c[names.AttrType] = aws.StringValue(pc.Type)
		c[names.AttrExpression] = aws.StringValue(pc.Expression)
		results = append(results, c)
	}
	return results
}

func flattenRuntimePlatform(rp *ecs.RuntimePlatform) []map[string]interface{} {
	if rp == nil {
		return nil
	}

	os := aws.StringValue(rp.OperatingSystemFamily)
	cpu := aws.StringValue(rp.CpuArchitecture)

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

func flattenProxyConfiguration(pc *ecs.ProxyConfiguration) []map[string]interface{} {
	if pc == nil {
		return nil
	}

	meshProperties := make(map[string]string)
	if pc.Properties != nil {
		for _, prop := range pc.Properties {
			meshProperties[aws.StringValue(prop.Name)] = aws.StringValue(prop.Value)
		}
	}

	config := make(map[string]interface{})
	config["container_name"] = aws.StringValue(pc.ContainerName)
	config[names.AttrType] = aws.StringValue(pc.Type)
	config[names.AttrProperties] = meshProperties

	return []map[string]interface{}{
		config,
	}
}

func flattenInferenceAccelerators(list []*ecs.InferenceAccelerator) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, iAcc := range list {
		l := map[string]interface{}{
			names.AttrDeviceName: aws.StringValue(iAcc.DeviceName),
			"device_type":        aws.StringValue(iAcc.DeviceType),
		}

		result = append(result, l)
	}
	return result
}

func expandInferenceAccelerators(configured []interface{}) []*ecs.InferenceAccelerator {
	iAccs := make([]*ecs.InferenceAccelerator, 0, len(configured))
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})
		l := &ecs.InferenceAccelerator{
			DeviceName: aws.String(data[names.AttrDeviceName].(string)),
			DeviceType: aws.String(data["device_type"].(string)),
		}
		iAccs = append(iAccs, l)
	}

	return iAccs
}

func expandTaskDefinitionPlacementConstraints(constraints []interface{}) ([]*ecs.TaskDefinitionPlacementConstraint, error) {
	var pc []*ecs.TaskDefinitionPlacementConstraint
	for _, raw := range constraints {
		p := raw.(map[string]interface{})
		t := p[names.AttrType].(string)
		e := p[names.AttrExpression].(string)
		if err := validPlacementConstraint(t, e); err != nil {
			return nil, err
		}
		pc = append(pc, &ecs.TaskDefinitionPlacementConstraint{
			Type:       aws.String(t),
			Expression: aws.String(e),
		})
	}

	return pc, nil
}

func expandTaskDefinitionRuntimePlatformConfiguration(runtimePlatformConfig []interface{}) *ecs.RuntimePlatform {
	config := runtimePlatformConfig[0]

	configMap := config.(map[string]interface{})
	ecsProxyConfig := &ecs.RuntimePlatform{}

	os := configMap["operating_system_family"].(string)
	if os != "" {
		ecsProxyConfig.OperatingSystemFamily = aws.String(os)
	}

	osFamily := configMap["cpu_architecture"].(string)
	if osFamily != "" {
		ecsProxyConfig.CpuArchitecture = aws.String(osFamily)
	}

	return ecsProxyConfig
}

func expandTaskDefinitionProxyConfiguration(proxyConfigs []interface{}) *ecs.ProxyConfiguration {
	proxyConfig := proxyConfigs[0]
	configMap := proxyConfig.(map[string]interface{})

	rawProperties := configMap[names.AttrProperties].(map[string]interface{})

	properties := make([]*ecs.KeyValuePair, len(rawProperties))
	i := 0
	for name, value := range rawProperties {
		properties[i] = &ecs.KeyValuePair{
			Name:  aws.String(name),
			Value: aws.String(value.(string)),
		}
		i++
	}

	ecsProxyConfig := &ecs.ProxyConfiguration{
		ContainerName: aws.String(configMap["container_name"].(string)),
		Type:          aws.String(configMap[names.AttrType].(string)),
		Properties:    properties,
	}

	return ecsProxyConfig
}

func expandVolumes(configured []interface{}) []*ecs.Volume {
	volumes := make([]*ecs.Volume, 0, len(configured))

	// Loop over our configured volumes and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &ecs.Volume{
			Name: aws.String(data[names.AttrName].(string)),
		}

		hostPath := data["host_path"].(string)
		if hostPath != "" {
			l.Host = &ecs.HostVolumeProperties{
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

func expandVolumesDockerVolume(configList []interface{}) *ecs.DockerVolumeConfiguration {
	config := configList[0].(map[string]interface{})
	dockerVol := &ecs.DockerVolumeConfiguration{}

	if v, ok := config[names.AttrScope].(string); ok && v != "" {
		dockerVol.Scope = aws.String(v)
	}

	if v, ok := config["autoprovision"]; ok && v != "" {
		if dockerVol.Scope == nil || aws.StringValue(dockerVol.Scope) != ecs.ScopeTask || v.(bool) {
			dockerVol.Autoprovision = aws.Bool(v.(bool))
		}
	}

	if v, ok := config["driver"].(string); ok && v != "" {
		dockerVol.Driver = aws.String(v)
	}

	if v, ok := config["driver_opts"].(map[string]interface{}); ok && len(v) > 0 {
		dockerVol.DriverOpts = flex.ExpandStringMap(v)
	}

	if v, ok := config["labels"].(map[string]interface{}); ok && len(v) > 0 {
		dockerVol.Labels = flex.ExpandStringMap(v)
	}

	return dockerVol
}

func expandVolumesEFSVolume(efsConfig []interface{}) *ecs.EFSVolumeConfiguration {
	config := efsConfig[0].(map[string]interface{})
	efsVol := &ecs.EFSVolumeConfiguration{}

	if v, ok := config[names.AttrFileSystemID].(string); ok && v != "" {
		efsVol.FileSystemId = aws.String(v)
	}

	if v, ok := config["root_directory"].(string); ok && v != "" {
		efsVol.RootDirectory = aws.String(v)
	}
	if v, ok := config["transit_encryption"].(string); ok && v != "" {
		efsVol.TransitEncryption = aws.String(v)
	}

	if v, ok := config["transit_encryption_port"].(int); ok && v > 0 {
		efsVol.TransitEncryptionPort = aws.Int64(int64(v))
	}
	if v, ok := config["authorization_config"].([]interface{}); ok && len(v) > 0 {
		efsVol.AuthorizationConfig = expandVolumesEFSVolumeAuthorizationConfig(v)
	}

	return efsVol
}

func expandVolumesEFSVolumeAuthorizationConfig(efsConfig []interface{}) *ecs.EFSAuthorizationConfig {
	authconfig := efsConfig[0].(map[string]interface{})
	auth := &ecs.EFSAuthorizationConfig{}

	if v, ok := authconfig["access_point_id"].(string); ok && v != "" {
		auth.AccessPointId = aws.String(v)
	}

	if v, ok := authconfig["iam"].(string); ok && v != "" {
		auth.Iam = aws.String(v)
	}

	return auth
}

func expandVolumesFSxWinVolume(fsxWinConfig []interface{}) *ecs.FSxWindowsFileServerVolumeConfiguration {
	config := fsxWinConfig[0].(map[string]interface{})
	fsxVol := &ecs.FSxWindowsFileServerVolumeConfiguration{}

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

func expandVolumesFSxWinVolumeAuthorizationConfig(config []interface{}) *ecs.FSxWindowsFileServerAuthorizationConfig {
	authconfig := config[0].(map[string]interface{})
	auth := &ecs.FSxWindowsFileServerAuthorizationConfig{}

	if v, ok := authconfig["credentials_parameter"].(string); ok && v != "" {
		auth.CredentialsParameter = aws.String(v)
	}

	if v, ok := authconfig[names.AttrDomain].(string); ok && v != "" {
		auth.Domain = aws.String(v)
	}

	return auth
}

func flattenVolumes(list []*ecs.Volume) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, volume := range list {
		l := map[string]interface{}{
			names.AttrName: aws.StringValue(volume.Name),
		}

		if volume.Host != nil && volume.Host.SourcePath != nil {
			l["host_path"] = aws.StringValue(volume.Host.SourcePath)
		}

		if volume.ConfiguredAtLaunch != nil {
			l["configure_at_launch"] = aws.BoolValue(volume.ConfiguredAtLaunch)
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

func flattenDockerVolumeConfiguration(config *ecs.DockerVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})

	if v := config.Scope; v != nil {
		m[names.AttrScope] = aws.StringValue(v)
	}

	if v := config.Autoprovision; v != nil {
		m["autoprovision"] = aws.BoolValue(v)
	}

	if v := config.Driver; v != nil {
		m["driver"] = aws.StringValue(v)
	}

	if config.DriverOpts != nil {
		m["driver_opts"] = flex.FlattenStringMap(config.DriverOpts)
	}

	if v := config.Labels; v != nil {
		m["labels"] = flex.FlattenStringMap(v)
	}

	items = append(items, m)
	return items
}

func flattenEFSVolumeConfiguration(config *ecs.EFSVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.FileSystemId; v != nil {
			m[names.AttrFileSystemID] = aws.StringValue(v)
		}

		if v := config.RootDirectory; v != nil {
			m["root_directory"] = aws.StringValue(v)
		}
		if v := config.TransitEncryption; v != nil {
			m["transit_encryption"] = aws.StringValue(v)
		}

		if v := config.TransitEncryptionPort; v != nil {
			m["transit_encryption_port"] = int(aws.Int64Value(v))
		}

		if v := config.AuthorizationConfig; v != nil {
			m["authorization_config"] = flattenEFSVolumeAuthorizationConfig(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenEFSVolumeAuthorizationConfig(config *ecs.EFSAuthorizationConfig) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.AccessPointId; v != nil {
			m["access_point_id"] = aws.StringValue(v)
		}
		if v := config.Iam; v != nil {
			m["iam"] = aws.StringValue(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenFSxWinVolumeConfiguration(config *ecs.FSxWindowsFileServerVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.FileSystemId; v != nil {
			m[names.AttrFileSystemID] = aws.StringValue(v)
		}

		if v := config.RootDirectory; v != nil {
			m["root_directory"] = aws.StringValue(v)
		}

		if v := config.AuthorizationConfig; v != nil {
			m["authorization_config"] = flattenFSxWinVolumeAuthorizationConfig(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenFSxWinVolumeAuthorizationConfig(config *ecs.FSxWindowsFileServerAuthorizationConfig) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.CredentialsParameter; v != nil {
			m["credentials_parameter"] = aws.StringValue(v)
		}
		if v := config.Domain; v != nil {
			m[names.AttrDomain] = aws.StringValue(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenContainerDefinitions(definitions []*ecs.ContainerDefinition) (string, error) {
	b, err := jsonutil.BuildJSON(definitions)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func expandContainerDefinitions(rawDefinitions string) ([]*ecs.ContainerDefinition, error) {
	var definitions []*ecs.ContainerDefinition

	err := json.Unmarshal([]byte(rawDefinitions), &definitions)
	if err != nil {
		return nil, fmt.Errorf("decoding JSON: %s", err)
	}

	for i, c := range definitions {
		if c == nil {
			return nil, fmt.Errorf("invalid container definition supplied at index (%d)", i)
		}
	}

	return definitions, nil
}

func expandTaskDefinitionEphemeralStorage(config []interface{}) *ecs.EphemeralStorage {
	configMap := config[0].(map[string]interface{})

	es := &ecs.EphemeralStorage{
		SizeInGiB: aws.Int64(int64(configMap["size_in_gib"].(int))),
	}

	return es
}

func flattenTaskDefinitionEphemeralStorage(pc *ecs.EphemeralStorage) []map[string]interface{} {
	if pc == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["size_in_gib"] = aws.Int64Value(pc.SizeInGiB)

	return []map[string]interface{}{m}
}

// StripRevision strips the trailing revision number from a task definition ARN
//
// Invalid ARNs will return an empty string. ARNs with an unexpected number of
// separators in the resource section are returned unmodified.
func StripRevision(s string) string {
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
