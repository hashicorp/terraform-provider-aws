// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package gamelift

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const containerGroupDefinitionIDPartCount = 2

// @SDKResource("aws_gamelift_container_group_definition", name="Container Group Definition")
// @Tags(identifierAttribute="arn")
func resourceContainerGroupDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContainerGroupDefinitionCreate,
		ReadWithoutTimeout:   resourceContainerGroupDefinitionRead,
		UpdateWithoutTimeout: resourceContainerGroupDefinitionUpdate,
		DeleteWithoutTimeout: resourceContainerGroupDefinitionDelete,
		CustomizeDiff:        resourceContainerGroupDefinitionCustomizeDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"container_group_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ContainerGroupTypeGameServer,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ContainerGroupType](),
			},
			"operating_system": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ContainerOperatingSystemAmazonLinux2023,
				ValidateDiagFunc: enum.Validate[awstypes.ContainerOperatingSystem](),
			},
			"total_memory_limit_mib": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"total_vcpu_limit": {
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatAtLeast(0),
			},
			"game_server_container_definition": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"image_uri": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"port_configuration": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_port_ranges": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"from_port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},
												"to_port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},
												names.AttrProtocol: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.IpProtocol](),
												},
											},
										},
									},
								},
							},
						},
						"server_sdk_version": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
						"depends_on": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									names.AttrCondition: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ContainerDependencyCondition](),
									},
								},
							},
						},
						"environment_override": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"mount_points": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"instance_path": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"access_level": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ContainerMountPointAccessLevel](),
									},
									"container_path": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
								},
							},
						},
						"resolved_image_digest": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"support_container_definitions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"image_uri": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"port_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_port_ranges": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"from_port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},
												"to_port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},
												names.AttrProtocol: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.IpProtocol](),
												},
											},
										},
									},
								},
							},
						},
						"depends_on": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									names.AttrCondition: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ContainerDependencyCondition](),
									},
								},
							},
						},
						"environment_override": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"essential": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						names.AttrHealthCheck: {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"command": {
										Type:     schema.TypeList,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrInterval: {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"retries": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"start_period": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									names.AttrTimeout: {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"memory_hard_limit_mib": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"mount_points": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"instance_path": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"access_level": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ContainerMountPointAccessLevel](),
									},
									"container_path": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
								},
							},
						},
						"vcpu": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ValidateFunc: validation.FloatAtLeast(0),
						},
						"resolved_image_digest": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"version_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"version_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceContainerGroupDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	name := d.Get(names.AttrName).(string)
	containerGroupType := awstypes.ContainerGroupType(d.Get("container_group_type").(string))
	gameServerDefinition := expandGameServerContainerDefinitionInput(d.Get("game_server_container_definition").([]any))

	if containerGroupType == awstypes.ContainerGroupTypeGameServer && gameServerDefinition == nil {
		return sdkdiag.AppendErrorf(diags, "game_server_container_definition is required when container_group_type is GAME_SERVER")
	}
	if containerGroupType == awstypes.ContainerGroupTypePerInstance && gameServerDefinition != nil {
		return sdkdiag.AppendErrorf(diags, "game_server_container_definition must be omitted when container_group_type is PER_INSTANCE")
	}

	input := &gamelift.CreateContainerGroupDefinitionInput{
		Name:                          aws.String(name),
		ContainerGroupType:            containerGroupType,
		GameServerContainerDefinition: gameServerDefinition,
		OperatingSystem:               awstypes.ContainerOperatingSystem(d.Get("operating_system").(string)),
		TotalMemoryLimitMebibytes:     aws.Int32(int32(d.Get("total_memory_limit_mib").(int))),
		TotalVcpuLimit:                aws.Float64(d.Get("total_vcpu_limit").(float64)),
		SupportContainerDefinitions:   expandSupportContainerDefinitionInputs(d.Get("support_container_definitions").([]any)),
		Tags:                          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("version_description"); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	output, err := conn.CreateContainerGroupDefinition(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Container Group Definition (%s): %s", name, err)
	}

	containerGroup := output.ContainerGroupDefinition
	if containerGroup == nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Container Group Definition (%s): empty response", name)
	}

	if err := setContainerGroupDefinitionID(d, containerGroup); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting GameLift Container Group Definition (%s) ID: %s", name, err)
	}

	name, version, err := parseContainerGroupDefinitionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing GameLift Container Group Definition ID (%s): %s", d.Id(), err)
	}

	if _, err := waitContainerGroupDefinitionReady(ctx, conn, name, version, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Container Group Definition (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceContainerGroupDefinitionRead(ctx, d, meta)...)
}

func resourceContainerGroupDefinitionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	name, version, err := parseContainerGroupDefinitionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing GameLift Container Group Definition ID (%s): %s", d.Id(), err)
	}

	containerGroup, err := findContainerGroupDefinition(ctx, conn, name, version)
	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] GameLift Container Group Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Container Group Definition (%s): %s", d.Id(), err)
	}

	if err := setContainerGroupDefinitionID(d, containerGroup); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting GameLift Container Group Definition (%s) ID: %s", d.Id(), err)
	}

	d.Set(names.AttrARN, containerGroup.ContainerGroupDefinitionArn)
	d.Set(names.AttrName, containerGroup.Name)
	d.Set("container_group_type", containerGroup.ContainerGroupType)
	d.Set("operating_system", containerGroup.OperatingSystem)
	d.Set("total_memory_limit_mib", aws.ToInt32(containerGroup.TotalMemoryLimitMebibytes))
	d.Set("total_vcpu_limit", aws.ToFloat64(containerGroup.TotalVcpuLimit))
	d.Set("version_description", containerGroup.VersionDescription)
	d.Set("version_number", aws.ToInt32(containerGroup.VersionNumber))
	d.Set(names.AttrStatus, containerGroup.Status)
	d.Set("status_reason", containerGroup.StatusReason)
	if containerGroup.CreationTime != nil {
		d.Set(names.AttrCreationTime, containerGroup.CreationTime.Format(time.RFC3339))
	}

	if err := d.Set("game_server_container_definition", flattenGameServerContainerDefinition(containerGroup.GameServerContainerDefinition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting game_server_container_definition: %s", err)
	}
	if err := d.Set("support_container_definitions", flattenSupportContainerDefinitions(containerGroup.SupportContainerDefinitions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting support_container_definitions: %s", err)
	}

	return diags
}

func resourceContainerGroupDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		name := d.Get(names.AttrName).(string)
		containerGroupType := awstypes.ContainerGroupType(d.Get("container_group_type").(string))
		gameServerDefinition := expandGameServerContainerDefinitionInput(d.Get("game_server_container_definition").([]any))

		if containerGroupType == awstypes.ContainerGroupTypeGameServer && gameServerDefinition == nil {
			return sdkdiag.AppendErrorf(diags, "game_server_container_definition is required when container_group_type is GAME_SERVER")
		}
		if containerGroupType == awstypes.ContainerGroupTypePerInstance && gameServerDefinition != nil {
			return sdkdiag.AppendErrorf(diags, "game_server_container_definition must be omitted when container_group_type is PER_INSTANCE")
		}

		input := &gamelift.UpdateContainerGroupDefinitionInput{
			Name:                          aws.String(name),
			GameServerContainerDefinition: gameServerDefinition,
			OperatingSystem:               awstypes.ContainerOperatingSystem(d.Get("operating_system").(string)),
			TotalMemoryLimitMebibytes:     aws.Int32(int32(d.Get("total_memory_limit_mib").(int))),
			TotalVcpuLimit:                aws.Float64(d.Get("total_vcpu_limit").(float64)),
			SupportContainerDefinitions:   expandSupportContainerDefinitionInputs(d.Get("support_container_definitions").([]any)),
		}

		if v, ok := d.GetOk("version_description"); ok {
			input.VersionDescription = aws.String(v.(string))
		}

		if v, ok := d.GetOk("version_number"); ok {
			input.SourceVersionNumber = aws.Int32(int32(v.(int)))
		}

		output, err := conn.UpdateContainerGroupDefinition(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Container Group Definition (%s): %s", d.Id(), err)
		}

		containerGroup := output.ContainerGroupDefinition
		if containerGroup == nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Container Group Definition (%s): empty response", d.Id())
		}

		if err := setContainerGroupDefinitionID(d, containerGroup); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting GameLift Container Group Definition (%s) ID: %s", d.Id(), err)
		}

		name, version, err := parseContainerGroupDefinitionID(d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing GameLift Container Group Definition ID (%s): %s", d.Id(), err)
		}

		if _, err := waitContainerGroupDefinitionReady(ctx, conn, name, version, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for GameLift Container Group Definition (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceContainerGroupDefinitionRead(ctx, d, meta)...)
}

func resourceContainerGroupDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	name, version, err := parseContainerGroupDefinitionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing GameLift Container Group Definition ID (%s): %s", d.Id(), err)
	}

	input := &gamelift.DeleteContainerGroupDefinitionInput{
		Name: aws.String(name),
	}
	if version != nil {
		input.VersionNumber = aws.Int32(*version)
	}

	log.Printf("[INFO] Deleting GameLift Container Group Definition: %s", d.Id())
	_, err = conn.DeleteContainerGroupDefinition(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Container Group Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func findContainerGroupDefinition(ctx context.Context, conn *gamelift.Client, name string, version *int32) (*awstypes.ContainerGroupDefinition, error) {
	input := &gamelift.DescribeContainerGroupDefinitionInput{
		Name: aws.String(name),
	}
	if version != nil {
		input.VersionNumber = aws.Int32(*version)
	}

	output, err := conn.DescribeContainerGroupDefinition(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &sdkretry.NotFoundError{LastError: err, LastRequest: input}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.ContainerGroupDefinition == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ContainerGroupDefinition, nil
}

func statusContainerGroupDefinition(ctx context.Context, conn *gamelift.Client, name string, version *int32) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findContainerGroupDefinition(ctx, conn, name, version)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitContainerGroupDefinitionReady(ctx context.Context, conn *gamelift.Client, name string, version *int32, timeout time.Duration) (*awstypes.ContainerGroupDefinition, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ContainerGroupDefinitionStatusCopying,
		),
		Target: enum.Slice(
			awstypes.ContainerGroupDefinitionStatusReady,
		),
		Refresh: statusContainerGroupDefinition(ctx, conn, name, version),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ContainerGroupDefinition); ok {
		if reason := aws.ToString(output.StatusReason); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func resourceContainerGroupDefinitionCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta any) error {
	if d.Id() != "" && needsContainerGroupDefinitionUpdate(d) && d.Get(names.AttrARN).(string) != "" {
		d.SetNewComputed(names.AttrARN)
		d.SetNewComputed(names.AttrID)
		d.SetNewComputed("version_number")
		d.SetNewComputed(names.AttrStatus)
		d.SetNewComputed("status_reason")
		d.SetNewComputed(names.AttrCreationTime)
	}

	return nil
}

func needsContainerGroupDefinitionUpdate(d *schema.ResourceDiff) bool {
	return d.HasChange("operating_system") ||
		d.HasChange("total_memory_limit_mib") ||
		d.HasChange("total_vcpu_limit") ||
		d.HasChange("game_server_container_definition") ||
		d.HasChange("support_container_definitions") ||
		d.HasChange("version_description")
}

func setContainerGroupDefinitionID(d *schema.ResourceData, definition *awstypes.ContainerGroupDefinition) error {
	if definition == nil {
		return fmt.Errorf("container group definition is nil")
	}

	name := aws.ToString(definition.Name)
	version := aws.ToInt32(definition.VersionNumber)
	id, err := flex.FlattenResourceId([]string{name, strconv.Itoa(int(version))}, containerGroupDefinitionIDPartCount, false)
	if err != nil {
		return err
	}
	d.SetId(id)

	return nil
}

func parseContainerGroupDefinitionID(id string) (string, *int32, error) {
	if strings.HasPrefix(id, "arn:") {
		parsed, err := arn.Parse(id)
		if err != nil {
			return "", nil, err
		}

		resource := strings.TrimPrefix(parsed.Resource, "containergroupdefinition/")
		parts := strings.Split(resource, ":")
		if len(parts) != 2 {
			return "", nil, fmt.Errorf("unexpected resource format: %s", parsed.Resource)
		}
		version, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", nil, err
		}
		return parts[0], aws.Int32(int32(version)), nil
	}

	parts, err := flex.ExpandResourceId(id, containerGroupDefinitionIDPartCount, false)
	if err != nil {
		return "", nil, err
	}
	version, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", nil, err
	}
	return parts[0], aws.Int32(int32(version)), nil
}

