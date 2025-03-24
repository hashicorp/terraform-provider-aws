// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"slices"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_gamelift_fleet", name="Fleet")
// @Tags(identifierAttribute="arn")
func resourceFleet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFleetCreate,
		ReadWithoutTimeout:   resourceFleetRead,
		UpdateWithoutTimeout: resourceFleetUpdate,
		DeleteWithoutTimeout: resourceFleetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(70 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
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
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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
	}
}

func resourceFleetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	startTime := time.Now()
	name := d.Get(names.AttrName).(string)
	input := &gamelift.CreateFleetInput{
		EC2InstanceType: awstypes.EC2InstanceType(d.Get("ec2_instance_type").(string)),
		Name:            aws.String(name),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("build_id"); ok {
		input.BuildId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_configuration"); ok {
		input.CertificateConfiguration = expandCertificateConfiguration(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("fleet_type"); ok {
		input.FleetType = awstypes.FleetType(v.(string))
	}

	if v, ok := d.GetOk("ec2_inbound_permission"); ok {
		input.EC2InboundPermissions = expandIPPermissions(v.(*schema.Set))
	}

	if v, ok := d.GetOk("instance_role_arn"); ok {
		input.InstanceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metric_groups"); ok {
		input.MetricGroups = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk("new_game_session_protection_policy"); ok {
		input.NewGameSessionProtectionPolicy = awstypes.ProtectionPolicy(v.(string))
	}

	if v, ok := d.GetOk("resource_creation_limit_policy"); ok {
		input.ResourceCreationLimitPolicy = expandResourceCreationLimitPolicy(v.([]any))
	}

	if v, ok := d.GetOk("runtime_configuration"); ok {
		input.RuntimeConfiguration = expandRuntimeConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("script_id"); ok {
		input.ScriptId = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidRequestException](ctx, propagationTimeout, func() (any, error) {
		return conn.CreateFleet(ctx, input)
	}, "GameLift is not authorized to perform")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Fleet (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*gamelift.CreateFleetOutput).FleetAttributes.FleetId))

	if _, err := waitFleetActive(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Fleet (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	fleet, err := findFleetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Fleet (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, fleet.FleetArn)
	d.Set("build_arn", fleet.BuildArn)
	d.Set("build_id", fleet.BuildId)
	if err := d.Set("certificate_configuration", flattenCertificateConfiguration(fleet.CertificateConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting certificate_configuration: %s", err)
	}
	d.Set(names.AttrDescription, fleet.Description)
	d.Set("log_paths", fleet.LogPaths)
	d.Set("metric_groups", fleet.MetricGroups)
	d.Set(names.AttrName, fleet.Name)
	d.Set("fleet_type", fleet.FleetType)
	d.Set("instance_role_arn", fleet.InstanceRoleArn)
	d.Set("ec2_instance_type", fleet.InstanceType)
	d.Set("new_game_session_protection_policy", fleet.NewGameSessionProtectionPolicy)
	d.Set("operating_system", fleet.OperatingSystem)
	if err := d.Set("resource_creation_limit_policy", flattenResourceCreationLimitPolicy(fleet.ResourceCreationLimitPolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resource_creation_limit_policy: %s", err)
	}
	d.Set("script_arn", fleet.ScriptArn)
	d.Set("script_id", fleet.ScriptId)

	input := &gamelift.DescribeFleetPortSettingsInput{
		FleetId: aws.String(d.Id()),
	}

	output, err := conn.DescribeFleetPortSettings(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Fleet (%s) inbound connection permissions: %s", d.Id(), err)
	}

	if err := d.Set("ec2_inbound_permission", flattenIPPermissions(output.InboundPermissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ec2_inbound_permission: %s", err)
	}

	return diags
}

func resourceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChanges(names.AttrDescription, "metric_groups", names.AttrName, "new_game_session_protection_policy", "resource_creation_limit_policy") {
		input := &gamelift.UpdateFleetAttributesInput{
			Description:                    aws.String(d.Get(names.AttrDescription).(string)),
			FleetId:                        aws.String(d.Id()),
			MetricGroups:                   flex.ExpandStringValueList(d.Get("metric_groups").([]any)),
			Name:                           aws.String(d.Get(names.AttrName).(string)),
			NewGameSessionProtectionPolicy: awstypes.ProtectionPolicy(d.Get("new_game_session_protection_policy").(string)),
			ResourceCreationLimitPolicy:    expandResourceCreationLimitPolicy(d.Get("resource_creation_limit_policy").([]any)),
		}

		_, err := conn.UpdateFleetAttributes(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Fleet (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("ec2_inbound_permission") {
		o, n := d.GetChange("ec2_inbound_permission")
		authorizations, revocations := diffPortSettings(o.(*schema.Set).List(), n.(*schema.Set).List())
		input := &gamelift.UpdateFleetPortSettingsInput{
			FleetId:                         aws.String(d.Id()),
			InboundPermissionAuthorizations: authorizations,
			InboundPermissionRevocations:    revocations,
		}

		_, err := conn.UpdateFleetPortSettings(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Fleet (%s) port settings: %s", d.Id(), err)
		}
	}

	if d.HasChange("runtime_configuration") {
		input := &gamelift.UpdateRuntimeConfigurationInput{
			FleetId:              aws.String(d.Id()),
			RuntimeConfiguration: expandRuntimeConfiguration(d.Get("runtime_configuration").([]any)),
		}

		_, err := conn.UpdateRuntimeConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Fleet (%s) runtime configuration: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	startTime := time.Now()

	log.Printf("[INFO] Deleting GameLift Fleet: %s", d.Id())
	// It can take ~ 1 hr as GameLift will keep retrying on errors like
	// invalid launch path and remain in state when it can't be deleted :/
	msg := fmt.Sprintf("Cannot delete fleet %s that is in status of ", d.Id())
	const (
		timeout = 60 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidRequestException](ctx, timeout, func() (any, error) {
		return conn.DeleteFleet(ctx, &gamelift.DeleteFleetInput{
			FleetId: aws.String(d.Id()),
		})
	}, msg)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Fleet (%s): %s", d.Id(), err)
	}

	if _, err := waitFleetTerminated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Fleet (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFleetByID(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.FleetAttributes, error) {
	input := &gamelift.DescribeFleetAttributesInput{
		FleetIds: []string{id},
	}

	return findFleet(ctx, conn, input)
}

func findFleet(ctx context.Context, conn *gamelift.Client, input *gamelift.DescribeFleetAttributesInput) (*awstypes.FleetAttributes, error) {
	output, err := findFleets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFleets(ctx context.Context, conn *gamelift.Client, input *gamelift.DescribeFleetAttributesInput) ([]awstypes.FleetAttributes, error) {
	var output []awstypes.FleetAttributes

	pages := gamelift.NewDescribeFleetAttributesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.FleetAttributes...)
	}

	return output, nil
}

func findFleetFailuresByID(ctx context.Context, conn *gamelift.Client, id string) ([]awstypes.Event, error) {
	input := &gamelift.DescribeFleetEventsInput{
		FleetId: aws.String(id),
	}

	return findFleetEvents(ctx, conn, input, isFailureEvent)
}

func findFleetEvents(ctx context.Context, conn *gamelift.Client, input *gamelift.DescribeFleetEventsInput, filter tfslices.Predicate[*awstypes.Event]) ([]awstypes.Event, error) {
	var output []awstypes.Event

	pages := gamelift.NewDescribeFleetEventsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Events {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusFleet(ctx context.Context, conn *gamelift.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFleetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitFleetActive(ctx context.Context, conn *gamelift.Client, id string, startTime time.Time, timeout time.Duration) (*awstypes.FleetAttributes, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.FleetStatusActivating,
			awstypes.FleetStatusBuilding,
			awstypes.FleetStatusDownloading,
			awstypes.FleetStatusNew,
			awstypes.FleetStatusValidating,
		),
		Target:  enum.Slice(awstypes.FleetStatusActive),
		Refresh: statusFleet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FleetAttributes); ok {
		if events, errFFF := findFleetFailuresByID(ctx, conn, id); errFFF == nil {
			tfresource.SetLastError(err, fleetFailuresError(events, startTime))
		}

		return output, err
	}

	return nil, err
}

func waitFleetTerminated(ctx context.Context, conn *gamelift.Client, id string, startTime time.Time, timeout time.Duration) (*awstypes.FleetAttributes, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.FleetStatusActive,
			awstypes.FleetStatusDeleting,
			awstypes.FleetStatusError,
			awstypes.FleetStatusTerminated,
		),
		Target:  []string{},
		Refresh: statusFleet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FleetAttributes); ok {
		if events, errFFF := findFleetFailuresByID(ctx, conn, id); errFFF == nil {
			tfresource.SetLastError(err, fleetFailuresError(events, startTime))
		}

		return output, err
	}

	return nil, err
}

func isFailureEvent(event *awstypes.Event) bool {
	failureCodes := []awstypes.EventCode{
		awstypes.EventCodeFleetActivationFailed,
		awstypes.EventCodeFleetActivationFailedNoInstances,
		awstypes.EventCodeFleetBinaryDownloadFailed,
		awstypes.EventCodeFleetInitializationFailed,
		awstypes.EventCodeFleetStateError,
		awstypes.EventCodeFleetValidationExecutableRuntimeFailure,
		awstypes.EventCodeFleetValidationLaunchPathNotFound,
		awstypes.EventCodeFleetValidationTimedOut,
		awstypes.EventCodeFleetVpcPeeringFailed,
		awstypes.EventCodeGameSessionActivationTimeout,
		awstypes.EventCodeServerProcessCrashed,
		awstypes.EventCodeServerProcessForceTerminated,
		awstypes.EventCodeServerProcessInvalidPath,
		awstypes.EventCodeServerProcessProcessExitTimeout,
		awstypes.EventCodeServerProcessProcessReadyTimeout,
		awstypes.EventCodeServerProcessSdkInitializationTimeout,
		awstypes.EventCodeServerProcessTerminatedUnhealthy,
	}

	return slices.Contains(failureCodes, event.EventCode)
}

func fleetFailuresError(events []awstypes.Event, startTime time.Time) error {
	events = tfslices.Filter(events, func(v awstypes.Event) bool {
		return startTime.Before(aws.ToTime(v.EventTime))
	})
	errs := tfslices.ApplyToAll(events, func(v awstypes.Event) error {
		return fmt.Errorf("(%s) %s: %s", aws.ToString(v.EventId), v.EventCode, aws.ToString(v.Message))
	})
	return errors.Join(errs...)
}

func expandIPPermissions(tfSet *schema.Set) []awstypes.IpPermission {
	if tfSet.Len() < 1 {
		return nil
	}

	apiObjects := make([]awstypes.IpPermission, 0)

	for _, tfMapRaw := range tfSet.List() {
		tfMap := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, *expandIPPermission(tfMap))
	}

	return apiObjects
}

func expandIPPermission(tfMap map[string]any) *awstypes.IpPermission {
	return &awstypes.IpPermission{
		FromPort: aws.Int32(int32(tfMap["from_port"].(int))),
		IpRange:  aws.String(tfMap["ip_range"].(string)),
		Protocol: awstypes.IpProtocol(tfMap[names.AttrProtocol].(string)),
		ToPort:   aws.Int32(int32(tfMap["to_port"].(int))),
	}
}

func flattenIPPermissions(apiObjects []awstypes.IpPermission) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any
	for _, apiObject := range apiObjects {
		if v := flattenIPPermission(&apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

func flattenIPPermission(apiObject *awstypes.IpPermission) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)
	tfMap["from_port"] = aws.ToInt32(apiObject.FromPort)
	tfMap["to_port"] = aws.ToInt32(apiObject.ToPort)
	tfMap[names.AttrProtocol] = apiObject.Protocol
	tfMap["ip_range"] = aws.ToString(apiObject.IpRange)

	return tfMap
}

func expandResourceCreationLimitPolicy(tfList []any) *awstypes.ResourceCreationLimitPolicy {
	if len(tfList) < 1 {
		return nil
	}

	apiObject := &awstypes.ResourceCreationLimitPolicy{}
	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["new_game_sessions_per_creator"]; ok {
		apiObject.NewGameSessionsPerCreator = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap["policy_period_in_minutes"]; ok {
		apiObject.PolicyPeriodInMinutes = aws.Int32(int32(v.(int)))
	}

	return apiObject
}

func flattenResourceCreationLimitPolicy(apiObject *awstypes.ResourceCreationLimitPolicy) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)
	tfMap["new_game_sessions_per_creator"] = aws.ToInt32(apiObject.NewGameSessionsPerCreator)
	tfMap["policy_period_in_minutes"] = aws.ToInt32(apiObject.PolicyPeriodInMinutes)

	return []any{tfMap}
}

func expandRuntimeConfiguration(tfList []any) *awstypes.RuntimeConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	apiObject := &awstypes.RuntimeConfiguration{}
	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["game_session_activation_timeout_seconds"].(int); ok && v > 0 {
		apiObject.GameSessionActivationTimeoutSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_concurrent_game_session_activations"].(int); ok && v > 0 {
		apiObject.MaxConcurrentGameSessionActivations = aws.Int32(int32(v))
	}

	if v, ok := tfMap["server_process"]; ok {
		apiObject.ServerProcesses = expandServerProcesses(v.([]any))
	}

	return apiObject
}

func expandServerProcesses(tfList []any) []awstypes.ServerProcess {
	if len(tfList) < 1 {
		return []awstypes.ServerProcess{}
	}

	apiObjects := make([]awstypes.ServerProcess, len(tfList))

	for i, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.ServerProcess{
			ConcurrentExecutions: aws.Int32(int32(tfMap["concurrent_executions"].(int))),
			LaunchPath:           aws.String(tfMap["launch_path"].(string)),
		}

		if v, ok := tfMap[names.AttrParameters].(string); ok && len(v) > 0 {
			apiObject.Parameters = aws.String(v)
		}

		apiObjects[i] = apiObject
	}

	return apiObjects
}

func expandCertificateConfiguration(tfList []any) *awstypes.CertificateConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	apiObject := &awstypes.CertificateConfiguration{}
	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["certificate_type"].(string); ok {
		apiObject.CertificateType = awstypes.CertificateType(v)
	}

	return apiObject
}

func flattenCertificateConfiguration(apiObject *awstypes.CertificateConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)
	tfMap["certificate_type"] = string(apiObject.CertificateType)

	return []any{tfMap}
}

func diffPortSettings(oldPerms, newPerms []any) (a []awstypes.IpPermission, r []awstypes.IpPermission) {
OUTER:
	for i, op := range oldPerms {
		oldPerm := op.(map[string]any)
		for j, np := range newPerms {
			newPerm := np.(map[string]any)

			// Remove permissions which have not changed so we're not wasting
			// API calls for removal & subseq. addition of the same ones
			if reflect.DeepEqual(oldPerm, newPerm) {
				oldPerms = slices.Delete(oldPerms, i, i+1)
				newPerms = slices.Delete(newPerms, j, j+1)
				continue OUTER
			}
		}

		// Add what's left for revocation
		r = append(r, *expandIPPermission(oldPerm))
	}

	for _, np := range newPerms {
		newPerm := np.(map[string]any)
		// Add what's left for authorization
		a = append(a, *expandIPPermission(newPerm))
	}
	return
}
