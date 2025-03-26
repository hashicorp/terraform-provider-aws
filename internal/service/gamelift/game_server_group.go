// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_gamelift_game_server_group", name="Game Server Group")
// @Tags(identifierAttribute="arn")
func resourceGameServerGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGameServerGroupCreate,
		ReadWithoutTimeout:   resourceGameServerGroupRead,
		UpdateWithoutTimeout: resourceGameServerGroupUpdate,
		DeleteWithoutTimeout: resourceGameServerGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_policy": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"estimated_instance_warmup": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"target_tracking_configuration": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_value": {
										Type:         schema.TypeFloat,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.FloatAtLeast(0),
									},
								},
							},
						},
					},
				},
			},
			"balancing_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.BalancingStrategy](),
			},
			"game_server_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"game_server_protection_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.GameServerProtectionPolicy](),
			},
			"instance_definition": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 2,
				MaxItems: 20,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrInstanceType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.GameServerGroupInstanceType](),
						},
						"weighted_capacity": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 3),
						},
					},
				},
			},
			names.AttrLaunchTemplate: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.name"},
							ValidateFunc:  verify.ValidLaunchTemplateID,
						},
						names.AttrName: {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.id"},
							ValidateFunc:  verify.ValidLaunchTemplateName,
						},
						names.AttrVersion: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
					},
				},
			},
			"max_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"min_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_subnets": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 20,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(15, 24),
				},
			},
		},
	}
}

func resourceGameServerGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	name := d.Get("game_server_group_name").(string)
	input := &gamelift.CreateGameServerGroupInput{
		GameServerGroupName: aws.String(name),
		InstanceDefinitions: expandInstanceDefinitions(d.Get("instance_definition").(*schema.Set).List()),
		LaunchTemplate:      expandLaunchTemplateSpecification(d.Get(names.AttrLaunchTemplate).([]any)[0].(map[string]any)),
		MaxSize:             aws.Int32(int32(d.Get("max_size").(int))),
		MinSize:             aws.Int32(int32(d.Get("min_size").(int))),
		RoleArn:             aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk("auto_scaling_policy"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.AutoScalingPolicy = expandGameServerGroupAutoScalingPolicy(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("balancing_strategy"); ok {
		input.BalancingStrategy = awstypes.BalancingStrategy(v.(string))
	}

	if v, ok := d.GetOk("game_server_protection_policy"); ok {
		input.GameServerProtectionPolicy = awstypes.GameServerProtectionPolicy(v.(string))
	}

	if v, ok := d.GetOk("vpc_subnets"); ok && v.(*schema.Set).Len() > 0 {
		input.VpcSubnets = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidRequestException](ctx, propagationTimeout, func() (any, error) {
		return conn.CreateGameServerGroup(ctx, input)
	}, "GameLift is not authorized to perform")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Game Server Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*gamelift.CreateGameServerGroupOutput).GameServerGroup.GameServerGroupName))

	if _, err := waitGameServerGroupActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Game Server Group (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGameServerGroupRead(ctx, d, meta)...)
}

func resourceGameServerGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)
	autoscalingConn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	gameServerGroup, err := findGameServerGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Game Server Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Game Server Group (%s): %s", d.Id(), err)
	}

	if asgArnParts := strings.Split(aws.ToString(gameServerGroup.AutoScalingGroupArn), "/"); len(asgArnParts) == 2 {
		asgName := asgArnParts[1]
		asg, err := tfautoscaling.FindGroupByName(ctx, autoscalingConn, asgName)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Group (%s): %s", asgName, err)
		}

		asgPolicy, err := tfautoscaling.FindScalingPolicyByTwoPartKey(ctx, autoscalingConn, asgName, d.Id())

		switch {
		case tfresource.NotFound(err):
		case err != nil:
			return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Policy (%s/%s): %s", asgName, d.Id(), err)
		}

		if asgPolicy != nil {
			if err := d.Set("auto_scaling_policy", []any{flattenGameServerGroupAutoScalingPolicy(asgPolicy)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting auto_scaling_policy: %s", err)
			}
		} else {
			d.Set("auto_scaling_policy", nil)
		}

		if err := d.Set(names.AttrLaunchTemplate, flattenAutoScalingLaunchTemplateSpecification(asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting launch_template: %s", err)
		}
		d.Set("max_size", asg.MaxSize)
		d.Set("min_size", asg.MinSize)
	}

	d.Set(names.AttrARN, gameServerGroup.GameServerGroupArn)
	d.Set("auto_scaling_group_arn", gameServerGroup.AutoScalingGroupArn)
	d.Set("balancing_strategy", gameServerGroup.BalancingStrategy)
	d.Set("game_server_group_name", gameServerGroup.GameServerGroupName)
	d.Set("game_server_protection_policy", gameServerGroup.GameServerProtectionPolicy)
	if err := d.Set("instance_definition", flattenInstanceDefinitions(gameServerGroup.InstanceDefinitions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_definition: %s", err)
	}
	d.Set(names.AttrRoleARN, gameServerGroup.RoleArn)

	return diags
}

func resourceGameServerGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChanges("balancing_strategy", "game_server_protection_policy", "instance_definition", names.AttrRoleARN) {
		input := &gamelift.UpdateGameServerGroupInput{
			GameServerGroupName: aws.String(d.Id()),
			InstanceDefinitions: expandInstanceDefinitions(d.Get("instance_definition").(*schema.Set).List()),
			RoleArn:             aws.String(d.Get(names.AttrRoleARN).(string)),
		}

		if v, ok := d.GetOk("balancing_strategy"); ok {
			input.BalancingStrategy = awstypes.BalancingStrategy(v.(string))
		}

		if v, ok := d.GetOk("game_server_protection_policy"); ok {
			input.GameServerProtectionPolicy = awstypes.GameServerProtectionPolicy(v.(string))
		}

		_, err := conn.UpdateGameServerGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Game Server Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGameServerGroupRead(ctx, d, meta)...)
}

func resourceGameServerGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Deleting GameLift Game Server Group: %s", d.Id())
	msg := fmt.Sprintf("Cannot delete game server group %s: ", d.Id())
	const (
		timeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidRequestException](ctx, timeout, func() (any, error) {
		return conn.DeleteGameServerGroup(ctx, &gamelift.DeleteGameServerGroupInput{
			GameServerGroupName: aws.String(d.Id()),
		})
	}, msg)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Game Server Group (%s): %s", d.Id(), err)
	}

	if _, err := waitGameServerGroupTerminated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Game Server Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findGameServerGroupByName(ctx context.Context, conn *gamelift.Client, name string) (*awstypes.GameServerGroup, error) {
	input := &gamelift.DescribeGameServerGroupInput{
		GameServerGroupName: aws.String(name),
	}

	output, err := conn.DescribeGameServerGroup(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.GameServerGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.GameServerGroup, nil
}

func statusGameServerGroup(ctx context.Context, conn *gamelift.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findGameServerGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitGameServerGroupActive(ctx context.Context, conn *gamelift.Client, name string, timeout time.Duration) (*awstypes.GameServerGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.GameServerGroupStatusNew,
			awstypes.GameServerGroupStatusActivating,
		),
		Target:  enum.Slice(awstypes.GameServerGroupStatusActive),
		Refresh: statusGameServerGroup(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GameServerGroup); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitGameServerGroupTerminated(ctx context.Context, conn *gamelift.Client, name string, timeout time.Duration) (*awstypes.GameServerGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.GameServerGroupStatusDeleteScheduled,
			awstypes.GameServerGroupStatusDeleting,
		),
		Target:  []string{},
		Refresh: statusGameServerGroup(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GameServerGroup); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func expandGameServerGroupAutoScalingPolicy(tfMap map[string]any) *awstypes.GameServerGroupAutoScalingPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.GameServerGroupAutoScalingPolicy{
		TargetTrackingConfiguration: expandTargetTrackingConfiguration(tfMap["target_tracking_configuration"].([]any)[0].(map[string]any)),
	}

	if v, ok := tfMap["estimated_instance_warmup"].(int); ok && v != 0 {
		apiObject.EstimatedInstanceWarmup = aws.Int32(int32(v))
	}

	return apiObject
}

func expandInstanceDefinition(tfMap map[string]any) *awstypes.InstanceDefinition {
	apiObject := &awstypes.InstanceDefinition{
		InstanceType: awstypes.GameServerGroupInstanceType(tfMap[names.AttrInstanceType].(string)),
	}

	if v, ok := tfMap["weighted_capacity"].(string); ok && v != "" {
		apiObject.WeightedCapacity = aws.String(v)
	}

	return apiObject
}

func expandInstanceDefinitions(tfList []any) []awstypes.InstanceDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.InstanceDefinition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, *expandInstanceDefinition(tfMap))
	}

	return apiObjects
}

func expandLaunchTemplateSpecification(tfMap map[string]any) *awstypes.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateSpecification{}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandTargetTrackingConfiguration(tfMap map[string]any) *awstypes.TargetTrackingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TargetTrackingConfiguration{
		TargetValue: aws.Float64(tfMap["target_value"].(float64)),
	}

	return apiObject
}

func flattenGameServerGroupAutoScalingPolicy(apiObject *autoscalingtypes.ScalingPolicy) map[string]any {
	tfMap := map[string]any{
		"target_tracking_configuration": []any{flattenTargetTrackingConfiguration(apiObject.TargetTrackingConfiguration)},
	}

	if v := apiObject.EstimatedInstanceWarmup; v != nil {
		tfMap["estimated_instance_warmup"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenInstanceDefinition(apiObject *awstypes.InstanceDefinition) map[string]any {
	tfMap := map[string]any{
		names.AttrInstanceType: string(apiObject.InstanceType),
	}

	if v := apiObject.WeightedCapacity; v != nil {
		tfMap["weighted_capacity"] = aws.ToString(v)
	}

	return tfMap
}

func flattenAutoScalingLaunchTemplateSpecification(apiObject *autoscalingtypes.LaunchTemplateSpecification) []map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrID:   aws.ToString(apiObject.LaunchTemplateId),
		names.AttrName: aws.ToString(apiObject.LaunchTemplateName),
	}

	// version is returned only if it was previously set
	if apiObject.Version != nil {
		tfMap[names.AttrVersion] = aws.ToString(apiObject.Version)
	} else {
		tfMap[names.AttrVersion] = nil
	}

	return []map[string]any{tfMap}
}

func flattenInstanceDefinitions(apiObjects []awstypes.InstanceDefinition) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenInstanceDefinition(&apiObject))
	}

	return tfList
}

func flattenTargetTrackingConfiguration(apiObject *autoscalingtypes.TargetTrackingConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"target_value": aws.ToFloat64(apiObject.TargetValue),
	}

	return tfMap
}
