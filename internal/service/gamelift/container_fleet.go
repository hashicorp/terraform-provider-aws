// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package gamelift

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_gamelift_container_fleet", name="Container Fleet")
// @Tags(identifierAttribute="arn")
func resourceContainerFleet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContainerFleetCreate,
		ReadWithoutTimeout:   resourceContainerFleetRead,
		UpdateWithoutTimeout: resourceContainerFleetUpdate,
		DeleteWithoutTimeout: resourceContainerFleetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fleet_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"billing_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ContainerFleetBillingTypeOnDemand,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ContainerFleetBillingType](),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"game_server_container_group_definition_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"game_server_container_groups_per_instance": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"game_session_creation_limit_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"new_game_sessions_per_creator": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"policy_period_in_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"instance_connection_port_range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
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
					},
				},
			},
			"instance_inbound_permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"ip_range": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidCIDRNetworkAddress,
						},
						"protocol": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.IpProtocol](),
						},
						"to_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},
			"instance_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"locations": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
					},
				},
			},
			"log_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_destination": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LogDestination](),
						},
						"log_group_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"s3_bucket_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"metric_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"new_game_session_protection_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ProtectionPolicyNoProtection,
				ValidateDiagFunc: enum.Validate[awstypes.ProtectionPolicy](),
			},
			"per_instance_container_group_definition_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"remove_per_instance_container_group_definition": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"deployment_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"impairment_strategy": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DeploymentImpairmentStrategy](),
						},
						"minimum_healthy_percentage": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100),
						},
						"protection_strategy": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DeploymentProtectionStrategy](),
						},
					},
				},
			},
			"deployment_details": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"latest_deployment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"game_server_container_group_definition_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"per_instance_container_group_definition_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maximum_game_server_container_groups_per_instance": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"location_attributes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceContainerFleetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	input := &gamelift.CreateContainerFleetInput{
		FleetRoleArn: aws.String(d.Get("fleet_role_arn").(string)),
		BillingType:  awstypes.ContainerFleetBillingType(d.Get("billing_type").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("game_server_container_group_definition_name"); ok {
		input.GameServerContainerGroupDefinitionName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("game_server_container_groups_per_instance"); ok {
		input.GameServerContainerGroupsPerInstance = aws.Int32(int32(v.(int)))
	}
	if v, ok := d.GetOk("game_session_creation_limit_policy"); ok {
		input.GameSessionCreationLimitPolicy = expandGameSessionCreationLimitPolicy(v.([]any))
	}
	if v, ok := d.GetOk("instance_connection_port_range"); ok {
		input.InstanceConnectionPortRange = expandConnectionPortRange(v.([]any))
	}
	if v, ok := d.GetOk("instance_inbound_permission"); ok {
		input.InstanceInboundPermissions = expandIPPermissions(v.(*schema.Set))
	}
	if v, ok := d.GetOk("instance_type"); ok {
		input.InstanceType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("locations"); ok {
		input.Locations = expandLocationConfigurations(v.([]any))
	}
	if v, ok := d.GetOk("log_configuration"); ok {
		input.LogConfiguration = expandLogConfiguration(v.([]any))
	}
	if v, ok := d.GetOk("metric_groups"); ok {
		input.MetricGroups = flex.ExpandStringValueList(v.([]any))
	}
	if v, ok := d.GetOk("new_game_session_protection_policy"); ok {
		input.NewGameSessionProtectionPolicy = awstypes.ProtectionPolicy(v.(string))
	}
	if v, ok := d.GetOk("per_instance_container_group_definition_name"); ok {
		input.PerInstanceContainerGroupDefinitionName = aws.String(v.(string))
	}

	output, err := conn.CreateContainerFleet(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Container Fleet: %s", err)
	}

	fleet := output.ContainerFleet
	if fleet == nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Container Fleet: empty response")
	}

	d.SetId(aws.ToString(fleet.FleetId))

	return append(diags, resourceContainerFleetRead(ctx, d, meta)...)
}

func resourceContainerFleetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	fleet, err := findContainerFleetByID(ctx, conn, d.Id())
	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] GameLift Container Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Container Fleet (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, fleet.FleetArn)
	d.Set("fleet_role_arn", aws.ToString(fleet.FleetRoleArn))
	d.Set("billing_type", fleet.BillingType)
	d.Set("description", aws.ToString(fleet.Description))
	d.Set("game_server_container_group_definition_name", aws.ToString(fleet.GameServerContainerGroupDefinitionName))
	d.Set("game_server_container_groups_per_instance", aws.ToInt32(fleet.GameServerContainerGroupsPerInstance))
	d.Set("instance_type", aws.ToString(fleet.InstanceType))
	d.Set("metric_groups", flex.FlattenStringValueList(fleet.MetricGroups))
	d.Set("new_game_session_protection_policy", fleet.NewGameSessionProtectionPolicy)
	d.Set("per_instance_container_group_definition_name", aws.ToString(fleet.PerInstanceContainerGroupDefinitionName))
	d.Set("status", fleet.Status)
	d.Set("game_server_container_group_definition_arn", aws.ToString(fleet.GameServerContainerGroupDefinitionArn))
	d.Set("per_instance_container_group_definition_arn", aws.ToString(fleet.PerInstanceContainerGroupDefinitionArn))
	d.Set("maximum_game_server_container_groups_per_instance", aws.ToInt32(fleet.MaximumGameServerContainerGroupsPerInstance))

	if err := d.Set("instance_connection_port_range", flattenConnectionPortRange(fleet.InstanceConnectionPortRange)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_connection_port_range: %s", err)
	}
	if err := d.Set("instance_inbound_permission", flattenIPPermissions(fleet.InstanceInboundPermissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_inbound_permission: %s", err)
	}
	if err := d.Set("log_configuration", flattenLogConfiguration(fleet.LogConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_configuration: %s", err)
	}
	if err := d.Set("game_session_creation_limit_policy", flattenGameSessionCreationLimitPolicy(fleet.GameSessionCreationLimitPolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting game_session_creation_limit_policy: %s", err)
	}
	if err := d.Set("location_attributes", flattenContainerFleetLocationAttributes(fleet.LocationAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting location_attributes: %s", err)
	}
	if err := d.Set("locations", flattenLocationAttributesToLocations(fleet.LocationAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting locations: %s", err)
	}
	if err := d.Set("deployment_details", flattenDeploymentDetails(fleet.DeploymentDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployment_details: %s", err)
	}

	return diags
}

func resourceContainerFleetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		if d.Get("remove_per_instance_container_group_definition").(bool) && d.Get("per_instance_container_group_definition_name").(string) != "" {
			return sdkdiag.AppendErrorf(diags, "remove_per_instance_container_group_definition cannot be set with per_instance_container_group_definition_name")
		}

		input := &gamelift.UpdateContainerFleetInput{
			FleetId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}
		if v, ok := d.GetOk("game_server_container_group_definition_name"); ok {
			input.GameServerContainerGroupDefinitionName = aws.String(v.(string))
		}
		if v, ok := d.GetOk("game_server_container_groups_per_instance"); ok {
			input.GameServerContainerGroupsPerInstance = aws.Int32(int32(v.(int)))
		}
		if v, ok := d.GetOk("game_session_creation_limit_policy"); ok {
			input.GameSessionCreationLimitPolicy = expandGameSessionCreationLimitPolicy(v.([]any))
		}
		if v, ok := d.GetOk("instance_connection_port_range"); ok {
			input.InstanceConnectionPortRange = expandConnectionPortRange(v.([]any))
		}
		if d.HasChange("instance_inbound_permission") {
			oldRaw, newRaw := d.GetChange("instance_inbound_permission")
			oldSet := oldRaw.(*schema.Set)
			newSet := newRaw.(*schema.Set)
			if added := expandIPPermissions(newSet.Difference(oldSet)); len(added) > 0 {
				input.InstanceInboundPermissionAuthorizations = added
			}
			if removed := expandIPPermissions(oldSet.Difference(newSet)); len(removed) > 0 {
				input.InstanceInboundPermissionRevocations = removed
			}
		}
		if v, ok := d.GetOk("log_configuration"); ok {
			input.LogConfiguration = expandLogConfiguration(v.([]any))
		}
		if v, ok := d.GetOk("metric_groups"); ok {
			input.MetricGroups = flex.ExpandStringValueList(v.([]any))
		}
		if v, ok := d.GetOk("new_game_session_protection_policy"); ok {
			input.NewGameSessionProtectionPolicy = awstypes.ProtectionPolicy(v.(string))
		}
		if v, ok := d.GetOk("per_instance_container_group_definition_name"); ok {
			input.PerInstanceContainerGroupDefinitionName = aws.String(v.(string))
		}
		if v, ok := d.GetOk("deployment_configuration"); ok {
			input.DeploymentConfiguration = expandDeploymentConfiguration(v.([]any))
		}
		if d.Get("remove_per_instance_container_group_definition").(bool) {
			input.RemoveAttributes = []awstypes.ContainerFleetRemoveAttribute{
				awstypes.ContainerFleetRemoveAttributePerInstanceContainerGroupDefinition,
			}
		}

		if _, err := conn.UpdateContainerFleet(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Container Fleet (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceContainerFleetRead(ctx, d, meta)...)
}

func resourceContainerFleetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Deleting GameLift Container Fleet: %s", d.Id())
	_, err := conn.DeleteContainerFleet(ctx, &gamelift.DeleteContainerFleetInput{
		FleetId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Container Fleet (%s): %s", d.Id(), err)
	}

	return diags
}

func findContainerFleetByID(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.ContainerFleet, error) {
	input := &gamelift.DescribeContainerFleetInput{
		FleetId: aws.String(id),
	}

	output, err := conn.DescribeContainerFleet(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &sdkretry.NotFoundError{LastError: err, LastRequest: input}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.ContainerFleet == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ContainerFleet, nil
}

func expandConnectionPortRange(tfList []any) *awstypes.ConnectionPortRange {
	if len(tfList) < 1 {
		return nil
	}

	m := tfList[0].(map[string]any)
	return &awstypes.ConnectionPortRange{
		FromPort: aws.Int32(int32(m["from_port"].(int))),
		ToPort:   aws.Int32(int32(m["to_port"].(int))),
	}
}

func flattenConnectionPortRange(apiObject *awstypes.ConnectionPortRange) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"from_port": aws.ToInt32(apiObject.FromPort),
		"to_port":   aws.ToInt32(apiObject.ToPort),
	}

	return []any{m}
}

func expandLocationConfigurations(tfList []any) []awstypes.LocationConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	apiObjects := make([]awstypes.LocationConfiguration, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		m := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, awstypes.LocationConfiguration{
			Location: aws.String(m["location"].(string)),
		})
	}

	return apiObjects
}

func expandLogConfiguration(tfList []any) *awstypes.LogConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	m := tfList[0].(map[string]any)
	apiObject := &awstypes.LogConfiguration{}

	if v, ok := m["log_destination"]; ok && v.(string) != "" {
		apiObject.LogDestination = awstypes.LogDestination(v.(string))
	}
	if v, ok := m["log_group_arn"]; ok && v.(string) != "" {
		apiObject.LogGroupArn = aws.String(v.(string))
	}
	if v, ok := m["s3_bucket_name"]; ok && v.(string) != "" {
		apiObject.S3BucketName = aws.String(v.(string))
	}

	return apiObject
}

func flattenLogConfiguration(apiObject *awstypes.LogConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{}
	if apiObject.LogDestination != "" {
		m["log_destination"] = apiObject.LogDestination
	}
	if apiObject.LogGroupArn != nil {
		m["log_group_arn"] = aws.ToString(apiObject.LogGroupArn)
	}
	if apiObject.S3BucketName != nil {
		m["s3_bucket_name"] = aws.ToString(apiObject.S3BucketName)
	}

	if len(m) == 0 {
		return nil
	}

	return []any{m}
}

func expandDeploymentConfiguration(tfList []any) *awstypes.DeploymentConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	m := tfList[0].(map[string]any)
	apiObject := &awstypes.DeploymentConfiguration{}

	if v, ok := m["impairment_strategy"]; ok && v.(string) != "" {
		apiObject.ImpairmentStrategy = awstypes.DeploymentImpairmentStrategy(v.(string))
	}
	if v, ok := m["minimum_healthy_percentage"]; ok {
		apiObject.MinimumHealthyPercentage = aws.Int32(int32(v.(int)))
	}
	if v, ok := m["protection_strategy"]; ok && v.(string) != "" {
		apiObject.ProtectionStrategy = awstypes.DeploymentProtectionStrategy(v.(string))
	}

	return apiObject
}

