// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appstream_fleet", name="Fleet")
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

		CustomizeDiff: customdiff.Sequence(
			resourceFleetCustDiff,
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_capacity": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"available": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"desired_instances": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"compute_capacity.0.desired_sessions",
							},
						},
						"desired_sessions": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"compute_capacity.0.desired_instances",
							},
						},
						"in_use": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"running": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"disconnect_timeout_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(60, 360000),
			},
			names.AttrDisplayName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"domain_join_info": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"organizational_unit_distinguished_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"enable_default_internet_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"fleet_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FleetType](),
			},
			names.AttrIAMRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"idle_disconnect_timeout_in_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ValidateFunc: validation.Any(
					validation.IntBetween(60, 360000),
					validation.IntInSlice([]int{0}),
				),
			},
			"image_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"image_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Required: true,
			},
			"max_sessions_per_instance": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"max_user_duration_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(600, 432000),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stream_view": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.StreamView](),
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceFleetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)
	input := &appstream.CreateFleetInput{
		Name:            aws.String(d.Get(names.AttrName).(string)),
		InstanceType:    aws.String(d.Get(names.AttrInstanceType).(string)),
		ComputeCapacity: expandComputeCapacity(d.Get("compute_capacity").([]interface{})),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disconnect_timeout_in_seconds"); ok {
		input.DisconnectTimeoutInSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("idle_disconnect_timeout_in_seconds"); ok {
		input.IdleDisconnectTimeoutInSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrDisplayName); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_join_info"); ok {
		input.DomainJoinInfo = expandDomainJoinInfo(v.([]interface{}))
	}

	if v, ok := d.GetOk("enable_default_internet_access"); ok {
		input.EnableDefaultInternetAccess = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("fleet_type"); ok {
		input.FleetType = awstypes.FleetType(v.(string))
	}

	if v, ok := d.GetOk("image_name"); ok {
		input.ImageName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_arn"); ok {
		input.ImageArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrIAMRoleARN); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_sessions_per_instance"); ok {
		input.MaxSessionsPerInstance = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("max_user_duration_in_seconds"); ok {
		input.MaxUserDurationInSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("stream_view"); ok {
		input.StreamView = awstypes.StreamView(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVPCConfig); ok {
		input.VpcConfig = expandVPCConfig(v.([]interface{}))
	}

	var err error
	var output *appstream.CreateFleetOutput
	err = retry.RetryContext(ctx, fleetOperationTimeout, func() *retry.RetryError {
		output, err = conn.CreateFleet(ctx, input)
		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.ConcurrentModificationException](err) {
				return retry.RetryableError(err)
			}

			// Retry for IAM eventual consistency on error:
			if errs.IsAErrorMessageContains[*awstypes.InvalidRoleException](err, "encountered an error because your IAM role") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateFleet(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appstream Fleet (%s): %s", d.Get(names.AttrName).(string), err)
	}

	d.SetId(aws.ToString(output.Fleet.Name))

	// Start fleet workflow
	_, err = conn.StartFleet(ctx, &appstream.StartFleetInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting Appstream Fleet (%s): %s", d.Id(), err)
	}

	if _, err = waitFleetStateRunning(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appstream Fleet (%s) to be running: %s", d.Id(), err)
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	resp, err := conn.DescribeFleets(ctx, &appstream.DescribeFleetsInput{Names: []string{d.Id()}})

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Appstream Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appstream Fleet (%s): %s", d.Id(), err)
	}

	if len(resp.Fleets) == 0 {
		return sdkdiag.AppendErrorf(diags, "reading Appstream Fleet (%s): %s", d.Id(), "empty response")
	}

	if len(resp.Fleets) > 1 {
		return sdkdiag.AppendErrorf(diags, "reading Appstream Fleet (%s): %s", d.Id(), "multiple fleets found")
	}

	fleet := resp.Fleets[0]

	d.Set(names.AttrARN, fleet.Arn)

	if fleet.ComputeCapacityStatus != nil {
		if err = d.Set("compute_capacity", []interface{}{flattenComputeCapacity(fleet.ComputeCapacityStatus)}); err != nil {
			return create.AppendDiagSettingError(diags, names.AppStream, "Fleet", d.Id(), "compute_capacity", err)
		}
	} else {
		d.Set("compute_capacity", nil)
	}

	d.Set(names.AttrCreatedTime, aws.ToTime(fleet.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, fleet.Description)
	d.Set(names.AttrDisplayName, fleet.DisplayName)
	d.Set("disconnect_timeout_in_seconds", fleet.DisconnectTimeoutInSeconds)

	if fleet.DomainJoinInfo != nil {
		if err = d.Set("domain_join_info", []interface{}{flattenDomainInfo(fleet.DomainJoinInfo)}); err != nil {
			return create.AppendDiagSettingError(diags, names.AppStream, "Fleet", d.Id(), "domain_join_info", err)
		}
	} else {
		d.Set("domain_join_info", nil)
	}

	d.Set("idle_disconnect_timeout_in_seconds", fleet.IdleDisconnectTimeoutInSeconds)
	d.Set("enable_default_internet_access", fleet.EnableDefaultInternetAccess)
	d.Set("fleet_type", fleet.FleetType)
	d.Set(names.AttrIAMRoleARN, fleet.IamRoleArn)
	d.Set("image_name", fleet.ImageName)
	d.Set("image_arn", fleet.ImageArn)
	d.Set(names.AttrInstanceType, fleet.InstanceType)
	d.Set("max_sessions_per_instance", fleet.MaxSessionsPerInstance)
	d.Set("max_user_duration_in_seconds", fleet.MaxUserDurationInSeconds)
	d.Set(names.AttrName, fleet.Name)
	d.Set(names.AttrState, fleet.State)
	d.Set("stream_view", fleet.StreamView)

	if fleet.VpcConfig != nil {
		if err = d.Set(names.AttrVPCConfig, []interface{}{flattenVPCConfig(fleet.VpcConfig)}); err != nil {
			return create.AppendDiagSettingError(diags, names.AppStream, "Fleet", d.Id(), names.AttrVPCConfig, err)
		}
	} else {
		d.Set(names.AttrVPCConfig, nil)
	}

	return diags
}

func resourceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)
	input := &appstream.UpdateFleetInput{
		Name: aws.String(d.Id()),
	}
	shouldStop := false

	if d.HasChanges(names.AttrDescription, "domain_join_info", "enable_default_internet_access", names.AttrIAMRoleARN, names.AttrInstanceType, "max_user_duration_in_seconds", "stream_view", names.AttrVPCConfig) {
		shouldStop = true
	}

	// Stop fleet workflow if needed
	if shouldStop {
		_, err := conn.StopFleet(ctx, &appstream.StopFleetInput{
			Name: aws.String(d.Id()),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "stopping Appstream Fleet (%s): %s", d.Id(), err)
		}
		if _, err = waitFleetStateStopped(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Appstream Fleet (%s) to be stopped: %s", d.Id(), err)
		}
	}

	if d.HasChange("compute_capacity") {
		input.ComputeCapacity = expandComputeCapacity(d.Get("compute_capacity").([]interface{}))
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange("domain_join_info") {
		input.DomainJoinInfo = expandDomainJoinInfo(d.Get("domain_join_info").([]interface{}))
	}

	if d.HasChange("disconnect_timeout_in_seconds") {
		input.DisconnectTimeoutInSeconds = aws.Int32(int32(d.Get("disconnect_timeout_in_seconds").(int)))
	}

	if d.HasChange("enable_default_internet_access") {
		input.EnableDefaultInternetAccess = aws.Bool(d.Get("enable_default_internet_access").(bool))
	}

	if d.HasChange("idle_disconnect_timeout_in_seconds") {
		input.IdleDisconnectTimeoutInSeconds = aws.Int32(int32(d.Get("idle_disconnect_timeout_in_seconds").(int)))
	}

	if d.HasChange(names.AttrDisplayName) {
		input.DisplayName = aws.String(d.Get(names.AttrDisplayName).(string))
	}

	if d.HasChange("image_name") {
		input.ImageName = aws.String(d.Get("image_name").(string))
	}

	if d.HasChange("image_arn") {
		input.ImageArn = aws.String(d.Get("image_arn").(string))
	}

	if d.HasChange(names.AttrIAMRoleARN) {
		input.IamRoleArn = aws.String(d.Get(names.AttrIAMRoleARN).(string))
	}

	if d.HasChange("stream_view") {
		input.StreamView = awstypes.StreamView(d.Get("stream_view").(string))
	}

	if d.HasChange(names.AttrInstanceType) {
		input.InstanceType = aws.String(d.Get(names.AttrInstanceType).(string))
	}

	if d.HasChange("max_sessions_per_instance") {
		input.MaxSessionsPerInstance = aws.Int32(int32(d.Get("max_sessions_per_instance").(int)))
	}

	if d.HasChange("max_user_duration_in_seconds") {
		input.MaxUserDurationInSeconds = aws.Int32(int32(d.Get("max_user_duration_in_seconds").(int)))
	}

	if d.HasChange(names.AttrVPCConfig) {
		input.VpcConfig = expandVPCConfig(d.Get(names.AttrVPCConfig).([]interface{}))
	}

	_, err := conn.UpdateFleet(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appstream Fleet (%s): %s", d.Id(), err)
	}

	// Start fleet workflow if stopped
	if shouldStop {
		_, err = conn.StartFleet(ctx, &appstream.StartFleetInput{
			Name: aws.String(d.Id()),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "starting Appstream Fleet (%s): %s", d.Id(), err)
		}

		if _, err = waitFleetStateRunning(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Appstream Fleet (%s) to be running: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	// Stop fleet workflow
	log.Printf("[DEBUG] Stopping AppStream Fleet: (%s)", d.Id())
	_, err := conn.StopFleet(ctx, &appstream.StopFleetInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "stopping Appstream Fleet (%s): %s", d.Id(), err)
	}

	if _, err = waitFleetStateStopped(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appstream Fleet (%s) to be stopped: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting AppStream Fleet: (%s)", d.Id())
	_, err = conn.DeleteFleet(ctx, &appstream.DeleteFleetInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appstream Fleet (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceFleetCustDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if diff.HasChange("domain_join_info") {
		o, n := diff.GetChange("domain_join_info")

		if reflect.DeepEqual(expandDomainJoinInfo(o.([]interface{})), expandDomainJoinInfo(n.([]interface{}))) {
			return diff.Clear("domain_join_info")
		}
	}
	return nil
}

func expandComputeCapacity(tfList []interface{}) *awstypes.ComputeCapacity {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.ComputeCapacity{}

	attr := tfList[0].(map[string]interface{})
	if v, ok := attr["desired_instances"]; ok && v != 0 {
		apiObject.DesiredInstances = aws.Int32(int32(v.(int)))
	}

	if v, ok := attr["desired_sessions"]; ok && v != 0 {
		apiObject.DesiredSessions = aws.Int32(int32(v.(int)))
	}

	if reflect.DeepEqual(&awstypes.ComputeCapacity{}, apiObject) {
		return nil
	}

	return apiObject
}

func flattenComputeCapacity(apiObject *awstypes.ComputeCapacityStatus) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DesiredUserSessions; v != nil {
		tfMap["desired_sessions"] = aws.ToInt32(v)
	}

	// desiredInstances is always returned by the API but cannot be used in conjunction with desiredSessions
	if v := apiObject.Desired; v != nil && tfMap["desired_sessions"] == nil {
		tfMap["desired_instances"] = aws.ToInt32(v)
	}

	if v := apiObject.Available; v != nil {
		tfMap["available"] = aws.ToInt32(v)
	}

	if v := apiObject.InUse; v != nil {
		tfMap["in_use"] = aws.ToInt32(v)
	}

	if v := apiObject.Running; v != nil {
		tfMap["running"] = aws.ToInt32(v)
	}

	if reflect.DeepEqual(map[string]interface{}{}, tfMap) {
		return nil
	}

	return tfMap
}

func expandDomainJoinInfo(tfList []interface{}) *awstypes.DomainJoinInfo {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.DomainJoinInfo{}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	if v, ok := tfMap["directory_name"]; ok && v != "" {
		apiObject.DirectoryName = aws.String(v.(string))
	}

	if v, ok := tfMap["organizational_unit_distinguished_name"]; ok && v != "" {
		apiObject.OrganizationalUnitDistinguishedName = aws.String(v.(string))
	}

	if reflect.DeepEqual(&awstypes.DomainJoinInfo{}, apiObject) {
		return nil
	}

	return apiObject
}

func flattenDomainInfo(apiObject *awstypes.DomainJoinInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DirectoryName; v != nil && aws.ToString(v) != "" {
		tfMap["directory_name"] = aws.ToString(v)
	}

	if v := apiObject.OrganizationalUnitDistinguishedName; v != nil && aws.ToString(v) != "" {
		tfMap["organizational_unit_distinguished_name"] = aws.ToString(v)
	}

	if reflect.DeepEqual(map[string]interface{}{}, tfMap) {
		return nil
	}

	return tfMap
}

func expandVPCConfig(tfList []interface{}) *awstypes.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.VpcConfig{}

	tfMap := tfList[0].(map[string]interface{})
	if v, ok := tfMap[names.AttrSecurityGroupIDs]; ok {
		apiObject.SecurityGroupIds = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := tfMap[names.AttrSubnetIDs]; ok {
		apiObject.SubnetIds = flex.ExpandStringValueList(v.([]interface{}))
	}

	if reflect.DeepEqual(&awstypes.VpcConfig{}, apiObject) {
		return nil
	}

	return apiObject
}

func flattenVPCConfig(apiObject *awstypes.VpcConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap[names.AttrSecurityGroupIDs] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = v
	}

	if reflect.DeepEqual(map[string]interface{}{}, tfMap) {
		return nil
	}

	return tfMap
}
