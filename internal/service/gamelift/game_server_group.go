package gamelift

import ( // nosemgrep:ci.aws-sdk-go-multiple-service-imports
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	gameServerGroupCreatedDefaultTimeout = 10 * time.Minute
	gameServerGroupDeletedDefaultTimeout = 30 * time.Minute
)

func ResourceGameServerGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGameServerGroupCreate,
		ReadWithoutTimeout:   resourceGameServerGroupRead,
		UpdateWithoutTimeout: resourceGameServerGroupUpdate,
		DeleteWithoutTimeout: resourceGameServerGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(gameServerGroupCreatedDefaultTimeout),
			Delete: schema.DefaultTimeout(gameServerGroupDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(gamelift.BalancingStrategy_Values(), false),
			},
			"game_server_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"game_server_protection_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(gamelift.GameServerProtectionPolicy_Values(), false),
			},
			"instance_definition": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 2,
				MaxItems: 20,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(gamelift.GameServerGroupInstanceType_Values(), false),
						},
						"weighted_capacity": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 3),
						},
					},
				},
			},
			"launch_template": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.name"},
							ValidateFunc:  verify.ValidLaunchTemplateID,
						},
						"name": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.id"},
							ValidateFunc:  verify.ValidLaunchTemplateName,
						},
						"version": {
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
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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

func resourceGameServerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &gamelift.CreateGameServerGroupInput{
		GameServerGroupName: aws.String(d.Get("game_server_group_name").(string)),
		InstanceDefinitions: expandInstanceDefinitions(d.Get("instance_definition").(*schema.Set).List()),
		LaunchTemplate:      expandLaunchTemplateSpecification(d.Get("launch_template").([]interface{})[0].(map[string]interface{})),
		MaxSize:             aws.Int64(int64(d.Get("max_size").(int))),
		MinSize:             aws.Int64(int64(d.Get("min_size").(int))),
		RoleArn:             aws.String(d.Get("role_arn").(string)),
		Tags:                Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("auto_scaling_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AutoScalingPolicy = expandGameServerGroupAutoScalingPolicy(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("balancing_strategy"); ok {
		input.BalancingStrategy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("game_server_protection_policy"); ok {
		input.GameServerProtectionPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_subnets"); ok && v.(*schema.Set).Len() > 0 {
		input.VpcSubnets = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[INFO] Creating GameLift Game Server Group: %s", input)
	var out *gamelift.CreateGameServerGroupOutput
	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateGameServerGroupWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "GameLift is not authorized to perform") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		out, err = conn.CreateGameServerGroupWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Game Server Group (%s): %s", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(out.GameServerGroup.GameServerGroupName))

	if output, err := waitGameServerGroupActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Game Server Group (%s) to become active (%s): %s", d.Id(), *output.StatusReason, err)
	}

	return append(diags, resourceGameServerGroupRead(ctx, d, meta)...)
}

func resourceGameServerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn()
	autoscalingConn := meta.(*conns.AWSClient).AutoScalingConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	gameServerGroupName := d.Id()

	log.Printf("[INFO] Describing GameLift Game Server Group: %s", gameServerGroupName)
	gameServerGroup, err := FindGameServerGroupByName(ctx, conn, gameServerGroupName)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Game Server Group (%s) not found, removing from state", gameServerGroupName)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Game Server Group (%s): %s", gameServerGroupName, err)
	}

	autoScalingGroupName := strings.Split(aws.StringValue(gameServerGroup.AutoScalingGroupArn), "/")[1]
	autoScalingGroupOutput, err := autoscalingConn.DescribeAutoScalingGroupsWithContext(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(autoScalingGroupName)},
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Game Server Group (%s): reading AutoScaling Group: %s", gameServerGroupName, err)
	}
	if autoScalingGroupOutput == nil || len(autoScalingGroupOutput.AutoScalingGroups) == 0 {
		return sdkdiag.AppendErrorf(diags, "describing Auto Scaling Group (%s): not found", autoScalingGroupName)
	}
	autoScalingGroup := autoScalingGroupOutput.AutoScalingGroups[0]

	describePoliciesOutput, err := autoscalingConn.DescribePoliciesWithContext(ctx, &autoscaling.DescribePoliciesInput{
		AutoScalingGroupName: aws.String(autoScalingGroupName),
		PolicyNames:          []*string{aws.String(gameServerGroupName)},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Auto Scaling Group Policies (%s): %s", autoScalingGroupName, err)
	}

	arn := aws.StringValue(gameServerGroup.GameServerGroupArn)
	d.Set("arn", arn)
	d.Set("auto_scaling_group_arn", gameServerGroup.AutoScalingGroupArn)
	d.Set("balancing_strategy", gameServerGroup.BalancingStrategy)
	d.Set("game_server_group_name", gameServerGroupName)
	d.Set("game_server_protection_policy", gameServerGroup.GameServerProtectionPolicy)
	d.Set("max_size", autoScalingGroup.MaxSize)
	d.Set("min_size", autoScalingGroup.MinSize)
	d.Set("role_arn", gameServerGroup.RoleArn)

	if len(describePoliciesOutput.ScalingPolicies) == 1 {
		if err := d.Set("auto_scaling_policy", []interface{}{flattenGameServerGroupAutoScalingPolicy(describePoliciesOutput.ScalingPolicies[0])}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting auto_scaling_policy: %s", err)
		}
	} else {
		d.Set("auto_scaling_policy", nil)
	}

	if err := d.Set("instance_definition", flattenInstanceDefinitions(gameServerGroup.InstanceDefinitions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_definition: %s", err)
	}

	if err := d.Set("launch_template", flattenAutoScalingLaunchTemplateSpecification(autoScalingGroup.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting launch_template: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, fmt.Sprintf("Resource %s is not in a taggable state", d.Id())) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Game Lift Game Server Group (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceGameServerGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn()

	log.Printf("[INFO] Updating GameLift Game Server Group: %s", d.Id())

	if d.HasChanges("balancing_strategy", "game_server_protection_policy", "instance_definition", "role_arn") {
		input := gamelift.UpdateGameServerGroupInput{
			GameServerGroupName: aws.String(d.Id()),
			InstanceDefinitions: expandInstanceDefinitions(d.Get("instance_definition").(*schema.Set).List()),
			RoleArn:             aws.String(d.Get("role_arn").(string)),
		}

		if v, ok := d.GetOk("balancing_strategy"); ok {
			input.BalancingStrategy = aws.String(v.(string))
		}

		if v, ok := d.GetOk("game_server_protection_policy"); ok {
			input.GameServerProtectionPolicy = aws.String(v.(string))
		}

		_, err := conn.UpdateGameServerGroupWithContext(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Game Server Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Game Lift Game Server Group (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceGameServerGroupRead(ctx, d, meta)...)
}

func resourceGameServerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn()

	log.Printf("[INFO] Deleting GameLift Game Server Group: %s", d.Id())
	input := &gamelift.DeleteGameServerGroupInput{
		GameServerGroupName: aws.String(d.Id()),
	}
	err := resource.RetryContext(ctx, gameServerGroupDeletedDefaultTimeout, func() *resource.RetryError {
		_, err := conn.DeleteGameServerGroupWithContext(ctx, input)
		if err != nil {
			msg := fmt.Sprintf("Cannot delete game server group %s: %s", d.Id(), err)
			if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, msg) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteGameServerGroupWithContext(ctx, input)
	}
	if err != nil {
		if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting GameLift game server group: %s", err)
	}

	if err := waitGameServerGroupTerminated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Game Server Group (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func expandGameServerGroupAutoScalingPolicy(tfMap map[string]interface{}) *gamelift.GameServerGroupAutoScalingPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &gamelift.GameServerGroupAutoScalingPolicy{
		TargetTrackingConfiguration: expandTargetTrackingConfiguration(tfMap["target_tracking_configuration"].([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := tfMap["estimated_instance_warmup"].(int); ok && v != 0 {
		apiObject.EstimatedInstanceWarmup = aws.Int64(int64(v))
	}

	return apiObject
}

func expandInstanceDefinition(tfMap map[string]interface{}) *gamelift.InstanceDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &gamelift.InstanceDefinition{
		InstanceType: aws.String(tfMap["instance_type"].(string)),
	}

	if v, ok := tfMap["weighted_capacity"].(string); ok && v != "" {
		apiObject.WeightedCapacity = aws.String(v)
	}

	return apiObject
}

func expandInstanceDefinitions(tfList []interface{}) []*gamelift.InstanceDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*gamelift.InstanceDefinition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandInstanceDefinition(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateSpecification(tfMap map[string]interface{}) *gamelift.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &gamelift.LaunchTemplateSpecification{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandTargetTrackingConfiguration(tfMap map[string]interface{}) *gamelift.TargetTrackingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &gamelift.TargetTrackingConfiguration{
		TargetValue: aws.Float64(tfMap["target_value"].(float64)),
	}

	return apiObject
}

func flattenGameServerGroupAutoScalingPolicy(apiObject *autoscaling.ScalingPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"target_tracking_configuration": []interface{}{flattenTargetTrackingConfiguration(apiObject.TargetTrackingConfiguration)},
	}

	if v := apiObject.EstimatedInstanceWarmup; v != nil {
		tfMap["estimated_instance_warmup"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenInstanceDefinition(apiObject *gamelift.InstanceDefinition) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"instance_type": aws.StringValue(apiObject.InstanceType),
	}

	if v := apiObject.WeightedCapacity; v != nil {
		tfMap["weighted_capacity"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenAutoScalingLaunchTemplateSpecification(apiObject *autoscaling.LaunchTemplateSpecification) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"id":   aws.StringValue(apiObject.LaunchTemplateId),
		"name": aws.StringValue(apiObject.LaunchTemplateName),
	}

	// version is returned only if it was previously set
	if apiObject.Version != nil {
		tfMap["version"] = aws.StringValue(apiObject.Version)
	} else {
		tfMap["version"] = nil
	}

	return []map[string]interface{}{tfMap}
}

func flattenInstanceDefinitions(apiObjects []*gamelift.InstanceDefinition) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenInstanceDefinition(apiObject))
	}

	return tfList
}

func flattenTargetTrackingConfiguration(apiObject *autoscaling.TargetTrackingConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"target_value": aws.Float64Value(apiObject.TargetValue),
	}

	return tfMap
}