func expandGameServerContainerDefinitionInput(tfList []any) *awstypes.GameServerContainerDefinitionInput {
	if len(tfList) < 1 {
		return nil
	}
	m := tfList[0].(map[string]any)

	apiObject := &awstypes.GameServerContainerDefinitionInput{
		ContainerName:     aws.String(m["container_name"].(string)),
		ImageUri:          aws.String(m["image_uri"].(string)),
		PortConfiguration: expandContainerPortConfiguration(m["port_configuration"].([]any)),
		ServerSdkVersion:  aws.String(m["server_sdk_version"].(string)),
	}

	if v, ok := m["depends_on"]; ok && len(v.([]any)) > 0 {
		apiObject.DependsOn = expandContainerDependencies(v.([]any))
	}
	if v, ok := m["environment_override"]; ok && len(v.([]any)) > 0 {
		apiObject.EnvironmentOverride = expandContainerEnvironmentOverrides(v.([]any))
	}
	if v, ok := m["mount_points"]; ok && len(v.([]any)) > 0 {
		apiObject.MountPoints = expandContainerMountPoints(v.([]any))
	}

	return apiObject
}

func expandSupportContainerDefinitionInputs(tfList []any) []awstypes.SupportContainerDefinitionInput {
	if len(tfList) < 1 {
		return nil
	}

	apiObjects := make([]awstypes.SupportContainerDefinitionInput, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		m := tfMapRaw.(map[string]any)
		apiObject := awstypes.SupportContainerDefinitionInput{
			ContainerName: aws.String(m["container_name"].(string)),
			ImageUri:      aws.String(m["image_uri"].(string)),
		}

		if v, ok := m["depends_on"]; ok && len(v.([]any)) > 0 {
			apiObject.DependsOn = expandContainerDependencies(v.([]any))
		}
		if v, ok := m["environment_override"]; ok && len(v.([]any)) > 0 {
			apiObject.EnvironmentOverride = expandContainerEnvironmentOverrides(v.([]any))
		}
		if v, ok := m["essential"]; ok {
			apiObject.Essential = aws.Bool(v.(bool))
		}
		if v, ok := m[names.AttrHealthCheck]; ok && len(v.([]any)) > 0 {
			apiObject.HealthCheck = expandContainerHealthCheck(v.([]any))
		}
		if v, ok := m["memory_hard_limit_mib"]; ok && v.(int) > 0 {
			apiObject.MemoryHardLimitMebibytes = aws.Int32(int32(v.(int)))
		}
		if v, ok := m["mount_points"]; ok && len(v.([]any)) > 0 {
			apiObject.MountPoints = expandContainerMountPoints(v.([]any))
		}
		if v, ok := m["port_configuration"]; ok && len(v.([]any)) > 0 {
			apiObject.PortConfiguration = expandContainerPortConfiguration(v.([]any))
		}
		if v, ok := m["vcpu"]; ok && v.(float64) > 0 {
			apiObject.Vcpu = aws.Float64(v.(float64))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandContainerDependencies(tfList []any) []awstypes.ContainerDependency {
	if len(tfList) < 1 {
		return nil
	}

	apiObjects := make([]awstypes.ContainerDependency, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		m := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, awstypes.ContainerDependency{
			ContainerName: aws.String(m["container_name"].(string)),
			Condition:     awstypes.ContainerDependencyCondition(m[names.AttrCondition].(string)),
		})
	}

	return apiObjects
}

func expandContainerEnvironmentOverrides(tfList []any) []awstypes.ContainerEnvironment {
	if len(tfList) < 1 {
		return nil
	}

	apiObjects := make([]awstypes.ContainerEnvironment, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		m := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, awstypes.ContainerEnvironment{
			Name:  aws.String(m[names.AttrName].(string)),
			Value: aws.String(m[names.AttrValue].(string)),
		})
	}

	return apiObjects
}

func expandContainerMountPoints(tfList []any) []awstypes.ContainerMountPoint {
	if len(tfList) < 1 {
		return nil
	}

	apiObjects := make([]awstypes.ContainerMountPoint, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		m := tfMapRaw.(map[string]any)
		apiObject := awstypes.ContainerMountPoint{
			InstancePath: aws.String(m["instance_path"].(string)),
		}

		if v, ok := m["access_level"]; ok && v.(string) != "" {
			apiObject.AccessLevel = awstypes.ContainerMountPointAccessLevel(v.(string))
		}
		if v, ok := m["container_path"]; ok && v.(string) != "" {
			apiObject.ContainerPath = aws.String(v.(string))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandContainerPortConfiguration(tfList []any) *awstypes.ContainerPortConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	m := tfList[0].(map[string]any)
	apiObject := &awstypes.ContainerPortConfiguration{}
	if v, ok := m["container_port_ranges"]; ok {
		apiObject.ContainerPortRanges = expandContainerPortRanges(v.([]any))
	}

	return apiObject
}

func expandContainerPortRanges(tfList []any) []awstypes.ContainerPortRange {
	if len(tfList) < 1 {
		return nil
	}

	apiObjects := make([]awstypes.ContainerPortRange, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		m := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, awstypes.ContainerPortRange{
			FromPort: aws.Int32(int32(m["from_port"].(int))),
			ToPort:   aws.Int32(int32(m["to_port"].(int))),
			Protocol: awstypes.IpProtocol(m[names.AttrProtocol].(string)),
		})
	}

	return apiObjects
}

func expandContainerHealthCheck(tfList []any) *awstypes.ContainerHealthCheck {
	if len(tfList) < 1 {
		return nil
	}

	m := tfList[0].(map[string]any)
	apiObject := &awstypes.ContainerHealthCheck{
		Command: flex.ExpandStringValueList(m["command"].([]any)),
	}

	if v, ok := m[names.AttrInterval]; ok {
		apiObject.Interval = aws.Int32(int32(v.(int)))
	}
	if v, ok := m["retries"]; ok {
		apiObject.Retries = aws.Int32(int32(v.(int)))
	}
	if v, ok := m["start_period"]; ok {
		apiObject.StartPeriod = aws.Int32(int32(v.(int)))
	}
	if v, ok := m[names.AttrTimeout]; ok {
		apiObject.Timeout = aws.Int32(int32(v.(int)))
	}

	return apiObject
}

func flattenGameServerContainerDefinition(apiObject *awstypes.GameServerContainerDefinition) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"container_name":     aws.ToString(apiObject.ContainerName),
		"image_uri":          aws.ToString(apiObject.ImageUri),
		"server_sdk_version": aws.ToString(apiObject.ServerSdkVersion),
	}
	if apiObject.ResolvedImageDigest != nil {
		m["resolved_image_digest"] = aws.ToString(apiObject.ResolvedImageDigest)
	}

	if apiObject.PortConfiguration != nil {
		m["port_configuration"] = flattenContainerPortConfiguration(apiObject.PortConfiguration)
	}
	if len(apiObject.DependsOn) > 0 {
		m["depends_on"] = flattenContainerDependencies(apiObject.DependsOn)
	}
	if len(apiObject.EnvironmentOverride) > 0 {
		m["environment_override"] = flattenContainerEnvironmentOverrides(apiObject.EnvironmentOverride)
	}
	if len(apiObject.MountPoints) > 0 {
		m["mount_points"] = flattenContainerMountPoints(apiObject.MountPoints)
	}

	return []any{m}
}

