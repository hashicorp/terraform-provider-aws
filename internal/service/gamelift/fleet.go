// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	fleetCreatedDefaultTimeout = 70 * time.Minute
	FleetDeletedDefaultTimeout = 20 * time.Minute
)

// @SDKResource("aws_gamelift_fleet", name="Fleet")
// @Tags(identifierAttribute="arn")
func ResourceFleet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFleetCreate,
		ReadWithoutTimeout:   resourceFleetRead,
		UpdateWithoutTimeout: resourceFleetUpdate,
		DeleteWithoutTimeout: resourceFleetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(fleetCreatedDefaultTimeout),
			Delete: schema.DefaultTimeout(FleetDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"build_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"build_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"build_id", "script_id"},
			},
			"certificate_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.CertificateTypeDisabled,
							ValidateDiagFunc: enum.Validate[awstypes.CertificateType](),
						},
					},
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"ec2_inbound_permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MaxItems: 50,
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
						names.AttrProtocol: {
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
			"ec2_instance_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.EC2InstanceType](),
			},
			"fleet_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.FleetTypeOnDemand,
				ValidateDiagFunc: enum.Validate[awstypes.FleetType](),
			},
			"instance_role_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				Optional:     true,
			},
			"log_paths": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"metric_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"new_game_session_protection_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ProtectionPolicyNoProtection,
				ValidateDiagFunc: enum.Validate[awstypes.ProtectionPolicy](),
			},
			"operating_system": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_creation_limit_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
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
			"runtime_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"game_session_activation_timeout_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 600),
						},
						"max_concurrent_game_session_activations": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 2147483647),
						},
						"server_process": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 50,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"concurrent_executions": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"launch_path": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									names.AttrParameters: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
								},
							},
						},
					},
				},
			},
			"script_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"script_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"build_id", "script_id"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFleetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	input := &gamelift.CreateFleetInput{
		EC2InstanceType: awstypes.EC2InstanceType(d.Get("ec2_instance_type").(string)),
		Name:            aws.String(d.Get(names.AttrName).(string)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("build_id"); ok {
		input.BuildId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("script_id"); ok {
		input.ScriptId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("fleet_type"); ok {
		input.FleetType = awstypes.FleetType(v.(string))
	}
	if v, ok := d.GetOk("ec2_inbound_permission"); ok {
		input.EC2InboundPermissions = slices.Values(expandIPPermissions(v.(*schema.Set)))
	}

	if v, ok := d.GetOk("instance_role_arn"); ok {
		input.InstanceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metric_groups"); ok {
		input.MetricGroups = flex.ExpandStringValueList(v.([]interface{}))
	}
	if v, ok := d.GetOk("new_game_session_protection_policy"); ok {
		input.NewGameSessionProtectionPolicy = awstypes.ProtectionPolicy(v.(string))
	}
	if v, ok := d.GetOk("resource_creation_limit_policy"); ok {
		input.ResourceCreationLimitPolicy = expandResourceCreationLimitPolicy(v.([]interface{}))
	}
	if v, ok := d.GetOk("runtime_configuration"); ok {
		input.RuntimeConfiguration = expandRuntimeConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("certificate_configuration"); ok {
		input.CertificateConfiguration = expandCertificateConfiguration(v.([]interface{}))
	}

	log.Printf("[INFO] Creating GameLift Fleet: %+v", input)
	var out *gamelift.CreateFleetOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		out, err = conn.CreateFleet(ctx, input)

		if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "GameLift is not authorized to perform") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		out, err = conn.CreateFleet(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Fleet (%s): %s", d.Get(names.AttrName).(string), err)
	}

	d.SetId(aws.ToString(out.FleetAttributes.FleetId))

	if _, err := waitFleetActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Fleet (%s) to active: %s", d.Id(), err)
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Describing GameLift Fleet: %s", d.Id())
	fleet, err := FindFleetByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Fleet (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(fleet.FleetArn)
	d.Set("build_arn", fleet.BuildArn)
	d.Set("build_id", fleet.BuildId)
	d.Set(names.AttrDescription, fleet.Description)
	d.Set(names.AttrARN, arn)
	d.Set("log_paths", fleet.LogPaths)
	d.Set("metric_groups", flex.FlattenStringValueList(fleet.MetricGroups))
	d.Set(names.AttrName, fleet.Name)
	d.Set("fleet_type", fleet.FleetType)
	d.Set("instance_role_arn", fleet.InstanceRoleArn)
	d.Set("ec2_instance_type", fleet.InstanceType)
	d.Set("new_game_session_protection_policy", fleet.NewGameSessionProtectionPolicy)
	d.Set("operating_system", fleet.OperatingSystem)
	d.Set("script_arn", fleet.ScriptArn)
	d.Set("script_id", fleet.ScriptId)

	if err := d.Set("certificate_configuration", flattenCertificateConfiguration(fleet.CertificateConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting certificate_configuration: %s", err)
	}

	if err := d.Set("resource_creation_limit_policy", flattenResourceCreationLimitPolicy(fleet.ResourceCreationLimitPolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resource_creation_limit_policy: %s", err)
	}

	portInput := &gamelift.DescribeFleetPortSettingsInput{
		FleetId: aws.String(d.Id()),
	}

	portConfig, err := conn.DescribeFleetPortSettings(ctx, portInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading for GameLift Fleet ec2 inbound permission (%s): %s", d.Id(), err)
	}

	if err := d.Set("ec2_inbound_permission", flattenIPPermissions(slices.ToPointers(portConfig.InboundPermissions))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ec2_inbound_permission: %s", err)
	}

	return diags
}

func resourceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Updating GameLift Fleet: %s", d.Id())

	if d.HasChanges(names.AttrDescription, "metric_groups", names.AttrName, "new_game_session_protection_policy", "resource_creation_limit_policy") {
		_, err := conn.UpdateFleetAttributes(ctx, &gamelift.UpdateFleetAttributesInput{
			Description:                    aws.String(d.Get(names.AttrDescription).(string)),
			FleetId:                        aws.String(d.Id()),
			MetricGroups:                   flex.ExpandStringValueList(d.Get("metric_groups").([]interface{})),
			Name:                           aws.String(d.Get(names.AttrName).(string)),
			NewGameSessionProtectionPolicy: awstypes.ProtectionPolicy(d.Get("new_game_session_protection_policy").(string)),
			ResourceCreationLimitPolicy:    expandResourceCreationLimitPolicy(d.Get("resource_creation_limit_policy").([]interface{})),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating for GameLift Fleet attributes (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("ec2_inbound_permission") {
		oldPerms, newPerms := d.GetChange("ec2_inbound_permission")
		authorizations, revocations := DiffPortSettings(oldPerms.(*schema.Set).List(), newPerms.(*schema.Set).List())

		_, err := conn.UpdateFleetPortSettings(ctx, &gamelift.UpdateFleetPortSettingsInput{
			FleetId:                         aws.String(d.Id()),
			InboundPermissionAuthorizations: slices.Values(authorizations),
			InboundPermissionRevocations:    slices.Values(revocations),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating for GameLift Fleet port settings (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("runtime_configuration") {
		_, err := conn.UpdateRuntimeConfiguration(ctx, &gamelift.UpdateRuntimeConfigurationInput{
			FleetId:              aws.String(d.Id()),
			RuntimeConfiguration: expandRuntimeConfiguration(d.Get("runtime_configuration").([]interface{})),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating for GameLift Fleet runtime configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Deleting GameLift Fleet: %s", d.Id())
	// It can take ~ 1 hr as GameLift will keep retrying on errors like
	// invalid launch path and remain in state when it can't be deleted :/
	input := &gamelift.DeleteFleetInput{
		FleetId: aws.String(d.Id()),
	}
	err := retry.RetryContext(ctx, 60*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteFleet(ctx, input)
		if err != nil {
			msg := fmt.Sprintf("Cannot delete fleet %s that is in status of ", d.Id())
			if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, msg) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteFleet(ctx, input)
	}
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting GameLift fleet: %s", err)
	}

	if _, err := waitFleetTerminated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Fleet (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func expandIPPermissions(cfgs *schema.Set) []*awstypes.IpPermission {
	if cfgs.Len() < 1 {
		return []*awstypes.IpPermission{}
	}

	perms := make([]*awstypes.IpPermission, cfgs.Len())
	for i, rawCfg := range cfgs.List() {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandIPPermission(cfg)
	}
	return perms
}

func expandIPPermission(cfg map[string]interface{}) *awstypes.IpPermission {
	return &awstypes.IpPermission{
		FromPort: aws.Int32(int32(cfg["from_port"].(int))),
		IpRange:  aws.String(cfg["ip_range"].(string)),
		Protocol: awstypes.IpProtocol(cfg[names.AttrProtocol].(string)),
		ToPort:   aws.Int32(int32(cfg["to_port"].(int))),
	}
}

func flattenIPPermissions(apiObjects []*awstypes.IpPermission) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if v := flattenIPPermission(apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

func flattenIPPermission(apiObject *awstypes.IpPermission) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["from_port"] = aws.ToInt32(apiObject.FromPort)
	tfMap["to_port"] = aws.ToInt32(apiObject.ToPort)
	tfMap[names.AttrProtocol] = string(apiObject.Protocol)
	tfMap["ip_range"] = aws.ToString(apiObject.IpRange)

	return tfMap
}

func expandResourceCreationLimitPolicy(cfg []interface{}) *awstypes.ResourceCreationLimitPolicy {
	if len(cfg) < 1 {
		return nil
	}
	out := awstypes.ResourceCreationLimitPolicy{}
	m := cfg[0].(map[string]interface{})

	if v, ok := m["new_game_sessions_per_creator"]; ok {
		out.NewGameSessionsPerCreator = aws.Int32(int32(v.(int)))
	}
	if v, ok := m["policy_period_in_minutes"]; ok {
		out.PolicyPeriodInMinutes = aws.Int32(int32(v.(int)))
	}

	return &out
}

func flattenResourceCreationLimitPolicy(policy *awstypes.ResourceCreationLimitPolicy) []interface{} {
	if policy == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	m["new_game_sessions_per_creator"] = aws.ToInt32(policy.NewGameSessionsPerCreator)
	m["policy_period_in_minutes"] = aws.ToInt32(policy.PolicyPeriodInMinutes)

	return []interface{}{m}
}

func expandRuntimeConfiguration(cfg []interface{}) *awstypes.RuntimeConfiguration {
	if len(cfg) < 1 {
		return nil
	}
	out := awstypes.RuntimeConfiguration{}
	m := cfg[0].(map[string]interface{})

	if v, ok := m["game_session_activation_timeout_seconds"].(int); ok && v > 0 {
		out.GameSessionActivationTimeoutSeconds = aws.Int32(int32(v))
	}
	if v, ok := m["max_concurrent_game_session_activations"].(int); ok && v > 0 {
		out.MaxConcurrentGameSessionActivations = aws.Int32(int32(v))
	}
	if v, ok := m["server_process"]; ok {
		out.ServerProcesses = expandServerProcesses(v.([]interface{}))
	}

	return &out
}

func expandServerProcesses(cfgs []interface{}) []awstypes.ServerProcess {
	if len(cfgs) < 1 {
		return []awstypes.ServerProcess{}
	}

	processes := make([]awstypes.ServerProcess, len(cfgs))
	for i, rawCfg := range cfgs {
		cfg := rawCfg.(map[string]interface{})
		process := awstypes.ServerProcess{
			ConcurrentExecutions: aws.Int32(int32(cfg["concurrent_executions"].(int))),
			LaunchPath:           aws.String(cfg["launch_path"].(string)),
		}
		if v, ok := cfg[names.AttrParameters].(string); ok && len(v) > 0 {
			process.Parameters = aws.String(v)
		}
		processes[i] = process
	}
	return processes
}

func expandCertificateConfiguration(cfg []interface{}) *awstypes.CertificateConfiguration {
	if len(cfg) < 1 {
		return nil
	}
	out := awstypes.CertificateConfiguration{}
	m := cfg[0].(map[string]interface{})

	if v, ok := m["certificate_type"].(string); ok {
		out.CertificateType = awstypes.CertificateType(v)
	}

	return &out
}

func flattenCertificateConfiguration(config *awstypes.CertificateConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	m["certificate_type"] = string(config.CertificateType)

	return []interface{}{m}
}

func DiffPortSettings(oldPerms, newPerms []interface{}) (a []*awstypes.IpPermission, r []*awstypes.IpPermission) {
OUTER:
	for i, op := range oldPerms {
		oldPerm := op.(map[string]interface{})
		for j, np := range newPerms {
			newPerm := np.(map[string]interface{})

			// Remove permissions which have not changed so we're not wasting
			// API calls for removal & subseq. addition of the same ones
			if reflect.DeepEqual(oldPerm, newPerm) {
				oldPerms = append(oldPerms[:i], oldPerms[i+1:]...)
				newPerms = append(newPerms[:j], newPerms[j+1:]...)
				continue OUTER
			}
		}

		// Add what's left for revocation
		r = append(r, expandIPPermission(oldPerm))
	}

	for _, np := range newPerms {
		newPerm := np.(map[string]interface{})
		// Add what's left for authorization
		a = append(a, expandIPPermission(newPerm))
	}
	return
}