func expandGameSessionCreationLimitPolicy(tfList []any) *awstypes.GameSessionCreationLimitPolicy {
	if len(tfList) < 1 {
		return nil
	}

	m := tfList[0].(map[string]any)
	apiObject := &awstypes.GameSessionCreationLimitPolicy{}

	if v, ok := m["new_game_sessions_per_creator"]; ok && v.(int) > 0 {
		apiObject.NewGameSessionsPerCreator = aws.Int32(int32(v.(int)))
	}
	if v, ok := m["policy_period_in_minutes"]; ok && v.(int) > 0 {
		apiObject.PolicyPeriodInMinutes = aws.Int32(int32(v.(int)))
	}

	return apiObject
}

func flattenGameSessionCreationLimitPolicy(apiObject *awstypes.GameSessionCreationLimitPolicy) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{}
	if apiObject.NewGameSessionsPerCreator != nil {
		m["new_game_sessions_per_creator"] = aws.ToInt32(apiObject.NewGameSessionsPerCreator)
	}
	if apiObject.PolicyPeriodInMinutes != nil {
		m["policy_period_in_minutes"] = aws.ToInt32(apiObject.PolicyPeriodInMinutes)
	}

	if len(m) == 0 {
		return nil
	}

	return []any{m}
}

func flattenDeploymentDetails(apiObject *awstypes.DeploymentDetails) []any {
	if apiObject == nil || apiObject.LatestDeploymentId == nil {
		return nil
	}

	m := map[string]any{
		"latest_deployment_id": aws.ToString(apiObject.LatestDeploymentId),
	}

	return []any{m}
}

func flattenContainerFleetLocationAttributes(apiObjects []awstypes.ContainerFleetLocationAttributes) []any {
	if len(apiObjects) < 1 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		m := map[string]any{
			"location": aws.ToString(apiObject.Location),
			"status":   apiObject.Status,
		}
		tfList = append(tfList, m)
	}

	return tfList
}

func flattenLocationAttributesToLocations(apiObjects []awstypes.ContainerFleetLocationAttributes) []any {
	if len(apiObjects) < 1 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		if location := aws.ToString(apiObject.Location); location != "" {
			tfList = append(tfList, map[string]any{"location": location})
		}
	}

	return tfList
}
