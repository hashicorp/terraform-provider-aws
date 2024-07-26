// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_configuration_recorder", name="Configuration Recorder")
func resourceConfigurationRecorder() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationRecorderPut,
		ReadWithoutTimeout:   resourceConfigurationRecorderRead,
		UpdateWithoutTimeout: resourceConfigurationRecorderPut,
		DeleteWithoutTimeout: resourceConfigurationRecorderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: resourceConfigurationRecorderCustomizeDiff,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      defaultConfigurationRecorderName,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"recording_group": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all_supported": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"exclusion_by_resource_types": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_types": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"include_global_resource_types": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"recording_strategy": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"use_only": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.RecordingStrategyType](),
									},
								},
							},
						},
						"resource_types": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"recording_mode": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recording_frequency": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.RecordingFrequencyContinuous,
							ValidateDiagFunc: enum.Validate[types.RecordingFrequency](),
						},
						"recording_mode_override": {
							Type:     schema.TypeList,
							Optional: true,
							// Even though the name is plural, the API only allows one override:
							// ValidationException: 1 validation error detected: Value '[com.amazonaws.starling.dove.RecordingModeOverride@aa179030, com.amazonaws.starling.dove.RecordingModeOverride@4b13c61c]' at 'configurationRecorder.recordingMode.recordingModeOverrides' failed to satisfy constraint: Member must have length less than or equal to 1
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDescription: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"recording_frequency": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.RecordingFrequency](),
									},
									"resource_types": {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceConfigurationRecorderPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &configservice.PutConfigurationRecorderInput{
		ConfigurationRecorder: &types.ConfigurationRecorder{
			Name:    aws.String(name),
			RoleARN: aws.String(d.Get(names.AttrRoleARN).(string)),
		},
	}

	if v, ok := d.GetOk("recording_group"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ConfigurationRecorder.RecordingGroup = expandRecordingGroup(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("recording_mode"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ConfigurationRecorder.RecordingMode = expandRecordingMode(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.PutConfigurationRecorder(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ConfigService Configuration Recorder (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceConfigurationRecorderRead(ctx, d, meta)...)
}

func resourceConfigurationRecorderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	recorder, err := findConfigurationRecorderByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Configuration Recorder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Configuration Recorder (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, recorder.Name)
	if recorder.RecordingGroup != nil {
		if err := d.Set("recording_group", flattenRecordingGroup(recorder.RecordingGroup)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting recording_group: %s", err)
		}
	}
	if recorder.RecordingMode != nil {
		if err := d.Set("recording_mode", flattenRecordingMode(recorder.RecordingMode)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting recording_mode: %s", err)
		}
	}
	d.Set(names.AttrRoleARN, recorder.RoleARN)

	return diags
}

func resourceConfigurationRecorderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	log.Printf("[DEBUG] Deleting ConfigService Configuration Recorder: %s", d.Id())
	_, err := conn.DeleteConfigurationRecorder(ctx, &configservice.DeleteConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(d.Id()),
	})

	if errs.IsA[*types.NoSuchConfigurationRecorderException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Configuration Recorder (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceConfigurationRecorderCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" { // New resource.
		if v, ok := diff.GetOk("recording_group"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if h, ok := tfMap["all_supported"]; ok {
				if i, ok := tfMap["recording_strategy"]; ok && len(i.([]interface{})) > 0 && i.([]interface{})[0] != nil {
					strategy := i.([]interface{})[0].(map[string]interface{})

					if j, ok := strategy["use_only"].(string); ok {
						if h.(bool) && j != string(types.RecordingStrategyTypeAllSupportedResourceTypes) {
							return errors.New(` Invalid record group strategy  , all_supported must be set to true  `)
						}

						if k, ok := tfMap["exclusion_by_resource_types"]; ok && len(k.([]interface{})) > 0 && k.([]interface{})[0] != nil {
							if h.(bool) {
								return errors.New(` Invalid record group , all_supported must be set to false when exclusion_by_resource_types is set `)
							}

							if j != string(types.RecordingStrategyTypeExclusionByResourceTypes) {
								return errors.New(` Invalid record group strategy ,  use only must be set to EXCLUSION_BY_RESOURCE_TYPES`)
							}

							if l, ok := tfMap["resource_types"]; ok {
								resourceTypes := flex.ExpandStringSet(l.(*schema.Set))
								if len(resourceTypes) > 0 {
									return errors.New(` Invalid record group , resource_types must not be set when exclusion_by_resource_types is set `)
								}
							}
						}

						if l, ok := tfMap["resource_types"]; ok {
							resourceTypes := flex.ExpandStringSet(l.(*schema.Set))
							if len(resourceTypes) > 0 {
								if h.(bool) {
									return errors.New(` Invalid record group , all_supported must be set to false when resource_types is set `)
								}

								if j != string(types.RecordingStrategyTypeInclusionByResourceTypes) {
									return errors.New(` Invalid record group strategy ,  use only must be set to INCLUSION_BY_RESOURCE_TYPES`)
								}

								if m, ok := tfMap["exclusion_by_resource_types"]; ok && len(m.([]interface{})) > 0 && i.([]interface{})[0] != nil {
									return errors.New(` Invalid record group , exclusion_by_resource_types must not be set when resource_types is set `)
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func findConfigurationRecorderByName(ctx context.Context, conn *configservice.Client, name string) (*types.ConfigurationRecorder, error) {
	input := &configservice.DescribeConfigurationRecordersInput{
		ConfigurationRecorderNames: []string{name},
	}

	return findConfigurationRecorder(ctx, conn, input)
}

func findConfigurationRecorder(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConfigurationRecordersInput) (*types.ConfigurationRecorder, error) {
	output, err := findConfigurationRecorders(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConfigurationRecorders(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConfigurationRecordersInput) ([]types.ConfigurationRecorder, error) {
	output, err := conn.DescribeConfigurationRecorders(ctx, input)

	if errs.IsA[*types.NoSuchConfigurationRecorderException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ConfigurationRecorders, nil
}

func expandRecordingGroup(tfMap map[string]interface{}) *types.RecordingGroup {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.RecordingGroup{}

	if v, ok := tfMap["all_supported"]; ok {
		apiObject.AllSupported = v.(bool)
	}

	if v, ok := tfMap["exclusion_by_resource_types"]; ok && len(v.([]interface{})) > 0 {
		apiObject.ExclusionByResourceTypes = expandExclusionByResourceTypes(v.([]interface{}))
	}

	if v, ok := tfMap["include_global_resource_types"]; ok {
		apiObject.IncludeGlobalResourceTypes = v.(bool)
	}

	if v, ok := tfMap["recording_strategy"]; ok && len(v.([]interface{})) > 0 {
		apiObject.RecordingStrategy = expandRecordingStrategy(v.([]interface{}))
	}

	if v, ok := tfMap["resource_types"]; ok && v.(*schema.Set).Len() > 0 {
		apiObject.ResourceTypes = flex.ExpandStringyValueSet[types.ResourceType](v.(*schema.Set))
	}

	return apiObject
}

func expandExclusionByResourceTypes(tfList []interface{}) *types.ExclusionByResourceTypes {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.ExclusionByResourceTypes{}

	if v, ok := tfMap["resource_types"]; ok && v.(*schema.Set).Len() > 0 {
		apiObject.ResourceTypes = flex.ExpandStringyValueSet[types.ResourceType](v.(*schema.Set))
	}

	return apiObject
}

func expandRecordingStrategy(tfList []interface{}) *types.RecordingStrategy {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.RecordingStrategy{}

	if v, ok := tfMap["use_only"].(string); ok {
		apiObject.UseOnly = types.RecordingStrategyType(v)
	}

	return apiObject
}

func expandRecordingMode(tfMap map[string]interface{}) *types.RecordingMode {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.RecordingMode{}

	if v, ok := tfMap["recording_frequency"].(string); ok {
		apiObject.RecordingFrequency = types.RecordingFrequency(v)
	}

	if v, ok := tfMap["recording_mode_override"]; ok && len(v.([]interface{})) > 0 {
		apiObject.RecordingModeOverrides = expandRecordingModeOverride(v.([]interface{}))
	}

	return apiObject
}

func expandRecordingModeOverride(tfList []interface{}) []types.RecordingModeOverride {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.RecordingModeOverride

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.RecordingModeOverride{}

		if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
			apiObject.Description = aws.String(v)
		}

		if v, ok := tfMap["recording_frequency"].(string); ok {
			apiObject.RecordingFrequency = types.RecordingFrequency(v)
		}

		if v, ok := tfMap["resource_types"]; ok && v.(*schema.Set).Len() > 0 {
			apiObject.ResourceTypes = flex.ExpandStringyValueSet[types.ResourceType](v.(*schema.Set))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenRecordingGroup(apiObject *types.RecordingGroup) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"all_supported":                 apiObject.AllSupported,
		"include_global_resource_types": apiObject.IncludeGlobalResourceTypes,
	}

	if apiObject.ExclusionByResourceTypes != nil {
		tfMap["exclusion_by_resource_types"] = flattenExclusionByResourceTypes(apiObject.ExclusionByResourceTypes)
	}

	if apiObject.RecordingStrategy != nil {
		tfMap["recording_strategy"] = flattenRecordingStrategy(apiObject.RecordingStrategy)
	}

	if apiObject.ResourceTypes != nil {
		tfMap["resource_types"] = apiObject.ResourceTypes
	}

	return []interface{}{tfMap}
}

func flattenExclusionByResourceTypes(apiObject *types.ExclusionByResourceTypes) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ResourceTypes != nil {
		tfMap["resource_types"] = apiObject.ResourceTypes
	}

	return []interface{}{tfMap}
}

func flattenRecordingStrategy(apiObject *types.RecordingStrategy) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"use_only": apiObject.UseOnly,
	}

	return []interface{}{tfMap}
}

func flattenRecordingMode(apiObject *types.RecordingMode) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"recording_frequency": apiObject.RecordingFrequency,
	}

	if apiObject.RecordingModeOverrides != nil && len(apiObject.RecordingModeOverrides) > 0 {
		tfMap["recording_mode_override"] = flattenRecordingModeOverrides(apiObject.RecordingModeOverrides)
	}

	return []interface{}{tfMap}
}

func flattenRecordingModeOverrides(apiObjects []types.RecordingModeOverride) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		m := map[string]interface{}{
			names.AttrDescription: aws.ToString(apiObject.Description),
			"recording_frequency": apiObject.RecordingFrequency,
			"resource_types":      apiObject.ResourceTypes,
		}

		tfList = append(tfList, m)
	}

	return tfList
}
