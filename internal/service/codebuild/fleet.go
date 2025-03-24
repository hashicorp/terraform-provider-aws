// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codebuild_fleet", name="Fleet")
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

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"compute_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"machine_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.MachineType](),
						},
						"memory": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"vcpu": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"compute_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.ComputeType](),
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.EnvironmentType](),
			},
			"fleet_service_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 128),
			},
			"overflow_behavior": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.FleetOverflowBehavior](),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" {
						return true
					}
					return old == new
				},
			},
			"scaling_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrMaxCapacity: {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"scaling_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.FleetScalingType](),
						},
						"target_tracking_scaling_configs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metric_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.FleetScalingMetricType](),
									},
									"target_value": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatAtLeast(0),
									},
								},
							},
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"context": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusCode: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCConfig: {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				RequiredWith: []string{"fleet_service_role"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							MinItems: 1,
							MaxItems: 5,
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							MinItems: 1,
							MaxItems: 16,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

const (
	resNameFleet = "Fleet"
)

func resourceFleetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	input := &codebuild.CreateFleetInput{
		BaseCapacity:    aws.Int32(int32(d.Get("base_capacity").(int))),
		ComputeType:     types.ComputeType(d.Get("compute_type").(string)),
		EnvironmentType: types.EnvironmentType(d.Get("environment_type").(string)),
		Name:            aws.String(d.Get(names.AttrName).(string)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("compute_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ComputeConfiguration = expandComputeConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("fleet_service_role"); ok {
		input.FleetServiceRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_id"); ok {
		input.ImageId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("overflow_behavior"); ok {
		input.OverflowBehavior = types.FleetOverflowBehavior(v.(string))
	}

	if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ScalingConfiguration = expandScalingConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrVPCConfig); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.VpcConfig = expandVPCConfig(v.([]any)[0].(map[string]any))
	}

	// InvalidInputException: CodeBuild is not authorized to perform
	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidInputException](ctx, propagationTimeout, func() (any, error) {
		return conn.CreateFleet(ctx, input)
	}, "ot authorized to perform")

	if err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionCreating, resNameFleet, d.Get(names.AttrName).(string), err)
	}

	d.SetId(aws.ToString(outputRaw.(*codebuild.CreateFleetOutput).Fleet.Arn))

	const (
		timeout = 20 * time.Minute
	)
	if _, err := waitFleetCreated(ctx, conn, d.Id(), timeout); err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionWaitingForCreation, resNameFleet, d.Id(), err)
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	fleet, err := findFleetByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeBuild Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionReading, resNameFleet, d.Id(), err)
	}

	d.Set(names.AttrARN, fleet.Arn)
	d.Set("base_capacity", fleet.BaseCapacity)

	if fleet.ComputeConfiguration != nil {
		if err := d.Set("compute_configuration", []any{flattenComputeConfiguration(fleet.ComputeConfiguration)}); err != nil {
			return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, resNameFleet, d.Id(), err)
		}
	} else {
		d.Set("compute_configuration", nil)
	}

	d.Set("compute_type", fleet.ComputeType)
	d.Set("created", aws.ToTime(fleet.Created).Format(time.RFC3339))
	d.Set("environment_type", fleet.EnvironmentType)
	d.Set("fleet_service_role", fleet.FleetServiceRole)
	d.Set(names.AttrID, fleet.Id)
	d.Set("image_id", fleet.ImageId)
	d.Set("last_modified", aws.ToTime(fleet.LastModified).Format(time.RFC3339))
	d.Set(names.AttrName, fleet.Name)
	d.Set("overflow_behavior", fleet.OverflowBehavior)

	if err := d.Set("scaling_configuration", flattenScalingConfiguration(fleet.ScalingConfiguration)); err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, resNameFleet, d.Id(), err)
	}

	if fleet.Status != nil {
		if err := d.Set(names.AttrStatus, []any{flattenStatus(fleet.Status)}); err != nil {
			return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, resNameFleet, d.Id(), err)
		}
	} else {
		d.Set(names.AttrStatus, nil)
	}

	if err := d.Set(names.AttrVPCConfig, flattenVPCConfig(fleet.VpcConfig)); err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, resNameFleet, d.Id(), err)
	}

	setTagsOut(ctx, fleet.Tags)

	return diags
}

func resourceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	input := &codebuild.UpdateFleetInput{
		Arn: aws.String(d.Id()),
	}

	if d.HasChange("base_capacity") {
		input.BaseCapacity = aws.Int32(int32(d.Get("base_capacity").(int)))
	}

	if d.HasChange("compute_configuration") {
		input.ComputeType = types.ComputeType(d.Get("compute_type").(string))

		if v, ok := d.GetOk("compute_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ComputeConfiguration = expandComputeConfiguration(v.([]any)[0].(map[string]any))
		}
	}

	if d.HasChange("compute_type") {
		input.ComputeType = types.ComputeType(d.Get("compute_type").(string))
	}

	if d.HasChange("environment_type") {
		input.EnvironmentType = types.EnvironmentType(d.Get("environment_type").(string))
	}

	if d.HasChange("fleet_service_role") {
		input.FleetServiceRole = aws.String(d.Get("fleet_service_role").(string))
	}

	if d.HasChange("image_id") {
		input.ImageId = aws.String(d.Get("image_id").(string))
	}

	// Make sure that overflow_behavior is set (if defined) on update - API omits it on updates.
	if v, ok := d.GetOk("overflow_behavior"); ok {
		input.OverflowBehavior = types.FleetOverflowBehavior(v.(string))
	}

	if d.HasChange("scaling_configuration") {
		if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ScalingConfiguration = expandScalingConfiguration(v.([]any)[0].(map[string]any))
		} else {
			input.ScalingConfiguration = &types.ScalingConfigurationInput{}
		}
	}

	if d.HasChange(names.AttrVPCConfig) {
		if v, ok := d.GetOk(names.AttrVPCConfig); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.VpcConfig = expandVPCConfig(v.([]any)[0].(map[string]any))
		}
	}

	input.Tags = getTagsIn(ctx)

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidInputException](ctx, propagationTimeout, func() (any, error) {
		return conn.UpdateFleet(ctx, input)
	}, "ot authorized to perform")

	if err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionUpdating, resNameFleet, d.Id(), err)
	}

	const (
		timeout = 20 * time.Minute
	)
	if _, err := waitFleetUpdated(ctx, conn, d.Id(), timeout); err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionWaitingForUpdate, resNameFleet, d.Id(), err)
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	log.Printf("[INFO] Deleting CodeBuild Fleet: %s", d.Id())
	input := codebuild.DeleteFleetInput{
		Arn: aws.String(d.Id()),
	}
	_, err := conn.DeleteFleet(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionDeleting, resNameFleet, d.Id(), err)
	}

	const (
		timeout = 20 * time.Minute
	)
	if _, err := waitFleetDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionWaitingForDeletion, resNameFleet, d.Id(), err)
	}

	return diags
}

func findFleetByARN(ctx context.Context, conn *codebuild.Client, arn string) (*types.Fleet, error) {
	input := &codebuild.BatchGetFleetsInput{
		Names: []string{arn},
	}

	output, err := findFleet(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if statusCode := output.Status.StatusCode; statusCode == types.FleetStatusCodePendingDeletion {
		return nil, &retry.NotFoundError{
			Message:     string(statusCode),
			LastRequest: input,
		}
	}

	return output, nil
}

func findFleet(ctx context.Context, conn *codebuild.Client, input *codebuild.BatchGetFleetsInput) (*types.Fleet, error) {
	output, err := findFleets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *types.Fleet) bool {
		return v.Status != nil
	})
}