func flattenSupportContainerDefinitions(apiObjects []awstypes.SupportContainerDefinition) []any {
	if len(apiObjects) < 1 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		m := map[string]any{
			"container_name": aws.ToString(apiObject.ContainerName),
			"image_uri":      aws.ToString(apiObject.ImageUri),
		}
		if apiObject.ResolvedImageDigest != nil {
			m["resolved_image_digest"] = aws.ToString(apiObject.ResolvedImageDigest)
		}

		if apiObject.Essential != nil {
			m["essential"] = aws.ToBool(apiObject.Essential)
		}
		if apiObject.MemoryHardLimitMebibytes != nil {
			m["memory_hard_limit_mib"] = aws.ToInt32(apiObject.MemoryHardLimitMebibytes)
		}
		if apiObject.Vcpu != nil {
			m["vcpu"] = aws.ToFloat64(apiObject.Vcpu)
		}

		if len(apiObject.DependsOn) > 0 {
			m["depends_on"] = flattenContainerDependencies(apiObject.DependsOn)
		}
		if len(apiObject.EnvironmentOverride) > 0 {
			m["environment_override"] = flattenContainerEnvironmentOverrides(apiObject.EnvironmentOverride)
		}
		if apiObject.HealthCheck != nil {
			m[names.AttrHealthCheck] = flattenContainerHealthCheck(apiObject.HealthCheck)
		}
		if apiObject.PortConfiguration != nil {
			m["port_configuration"] = flattenContainerPortConfiguration(apiObject.PortConfiguration)
		}
		if len(apiObject.MountPoints) > 0 {
			m["mount_points"] = flattenContainerMountPoints(apiObject.MountPoints)
		}

		tfList = append(tfList, m)
	}

	return tfList
}

