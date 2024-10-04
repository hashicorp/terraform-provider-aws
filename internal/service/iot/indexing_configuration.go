// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_indexing_configuration", name="Indexing Configuration")
func resourceIndexingConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIndexingConfigurationPut,
		ReadWithoutTimeout:   resourceIndexingConfigurationRead,
		UpdateWithoutTimeout: resourceIndexingConfigurationPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"thing_group_indexing_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_field": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.FieldType](),
									},
								},
							},
						},
						"managed_field": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.FieldType](),
									},
								},
							},
						},
						"thing_group_indexing_mode": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ThingGroupIndexingMode](),
						},
					},
				},
				AtLeastOneOf: []string{"thing_group_indexing_configuration", "thing_indexing_configuration"},
			},
			"thing_indexing_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_field": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.FieldType](),
									},
								},
							},
						},
						"device_defender_indexing_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.DeviceDefenderIndexingModeOff,
							ValidateDiagFunc: enum.Validate[awstypes.DeviceDefenderIndexingMode](),
						},
						names.AttrFilter: {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"named_shadow_names": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 64),
												validation.StringMatch(regexache.MustCompile(`^[$a-zA-Z0-9:_-]+`), "must contain only alphanumeric characters, underscores, colons, and hyphens (^[$a-zA-Z0-9:_-]+)"),
											),
										},
									},
								},
							},
						},
						"managed_field": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.FieldType](),
									},
								},
							},
						},
						"named_shadow_indexing_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.NamedShadowIndexingModeOff,
							ValidateDiagFunc: enum.Validate[awstypes.NamedShadowIndexingMode](),
						},
						"thing_connectivity_indexing_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.ThingConnectivityIndexingModeOff,
							ValidateDiagFunc: enum.Validate[awstypes.ThingConnectivityIndexingMode](),
						},
						"thing_indexing_mode": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ThingIndexingMode](),
						},
					},
				},
				AtLeastOneOf: []string{"thing_indexing_configuration", "thing_group_indexing_configuration"},
			},
		},
	}
}

func resourceIndexingConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	input := &iot.UpdateIndexingConfigurationInput{}

	if v, ok := d.GetOk("thing_group_indexing_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ThingGroupIndexingConfiguration = expandThingGroupIndexingConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("thing_indexing_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ThingIndexingConfiguration = expandThingIndexingConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.UpdateIndexingConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IoT Indexing Configuration: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).Region)
	}

	return append(diags, resourceIndexingConfigurationRead(ctx, d, meta)...)
}

func resourceIndexingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := conn.GetIndexingConfiguration(ctx, &iot.GetIndexingConfigurationInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Indexing Configuration (%s): %s", d.Id(), err)
	}

	if output.ThingGroupIndexingConfiguration != nil {
		if err := d.Set("thing_group_indexing_configuration", []interface{}{flattenThingGroupIndexingConfiguration(output.ThingGroupIndexingConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting thing_group_indexing_configuration: %s", err)
		}
	} else {
		d.Set("thing_group_indexing_configuration", nil)
	}
	if output.ThingIndexingConfiguration != nil {
		if err := d.Set("thing_indexing_configuration", []interface{}{flattenThingIndexingConfiguration(output.ThingIndexingConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting thing_indexing_configuration: %s", err)
		}
	} else {
		d.Set("thing_indexing_configuration", nil)
	}

	return diags
}

func flattenThingGroupIndexingConfiguration(apiObject *awstypes.ThingGroupIndexingConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"thing_group_indexing_mode": apiObject.ThingGroupIndexingMode,
	}

	if v := apiObject.CustomFields; v != nil {
		tfMap["custom_field"] = flattenFields(v)
	}

	if v := apiObject.ManagedFields; v != nil {
		tfMap["managed_field"] = flattenFields(v)
	}

	return tfMap
}

func flattenThingIndexingConfiguration(apiObject *awstypes.ThingIndexingConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"device_defender_indexing_mode":    apiObject.DeviceDefenderIndexingMode,
		"named_shadow_indexing_mode":       apiObject.NamedShadowIndexingMode,
		"thing_connectivity_indexing_mode": apiObject.ThingConnectivityIndexingMode,
		"thing_indexing_mode":              apiObject.ThingIndexingMode,
	}

	if v := apiObject.CustomFields; v != nil {
		tfMap["custom_field"] = flattenFields(v)
	}

	if v := apiObject.Filter; v != nil {
		tfMap[names.AttrFilter] = []interface{}{flattenIndexingFilter(v)}
	}

	if v := apiObject.ManagedFields; v != nil {
		tfMap["managed_field"] = flattenFields(v)
	}

	return tfMap
}

func flattenIndexingFilter(apiObject *awstypes.IndexingFilter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.NamedShadowNames; v != nil {
		tfMap["named_shadow_names"] = aws.StringSlice(v)
	}

	return tfMap
}

func flattenField(apiObject awstypes.Field) map[string]interface{} {
	tfMap := map[string]interface{}{
		names.AttrType: apiObject.Type,
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func flattenFields(apiObjects []awstypes.Field) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenField(apiObject))
	}

	return tfList
}

func expandThingGroupIndexingConfiguration(tfMap map[string]interface{}) *awstypes.ThingGroupIndexingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ThingGroupIndexingConfiguration{}

	if v, ok := tfMap["custom_field"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CustomFields = expandFields(v.List())
	}

	if v, ok := tfMap["managed_field"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ManagedFields = expandFields(v.List())
	}

	if v, ok := tfMap["thing_group_indexing_mode"].(string); ok && v != "" {
		apiObject.ThingGroupIndexingMode = awstypes.ThingGroupIndexingMode(v)
	}

	return apiObject
}

func expandThingIndexingConfiguration(tfMap map[string]interface{}) *awstypes.ThingIndexingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ThingIndexingConfiguration{}

	if v, ok := tfMap["custom_field"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CustomFields = expandFields(v.List())
	}

	if v, ok := tfMap["device_defender_indexing_mode"].(string); ok && v != "" {
		apiObject.DeviceDefenderIndexingMode = awstypes.DeviceDefenderIndexingMode(v)
	}

	if v, ok := tfMap[names.AttrFilter]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Filter = expandIndexingFilter(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["managed_field"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ManagedFields = expandFields(v.List())
	}

	if v, ok := tfMap["named_shadow_indexing_mode"].(string); ok && v != "" {
		apiObject.NamedShadowIndexingMode = awstypes.NamedShadowIndexingMode(v)
	}

	if v, ok := tfMap["thing_connectivity_indexing_mode"].(string); ok && v != "" {
		apiObject.ThingConnectivityIndexingMode = awstypes.ThingConnectivityIndexingMode(v)
	}

	if v, ok := tfMap["thing_indexing_mode"].(string); ok && v != "" {
		apiObject.ThingIndexingMode = awstypes.ThingIndexingMode(v)
	}

	return apiObject
}

func expandIndexingFilter(tfMap map[string]interface{}) *awstypes.IndexingFilter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.IndexingFilter{}

	if v, ok := tfMap["named_shadow_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NamedShadowNames = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandField(tfMap map[string]interface{}) *awstypes.Field {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Field{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.FieldType(v)
	}

	return apiObject
}

func expandFields(tfList []interface{}) []awstypes.Field {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Field

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandField(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}
