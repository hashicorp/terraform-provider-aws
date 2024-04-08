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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codebuild_fleet", name="Fleet")
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

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"compute_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(computeTypeValues(types.ComputeType("").Values()), false),
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(environmentTypeValues(types.EnvironmentType("").Values()), false),
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 128),
			},
			"overflow_behavior": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(fleetOverflowBehaviorValues(types.FleetOverflowBehavior("").Values()), false),
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
						"max_capacity": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"scaling_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"target_tracking_scaling_configs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metric_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(fleetScalingMetricTypeValues(types.FleetScalingMetricType("").Values()), false),
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
			"status": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"context": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_code": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameFleet = "Fleet"
)

func resourceFleetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	input := &codebuild.CreateFleetInput{
		BaseCapacity:    aws.Int32(int32(d.Get("base_capacity").(int))),
		ComputeType:     types.ComputeType(d.Get("compute_type").(string)),
		EnvironmentType: types.EnvironmentType(d.Get("environment_type").(string)),
		Name:            aws.String(d.Get("name").(string)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("overflow_behavior"); ok {
		input.OverflowBehavior = types.FleetOverflowBehavior(v.(string))
	}

	if v, ok := d.GetOk("scaling_configuration"); ok {
		if len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionCreating, ResNameFleet, d.Get("name").(string), errors.New("scaling_configuration cannot be empty"))
		} else {
			input.ScalingConfiguration = expandScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	out, err := conn.CreateFleet(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionCreating, ResNameFleet, d.Get("name").(string), err)
	}

	if out == nil || out.Fleet == nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionCreating, ResNameFleet, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Fleet.Arn))

	if err := waitFleetActive(ctx, conn, d.Id()); err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionWaitingForCreation, ResNameFleet, d.Id(), err)
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	out, err := findFleetByARNOrNames(ctx, conn, d.Id())

	if out == nil || len(out.Fleets) == 0 {
		log.Printf("[WARN] CodeBuild Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionReading, ResNameFleet, d.Id(), err)
	}

	d.Set("arn", out.Fleets[0].Arn)
	d.Set("base_capacity", out.Fleets[0].BaseCapacity)
	d.Set("compute_type", out.Fleets[0].ComputeType)
	d.Set("created", out.Fleets[0].Created.String())
	d.Set("environment_type", out.Fleets[0].EnvironmentType)
	d.Set("id", out.Fleets[0].Id)
	d.Set("last_modified", out.Fleets[0].LastModified.String())
	d.Set("name", out.Fleets[0].Name)
	d.Set("overflow_behavior", out.Fleets[0].OverflowBehavior)
	if len(out.Fleets) > 0 && out.Fleets[0].ScalingConfiguration != nil {
		empty_scaling_configuration := out.Fleets[0].ScalingConfiguration.DesiredCapacity == nil &&
			out.Fleets[0].ScalingConfiguration.MaxCapacity == nil &&
			out.Fleets[0].ScalingConfiguration.ScalingType == "" &&
			out.Fleets[0].ScalingConfiguration.TargetTrackingScalingConfigs == nil

		if !empty_scaling_configuration {
			if err := d.Set("scaling_configuration", []interface{}{flattenScalingConfiguration(out.Fleets[0].ScalingConfiguration)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting overflow behavior: %s", err)
			}
		} else {
			d.Set("scaling_configuration", nil)
		}
	} else {
		d.Set("scaling_configuration", nil)
	}
	if out.Fleets[0].Status != nil {
		if err := d.Set("status", []interface{}{flattenStatus(out.Fleets[0].Status)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting status: %s", err)
		}
	} else {
		d.Set("status", nil)
	}
	setTagsOut(ctx, out.Fleets[0].Tags)
	return diags
}

func resourceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	input := &codebuild.UpdateFleetInput{
		Arn: aws.String(d.Id()),
	}

	if d.HasChange("base_capacity") {
		input.BaseCapacity = aws.Int32(int32(d.Get("base_capacity").(int)))
	}

	if d.HasChange("compute_type") {
		input.ComputeType = types.ComputeType(d.Get("compute_type").(string))
	}

	if d.HasChange("environment_type") {
		input.EnvironmentType = types.EnvironmentType(d.Get("environment_type").(string))
	}

	// Make sure that overflow_behavior is set (if defined) on update - api ommit it on updates
	if v, ok := d.GetOk("overflow_behavior"); ok {
		input.OverflowBehavior = types.FleetOverflowBehavior(v.(string))
	}

	if d.HasChange("scaling_configuration") {
		if v, ok := d.GetOk("scaling_configuration"); ok {
			if len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
				return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionCreating, ResNameFleet, d.Get("name").(string), errors.New("scaling_configuration cannot be empty"))
			} else {
				input.ScalingConfiguration = expandScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
			}
		} else {
			input.ScalingConfiguration = &types.ScalingConfigurationInput{}
		}
	}

	input.Tags = getTagsIn(ctx)

	log.Printf("[DEBUG] Updating CodeBuild Fleet (%s): %#v", d.Id(), input)
	_, err := conn.UpdateFleet(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CodeBuild Fleet (%s): %s", d.Id(), err)
	}

	if err := waitFleetActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CodeBuild Fleet (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	log.Printf("[INFO] Deleting CodeBuild Fleet %s", d.Id())

	_, err := conn.DeleteFleet(ctx, &codebuild.DeleteFleetInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Fleet (%s): %s", d.Id(), err)
	}

	if err := waitFleetDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Fleet (%s): %s", d.Id(), err)
	}

	return diags
}

