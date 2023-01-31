package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceIndexingConfiguration() *schema.Resource {
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
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(iot.FieldType_Values(), false),
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
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(iot.FieldType_Values(), false),
									},
								},
							},
						},
						"thing_group_indexing_mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(iot.ThingGroupIndexingMode_Values(), false),
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
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(iot.FieldType_Values(), false),
									},
								},
							},
						},
						"device_defender_indexing_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      iot.DeviceDefenderIndexingModeOff,
							ValidateFunc: validation.StringInSlice(iot.DeviceDefenderIndexingMode_Values(), false),
						},
						"managed_field": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(iot.FieldType_Values(), false),
									},
								},
							},
						},
						"named_shadow_indexing_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      iot.NamedShadowIndexingModeOff,
							ValidateFunc: validation.StringInSlice(iot.NamedShadowIndexingMode_Values(), false),
						},
						"thing_connectivity_indexing_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      iot.ThingConnectivityIndexingModeOff,
							ValidateFunc: validation.StringInSlice(iot.ThingConnectivityIndexingMode_Values(), false),
						},
						"thing_indexing_mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(iot.ThingIndexingMode_Values(), false),
						},
					},
				},
				AtLeastOneOf: []string{"thing_indexing_configuration", "thing_group_indexing_configuration"},
			},
		},
	}
}

func resourceIndexingConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn()

	input := &iot.UpdateIndexingConfigurationInput{}

	if v, ok := d.GetOk("thing_group_indexing_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ThingGroupIndexingConfiguration = expandThingGroupIndexingConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("thing_indexing_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ThingIndexingConfiguration = expandThingIndexingConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Updating IoT Indexing Configuration: %s", input)
	_, err := conn.UpdateIndexingConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating IoT Indexing Configuration: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return resourceIndexingConfigurationRead(ctx, d, meta)
}

func resourceIndexingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn()

	output, err := conn.GetIndexingConfigurationWithContext(ctx, &iot.GetIndexingConfigurationInput{})

	if err != nil {
		return diag.Errorf("error reading IoT Indexing Configuration: %s", err)
	}

	if output.ThingGroupIndexingConfiguration != nil {
		if err := d.Set("thing_group_indexing_configuration", []interface{}{flattenThingGroupIndexingConfiguration(output.ThingGroupIndexingConfiguration)}); err != nil {
			return diag.Errorf("error setting thing_group_indexing_configuration: %s", err)
		}
	} else {
		d.Set("thing_group_indexing_configuration", nil)
	}
	if output.ThingIndexingConfiguration != nil {
		if err := d.Set("thing_indexing_configuration", []interface{}{flattenThingIndexingConfiguration(output.ThingIndexingConfiguration)}); err != nil {
			return diag.Errorf("error setting thing_indexing_configuration: %s", err)
		}
	} else {
		d.Set("thing_indexing_configuration", nil)
	}

	return nil
}

func flattenThingGroupIndexingConfiguration(apiObject *iot.ThingGroupIndexingConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CustomFields; v != nil {
		tfMap["custom_field"] = flattenFields(v)
	}

	if v := apiObject.ManagedFields; v != nil {
		tfMap["managed_field"] = flattenFields(v)
	}

	if v := apiObject.ThingGroupIndexingMode; v != nil {
		tfMap["thing_group_indexing_mode"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenThingIndexingConfiguration(apiObject *iot.ThingIndexingConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CustomFields; v != nil {
		tfMap["custom_field"] = flattenFields(v)
	}

	if v := apiObject.DeviceDefenderIndexingMode; v != nil {
		tfMap["device_defender_indexing_mode"] = aws.StringValue(v)
	}

	if v := apiObject.ManagedFields; v != nil {
		tfMap["managed_field"] = flattenFields(v)
	}

	if v := apiObject.NamedShadowIndexingMode; v != nil {
		tfMap["named_shadow_indexing_mode"] = aws.StringValue(v)
	}

	if v := apiObject.ThingConnectivityIndexingMode; v != nil {
		tfMap["thing_connectivity_indexing_mode"] = aws.StringValue(v)
	}

	if v := apiObject.ThingIndexingMode; v != nil {
		tfMap["thing_indexing_mode"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenField(apiObject *iot.Field) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenFields(apiObjects []*iot.Field) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenField(apiObject))
	}

	return tfList
}

func expandThingGroupIndexingConfiguration(tfMap map[string]interface{}) *iot.ThingGroupIndexingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &iot.ThingGroupIndexingConfiguration{}

	if v, ok := tfMap["custom_field"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CustomFields = expandFields(v.List())
	}

	if v, ok := tfMap["managed_field"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ManagedFields = expandFields(v.List())
	}

	if v, ok := tfMap["thing_group_indexing_mode"].(string); ok && v != "" {
		apiObject.ThingGroupIndexingMode = aws.String(v)
	}

	return apiObject
}

func expandThingIndexingConfiguration(tfMap map[string]interface{}) *iot.ThingIndexingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &iot.ThingIndexingConfiguration{}

	if v, ok := tfMap["custom_field"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CustomFields = expandFields(v.List())
	}

	if v, ok := tfMap["device_defender_indexing_mode"].(string); ok && v != "" {
		apiObject.DeviceDefenderIndexingMode = aws.String(v)
	}

	if v, ok := tfMap["managed_field"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ManagedFields = expandFields(v.List())
	}

	if v, ok := tfMap["named_shadow_indexing_mode"].(string); ok && v != "" {
		apiObject.NamedShadowIndexingMode = aws.String(v)
	}

	if v, ok := tfMap["thing_connectivity_indexing_mode"].(string); ok && v != "" {
		apiObject.ThingConnectivityIndexingMode = aws.String(v)
	}

	if v, ok := tfMap["thing_indexing_mode"].(string); ok && v != "" {
		apiObject.ThingIndexingMode = aws.String(v)
	}

	return apiObject
}

func expandField(tfMap map[string]interface{}) *iot.Field {
	if tfMap == nil {
		return nil
	}

	apiObject := &iot.Field{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandFields(tfList []interface{}) []*iot.Field {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*iot.Field

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandField(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