func findFleets(ctx context.Context, conn *codebuild.Client, input *codebuild.BatchGetFleetsInput) ([]types.Fleet, error) {
	output, err := conn.BatchGetFleets(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Fleets, nil
}

func statusFleet(ctx context.Context, conn *codebuild.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFleetByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func waitFleetCreated(ctx context.Context, conn *codebuild.Client, arn string, timeout time.Duration) (*types.Fleet, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.FleetStatusCodeCreating, types.FleetStatusCodeRotating),
		Target:     enum.Slice(types.FleetStatusCodeActive),
		Refresh:    statusFleet(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 15 * time.Second,
		Delay:      15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Fleet); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitFleetUpdated(ctx context.Context, conn *codebuild.Client, arn string, timeout time.Duration) (*types.Fleet, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.FleetStatusCodeUpdating, types.FleetStatusCodeRotating),
		Target:     enum.Slice(types.FleetStatusCodeActive),
		Refresh:    statusFleet(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 15 * time.Second,
		Delay:      15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Fleet); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitFleetDeleted(ctx context.Context, conn *codebuild.Client, arn string, timeout time.Duration) (*types.Fleet, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.FleetStatusCodeDeleting),
		Target:     []string{},
		Refresh:    statusFleet(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 15 * time.Second,
		Delay:      15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Fleet); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func expandComputeConfiguration(tfMap map[string]any) *types.ComputeConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ComputeConfiguration{}

	if v, ok := tfMap["disk"].(int); ok {
		apiObject.Disk = aws.Int64(int64(v))
	}

	if v, ok := tfMap["machine_type"].(string); ok && v != "" {
		apiObject.MachineType = types.MachineType(v)
	}

	if v, ok := tfMap["memory"].(int); ok {
		apiObject.Memory = aws.Int64(int64(v))
	}

	if v, ok := tfMap["vcpu"].(int); ok {
		apiObject.VCpu = aws.Int64(int64(v))
	}

	return apiObject
}

func expandScalingConfiguration(tfMap map[string]any) *types.ScalingConfigurationInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ScalingConfigurationInput{}

	if v, ok := tfMap[names.AttrMaxCapacity].(int); ok {
		apiObject.MaxCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["scaling_type"].(string); ok && v != "" {
		apiObject.ScalingType = types.FleetScalingType(v)
	}

	if v, ok := tfMap["target_tracking_scaling_configs"].([]any); ok && len(v) > 0 {
		apiObject.TargetTrackingScalingConfigs = expandTargetTrackingScalingConfigs(v)
	}

	return apiObject
}

func expandTargetTrackingScalingConfigs(tfList []any) []types.TargetTrackingScalingConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.TargetTrackingScalingConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandTargetTrackingScalingConfig(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}
	return apiObjects
}

func expandTargetTrackingScalingConfig(tfMap map[string]any) *types.TargetTrackingScalingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.TargetTrackingScalingConfiguration{}

	if v, ok := tfMap["metric_type"].(string); ok {
		apiObject.MetricType = types.FleetScalingMetricType(v)
	}

	if v, ok := tfMap["target_value"].(float64); ok {
		apiObject.TargetValue = aws.Float64(v)
	}

	return apiObject
}

func flattenComputeConfiguration(apiObject *types.ComputeConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Disk; v != nil {
		tfMap["disk"] = aws.ToInt64(v)
	}

	if v := apiObject.MachineType; v != "" {
		tfMap["machine_type"] = v
	}

	if v := apiObject.Memory; v != nil {
		tfMap["memory"] = aws.ToInt64(v)
	}

	if v := apiObject.VCpu; v != nil {
		tfMap["vcpu"] = aws.ToInt64(v)
	}

	return tfMap
}

func flattenScalingConfiguration(apiObject *types.ScalingConfigurationOutput) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.DesiredCapacity; v != nil {
		tfMap["desired_capacity"] = aws.ToInt32(v)
	}

	if v := apiObject.MaxCapacity; v != nil {
		tfMap[names.AttrMaxCapacity] = aws.ToInt32(v)
	}

	if v := apiObject.ScalingType; v != "" {
		tfMap["scaling_type"] = v
	}

	if v := apiObject.TargetTrackingScalingConfigs; v != nil {
		tfMap["target_tracking_scaling_configs"] = flattenTargetTrackingScalingConfigs(v)
	}

	if len(tfMap) == 0 {
		return nil
	}

	return []any{tfMap}
}

func flattenTargetTrackingScalingConfigs(apiObjects []types.TargetTrackingScalingConfiguration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenTargetTrackingScalingConfig(&apiObject))
	}

	return tfList
}

func flattenTargetTrackingScalingConfig(apiObject *types.TargetTrackingScalingConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.MetricType; v != "" {
		tfMap["metric_type"] = v
	}

	if v := apiObject.TargetValue; v != nil {
		tfMap["target_value"] = aws.ToFloat64(v)
	}

	return tfMap
}

func flattenStatus(apiObject *types.FleetStatus) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Context; v != "" {
		tfMap["context"] = v
	}

	if v := apiObject.Message; v != nil {
		tfMap[names.AttrMessage] = aws.ToString(v)
	}

	if v := apiObject.StatusCode; v != "" {
		tfMap[names.AttrStatusCode] = v
	}

	return tfMap
}