func flattenContainerDependencies(apiObjects []awstypes.ContainerDependency) []any {
	if len(apiObjects) < 1 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		m := map[string]any{
			"container_name":    aws.ToString(apiObject.ContainerName),
			names.AttrCondition: apiObject.Condition,
		}
		tfList = append(tfList, m)
	}

	return tfList
}

func flattenContainerEnvironmentOverrides(apiObjects []awstypes.ContainerEnvironment) []any {
	if len(apiObjects) < 1 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		m := map[string]any{
			names.AttrName:  aws.ToString(apiObject.Name),
			names.AttrValue: aws.ToString(apiObject.Value),
		}
		tfList = append(tfList, m)
	}

	return tfList
}

func flattenContainerMountPoints(apiObjects []awstypes.ContainerMountPoint) []any {
	if len(apiObjects) < 1 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		m := map[string]any{
			"instance_path": aws.ToString(apiObject.InstancePath),
		}
		if apiObject.AccessLevel != "" {
			m["access_level"] = apiObject.AccessLevel
		}
		if apiObject.ContainerPath != nil {
			m["container_path"] = aws.ToString(apiObject.ContainerPath)
		}
		tfList = append(tfList, m)
	}

	return tfList
}

func flattenContainerPortConfiguration(apiObject *awstypes.ContainerPortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{}
	if len(apiObject.ContainerPortRanges) > 0 {
		m["container_port_ranges"] = flattenContainerPortRanges(apiObject.ContainerPortRanges)
	}

	return []any{m}
}

func flattenContainerPortRanges(apiObjects []awstypes.ContainerPortRange) []any {
	if len(apiObjects) < 1 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		m := map[string]any{
			"from_port":        aws.ToInt32(apiObject.FromPort),
			"to_port":          aws.ToInt32(apiObject.ToPort),
			names.AttrProtocol: apiObject.Protocol,
		}
		tfList = append(tfList, m)
	}

	return tfList
}

func flattenContainerHealthCheck(apiObject *awstypes.ContainerHealthCheck) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"command": flex.FlattenStringValueList(apiObject.Command),
	}
	if apiObject.Interval != nil {
		m[names.AttrInterval] = aws.ToInt32(apiObject.Interval)
	}
	if apiObject.Retries != nil {
		m["retries"] = aws.ToInt32(apiObject.Retries)
	}
	if apiObject.StartPeriod != nil {
		m["start_period"] = aws.ToInt32(apiObject.StartPeriod)
	}
	if apiObject.Timeout != nil {
		m[names.AttrTimeout] = aws.ToInt32(apiObject.Timeout)
	}

	return []any{m}
}