func waitFleetActive(ctx context.Context, conn *codebuild.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.FleetStatusCodeCreating, types.FleetStatusCodeUpdating, types.FleetStatusCodeDeleting, types.FleetStatusCodeRotating),
		Target:                    enum.Slice(types.FleetStatusCodeActive),
		Refresh:                   statusFleet(ctx, conn, id, false),
		Timeout:                   20 * time.Minute,
		MinTimeout:                15 * time.Second,
		Delay:                     15 * time.Second,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitFleetDeleted(ctx context.Context, conn *codebuild.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.FleetStatusCodeDeleting, "PENDING_DELETION"),
		Target:     []string{},
		Refresh:    statusFleet(ctx, conn, id, true),
		Timeout:    90 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      15 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func statusFleet(ctx context.Context, conn *codebuild.Client, id string, delete bool) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findFleetByARNOrNames(ctx, conn, id)
		if tfresource.NotFound(err) && delete {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Fleets[0].Status.StatusCode)), nil
	}
}

func findFleetByARNOrNames(ctx context.Context, conn *codebuild.Client, arn string) (*codebuild.BatchGetFleetsOutput, error) {
	input := &codebuild.BatchGetFleetsInput{
		Names: []string{arn},
	}
	output, err := conn.BatchGetFleets(ctx, input)

	if output == nil || len(output.Fleets) == 0 || len(output.FleetsNotFound) > 0 {
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

func expandScalingConfiguration(tfMap map[string]interface{}) *types.ScalingConfigurationInput {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &types.ScalingConfigurationInput{}
	if v, ok := tfMap["max_capacity"].(int); ok {
		int32Value := int32(v)
		apiObject.MaxCapacity = &int32Value
	}

	if v, ok := tfMap["scaling_type"].(string); ok && v != "" {
		apiObject.ScalingType = types.FleetScalingType(v)
	}

	if v, ok := tfMap["target_tracking_scaling_configs"].([]interface{}); ok && len(v) > 0 {
		apiObject.TargetTrackingScalingConfigs = expandTargetTrackingScalingConfigs(v)
	}

	return apiObject
}

func expandTargetTrackingScalingConfigs(tfList []interface{}) []types.TargetTrackingScalingConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.TargetTrackingScalingConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
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

func expandTargetTrackingScalingConfig(tfMap map[string]interface{}) *types.TargetTrackingScalingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.TargetTrackingScalingConfiguration{}

	if v, ok := tfMap["metric_type"].(string); ok {
		apiObject.MetricType = types.FleetScalingMetricType(v)
	}

	if v, ok := tfMap["target_value"].(float64); ok {
		apiObject.TargetValue = &v
	}

	return apiObject
}

func flattenScalingConfiguration(apiObject *types.ScalingConfigurationOutput) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if apiObject != nil {
		if v := apiObject.DesiredCapacity; v != nil {
			tfMap["desired_capacity"] = aws.ToInt32(v)
		}

		if v := apiObject.MaxCapacity; v != nil {
			tfMap["max_capacity"] = aws.ToInt32(v)
		}

		if v := apiObject.ScalingType; v != "" {
			tfMap["scaling_type"] = v
		}

		if v := apiObject.TargetTrackingScalingConfigs; v != nil {
			tfMap["target_tracking_scaling_configs"] = flattenTargetTrackingScalingConfigs(v)
		}
	}

	return tfMap
}

func flattenTargetTrackingScalingConfigs(apiObjects []types.TargetTrackingScalingConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfMaps []interface{}

	for _, apiObject := range apiObjects {
		tfMaps = append(tfMaps, flattenTargetTrackingScalingConfig(apiObject))
	}

	return tfMaps
}

func flattenTargetTrackingScalingConfig(apiObject types.TargetTrackingScalingConfiguration) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.MetricType; v != "" {
		tfMap["metric_type"] = v
	}

	if v := apiObject.TargetValue; v != nil {
		tfMap["target_value"] = aws.ToFloat64(v)
	}

	return tfMap
}

func flattenStatus(apiObject *types.FleetStatus) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Context; v != "" {
		tfMap["context"] = v
	}

	if v := apiObject.Message; v != nil {
		tfMap["message"] = aws.ToString(v)
	}

	if v := apiObject.StatusCode; v != "" {
		tfMap["status_code"] = v
	}

	return tfMap
}

func fleetOverflowBehaviorValues(in []types.FleetOverflowBehavior) []string {
	var out []string

	for _, v := range in {
		out = append(out, string(v))
	}

	return out
}

func computeTypeValues(in []types.ComputeType) []string {
	var out []string

	for _, v := range in {
		out = append(out, string(v))
	}

	return out
}

func environmentTypeValues(in []types.EnvironmentType) []string {
	var out []string

	for _, v := range in {
		out = append(out, string(v))
	}

	return out
}

func fleetScalingMetricTypeValues(in []types.FleetScalingMetricType) []string {
	var out []string

	for _, v := range in {
		out = append(out, string(v))
	}

	return out
}
