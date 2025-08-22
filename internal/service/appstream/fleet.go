// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appstream_fleet", name="Fleet")
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

		CustomizeDiff: resourceFleetCustDiff,

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

func resourceFleetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := appstream.CreateFleetInput{
		ComputeCapacity: expandComputeCapacity(d.Get("compute_capacity").([]any)),
		InstanceType:    aws.String(d.Get(names.AttrInstanceType).(string)),
		Name:            aws.String(name),
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
		input.DomainJoinInfo = expandDomainJoinInfo(v.([]any))
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
		input.VpcConfig = expandVPCConfig(v.([]any))
	}

	const (
		timeout = 15 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhen(ctx, timeout,
		func() (any, error) {
			return conn.CreateFleet(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.ConcurrentModificationException](err) {
				return true, err
			}

			// Retry for IAM eventual consistency.
			if errs.IsAErrorMessageContains[*awstypes.InvalidRoleException](err, "encountered an error because your IAM role") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream Fleet (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*appstream.CreateFleetOutput).Fleet.Name))

	if err := startFleet(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	fleet, err := findFleetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppStream Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream Fleet (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, fleet.Arn)
	if fleet.ComputeCapacityStatus != nil {
		if err = d.Set("compute_capacity", []any{flattenComputeCapacity(fleet.ComputeCapacityStatus)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting compute_capacity: %s", err)
		}
	} else {
		d.Set("compute_capacity", nil)
	}
	d.Set(names.AttrCreatedTime, aws.ToTime(fleet.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, fleet.Description)
	d.Set(names.AttrDisplayName, fleet.DisplayName)
	d.Set("disconnect_timeout_in_seconds", fleet.DisconnectTimeoutInSeconds)
	if fleet.DomainJoinInfo != nil {
		if err = d.Set("domain_join_info", []any{flattenDomainInfo(fleet.DomainJoinInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting domain_join_info: %s", err)
		}
	} else {
		d.Set("domain_join_info", nil)
	}
	d.Set("enable_default_internet_access", fleet.EnableDefaultInternetAccess)
	d.Set("fleet_type", fleet.FleetType)
	d.Set(names.AttrIAMRoleARN, fleet.IamRoleArn)
	d.Set("idle_disconnect_timeout_in_seconds", fleet.IdleDisconnectTimeoutInSeconds)
	d.Set("image_arn", fleet.ImageArn)
	d.Set("image_name", fleet.ImageName)
	d.Set(names.AttrInstanceType, fleet.InstanceType)
	d.Set("max_sessions_per_instance", fleet.MaxSessionsPerInstance)
	d.Set("max_user_duration_in_seconds", fleet.MaxUserDurationInSeconds)
	d.Set(names.AttrName, fleet.Name)
	d.Set(names.AttrState, fleet.State)
	d.Set("stream_view", fleet.StreamView)
	if fleet.VpcConfig != nil {
		if err = d.Set(names.AttrVPCConfig, []any{flattenVPCConfig(fleet.VpcConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
		}
	} else {
		d.Set(names.AttrVPCConfig, nil)
	}

	return diags
}

func resourceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		shouldStop := false
		if d.HasChanges(names.AttrDescription, "domain_join_info", "enable_default_internet_access", names.AttrIAMRoleARN, names.AttrInstanceType, "max_user_duration_in_seconds", "stream_view", names.AttrVPCConfig) {
			shouldStop = true
		}

		if shouldStop {
			if err := stopFleet(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		input := appstream.UpdateFleetInput{
			Name: aws.String(d.Id()),
		}

		if d.HasChange("compute_capacity") {
			input.ComputeCapacity = expandComputeCapacity(d.Get("compute_capacity").([]any))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("disconnect_timeout_in_seconds") {
			input.DisconnectTimeoutInSeconds = aws.Int32(int32(d.Get("disconnect_timeout_in_seconds").(int)))
		}

		if d.HasChange(names.AttrDisplayName) {
			input.DisplayName = aws.String(d.Get(names.AttrDisplayName).(string))
		}

		if d.HasChange("domain_join_info") {
			input.DomainJoinInfo = expandDomainJoinInfo(d.Get("domain_join_info").([]any))
		}

		if d.HasChange("enable_default_internet_access") {
			input.EnableDefaultInternetAccess = aws.Bool(d.Get("enable_default_internet_access").(bool))
		}

		if d.HasChange(names.AttrIAMRoleARN) {
			input.IamRoleArn = aws.String(d.Get(names.AttrIAMRoleARN).(string))
		}

		if d.HasChange("idle_disconnect_timeout_in_seconds") {
			input.IdleDisconnectTimeoutInSeconds = aws.Int32(int32(d.Get("idle_disconnect_timeout_in_seconds").(int)))
		}

		if d.HasChange("image_name") {
			input.ImageName = aws.String(d.Get("image_name").(string))
		}

		if d.HasChange("image_arn") {
			input.ImageArn = aws.String(d.Get("image_arn").(string))
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

		if d.HasChange("stream_view") {
			input.StreamView = awstypes.StreamView(d.Get("stream_view").(string))
		}

		if d.HasChange(names.AttrVPCConfig) {
			input.VpcConfig = expandVPCConfig(d.Get(names.AttrVPCConfig).([]any))
		}

		_, err := conn.UpdateFleet(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppStream Fleet (%s): %s", d.Id(), err)
		}

		if shouldStop {
			if err := startFleet(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	log.Printf("[DEBUG] Stopping AppStream Fleet: %s", d.Id())
	err := stopFleet(ctx, conn, d.Id())

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting AppStream Fleet: %s", d.Id())
	input := appstream.DeleteFleetInput{
		Name: aws.String(d.Id()),
	}
	_, err = conn.DeleteFleet(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream Fleet (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceFleetCustDiff(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	if diff.HasChange("domain_join_info") {
		o, n := diff.GetChange("domain_join_info")

		if reflect.DeepEqual(expandDomainJoinInfo(o.([]any)), expandDomainJoinInfo(n.([]any))) {
			return diff.Clear("domain_join_info")
		}
	}

	return nil
}

func startFleet(ctx context.Context, conn *appstream.Client, id string) error {
	input := appstream.StartFleetInput{
		Name: aws.String(id),
	}

	_, err := conn.StartFleet(ctx, &input)

	if err != nil {
		return fmt.Errorf("starting AppStream Fleet (%s): %w", id, err)
	}

	if _, err := waitFleetRunning(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for AppStream Fleet (%s) start: %w", id, err)
	}

	return nil
}

func stopFleet(ctx context.Context, conn *appstream.Client, id string) error {
	input := appstream.StopFleetInput{
		Name: aws.String(id),
	}

	_, err := conn.StopFleet(ctx, &input)

	if err != nil {
		return fmt.Errorf("stopping AppStream Fleet (%s): %w", id, err)
	}

	if _, err := waitFleetStopped(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for AppStream Fleet (%s) stop: %w", id, err)
	}

	return nil
}

func findFleetByID(ctx context.Context, conn *appstream.Client, id string) (*awstypes.Fleet, error) {
	input := appstream.DescribeFleetsInput{
		Names: []string{id},
	}

	return findFleet(ctx, conn, &input)
}

func findFleet(ctx context.Context, conn *appstream.Client, input *appstream.DescribeFleetsInput) (*awstypes.Fleet, error) {
	output, err := findFleets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFleets(ctx context.Context, conn *appstream.Client, input *appstream.DescribeFleetsInput) ([]awstypes.Fleet, error) {
	var output []awstypes.Fleet

	err := describeFleetsPages(ctx, conn, input, func(page *appstream.DescribeFleetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Fleets...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusFleet(ctx context.Context, conn *appstream.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFleetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitFleetRunning(ctx context.Context, conn *appstream.Client, id string) (*awstypes.Fleet, error) { //nolint:unparam
	const (
		timeout = 180 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FleetStateStarting),
		Target:  enum.Slice(awstypes.FleetStateRunning),
		Refresh: statusFleet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Fleet); ok {
		tfresource.SetLastError(err, fleetsError(output.FleetErrors))

		return output, err
	}

	return nil, err
}

func waitFleetStopped(ctx context.Context, conn *appstream.Client, id string) (*awstypes.Fleet, error) { //nolint:unparam
	const (
		timeout = 180 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FleetStateStopping),
		Target:  enum.Slice(awstypes.FleetStateStopped),
		Refresh: statusFleet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Fleet); ok {
		tfresource.SetLastError(err, fleetsError(output.FleetErrors))

		return output, err
	}

	return nil, err
}

func fleetError(apiObject *awstypes.FleetError) error {
	if apiObject == nil {
		return nil
	}

	return errs.APIError(apiObject.ErrorCode, aws.ToString(apiObject.ErrorMessage))
}

func fleetsError(apiObjects []awstypes.FleetError) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := fleetError(&apiObject); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func expandComputeCapacity(tfList []any) *awstypes.ComputeCapacity {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.ComputeCapacity{}

	attr := tfList[0].(map[string]any)

	if v, ok := attr["desired_instances"]; ok && v != 0 {
		apiObject.DesiredInstances = aws.Int32(int32(v.(int)))
	}

	if v, ok := attr["desired_sessions"]; ok && v != 0 {
		apiObject.DesiredSessions = aws.Int32(int32(v.(int)))
	}

	if itypes.IsZero(apiObject) {
		return nil
	}

	return apiObject
}

func flattenComputeCapacity(apiObject *awstypes.ComputeCapacityStatus) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Available; v != nil {
		tfMap["available"] = aws.ToInt32(v)
	}

	if v := apiObject.DesiredUserSessions; v != nil {
		tfMap["desired_sessions"] = aws.ToInt32(v)
	}

	// desiredInstances is always returned by the API but cannot be used in conjunction with desiredSessions
	if v := apiObject.Desired; v != nil && tfMap["desired_sessions"] == nil {
		tfMap["desired_instances"] = aws.ToInt32(v)
	}

	if v := apiObject.InUse; v != nil {
		tfMap["in_use"] = aws.ToInt32(v)
	}

	if v := apiObject.Running; v != nil {
		tfMap["running"] = aws.ToInt32(v)
	}

	if len(tfMap) == 0 {
		return nil
	}

	return tfMap
}

func expandDomainJoinInfo(tfList []any) *awstypes.DomainJoinInfo {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.DomainJoinInfo{}

	tfMap, ok := tfList[0].(map[string]any)

	if !ok {
		return nil
	}

	if v, ok := tfMap["directory_name"]; ok && v != "" {
		apiObject.DirectoryName = aws.String(v.(string))
	}

	if v, ok := tfMap["organizational_unit_distinguished_name"]; ok && v != "" {
		apiObject.OrganizationalUnitDistinguishedName = aws.String(v.(string))
	}

	if itypes.IsZero(apiObject) {
		return nil
	}

	return apiObject
}

func flattenDomainInfo(apiObject *awstypes.DomainJoinInfo) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.DirectoryName; v != nil && aws.ToString(v) != "" {
		tfMap["directory_name"] = aws.ToString(v)
	}

	if v := apiObject.OrganizationalUnitDistinguishedName; v != nil && aws.ToString(v) != "" {
		tfMap["organizational_unit_distinguished_name"] = aws.ToString(v)
	}

	if len(tfMap) == 0 {
		return nil
	}

	return tfMap
}

func expandVPCConfig(tfList []any) *awstypes.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.VpcConfig{}

	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap[names.AttrSecurityGroupIDs]; ok {
		apiObject.SecurityGroupIds = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := tfMap[names.AttrSubnetIDs]; ok {
		apiObject.SubnetIds = flex.ExpandStringValueList(v.([]any))
	}

	if itypes.IsZero(apiObject) {
		return nil
	}

	return apiObject
}

func flattenVPCConfig(apiObject *awstypes.VpcConfig) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap[names.AttrSecurityGroupIDs] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = v
	}

	if len(tfMap) == 0 {
		return nil
	}

	return tfMap
}
