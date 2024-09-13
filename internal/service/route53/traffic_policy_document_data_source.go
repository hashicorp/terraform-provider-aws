// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_traffic_policy_document", name="Traffic Policy Document")
func dataSourceTrafficPolicyDocument() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTrafficPolicyDocumentRead,

		Schema: map[string]*schema.Schema{
			names.AttrEndpoint: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrRegion: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[trafficPolicyDocEndpointType](),
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrJSON: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"record_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrRule: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"geo_proximity_location": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bias": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"endpoint_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"evaluate_target_health": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									names.AttrHealthCheck: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"latitude": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"longitude": {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"rule_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Required: true,
						},
						"items": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"endpoint_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrHealthCheck: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrLocation: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"continent": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"country": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"endpoint_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"evaluate_target_health": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									names.AttrHealthCheck: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"is_default": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"rule_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"subdivision": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"primary": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"endpoint_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"evaluate_target_health": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									names.AttrHealthCheck: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"rule_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrRegion: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"endpoint_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"evaluate_target_health": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									names.AttrHealthCheck: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"rule_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"secondary": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"endpoint_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"evaluate_target_health": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									names.AttrHealthCheck: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"rule_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"start_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"start_rule": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2015-10-01",
				ValidateFunc: validation.StringInSlice([]string{
					"2015-10-01",
				}, false),
			},
		},
	}
}

func dataSourceTrafficPolicyDocumentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	trafficDoc := &route53TrafficPolicyDoc{}

	if v, ok := d.GetOk(names.AttrEndpoint); ok {
		trafficDoc.Endpoints = expandDataTrafficPolicyEndpointsDoc(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("record_type"); ok {
		trafficDoc.RecordType = v.(string)
	}
	if v, ok := d.GetOk(names.AttrRule); ok {
		trafficDoc.Rules = expandDataTrafficPolicyRulesDoc(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("start_endpoint"); ok {
		trafficDoc.StartEndpoint = v.(string)
	}
	if v, ok := d.GetOk("start_rule"); ok {
		trafficDoc.StartRule = v.(string)
	}
	if v, ok := d.GetOk(names.AttrVersion); ok {
		trafficDoc.AWSPolicyFormatVersion = v.(string)
	}

	jsonDoc, err := json.Marshal(trafficDoc)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	jsonString := string(jsonDoc)

	d.Set(names.AttrJSON, jsonString)

	d.SetId(strconv.Itoa(schema.HashString(jsonString)))

	return diags
}

func expandDataTrafficPolicyEndpointDoc(tfMap map[string]interface{}) *trafficPolicyEndpoint {
	if tfMap == nil {
		return nil
	}

	apiObject := &trafficPolicyEndpoint{}

	if v, ok := tfMap[names.AttrType]; ok && v.(string) != "" {
		apiObject.Type = v.(string)
	}
	if v, ok := tfMap[names.AttrRegion]; ok && v.(string) != "" {
		apiObject.Region = v.(string)
	}
	if v, ok := tfMap[names.AttrValue]; ok && v.(string) != "" {
		apiObject.Value = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyEndpointsDoc(tfList []interface{}) map[string]*trafficPolicyEndpoint {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]*trafficPolicyEndpoint)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		id := tfMap[names.AttrID].(string)

		apiObject := expandDataTrafficPolicyEndpointDoc(tfMap)

		apiObjects[id] = apiObject
	}

	return apiObjects
}

func expandDataTrafficPolicyRuleDoc(tfMap map[string]interface{}) *trafficPolicyRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &trafficPolicyRule{}

	if v, ok := tfMap[names.AttrType]; ok && v.(string) != "" {
		apiObject.RuleType = v.(string)
	}
	if v, ok := tfMap["primary"]; ok && len(v.([]interface{})) > 0 {
		apiObject.Primary = expandDataTrafficPolicyFailOverDoc(v.([]interface{}))
	}
	if v, ok := tfMap["secondary"]; ok && len(v.([]interface{})) > 0 {
		apiObject.Secondary = expandDataTrafficPolicyFailOverDoc(v.([]interface{}))
	}
	if v, ok := tfMap[names.AttrLocation]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.Locations = expandDataTrafficPolicyLocationsDoc(v.(*schema.Set).List())
	}
	if v, ok := tfMap["geo_proximity_location"]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.GeoProximityLocations = expandDataTrafficPolicyProximitiesDoc(v.(*schema.Set).List())
	}
	if v, ok := tfMap[names.AttrRegion]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.Regions = expandDataTrafficPolicyRegionsDoc(v.(*schema.Set).List())
	}
	if v, ok := tfMap["items"]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.Items = expandDataTrafficPolicyItemsDoc(v.(*schema.Set).List())
	}

	return apiObject
}

func expandDataTrafficPolicyRulesDoc(tfList []interface{}) map[string]*trafficPolicyRule {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]*trafficPolicyRule)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		id := tfMap[names.AttrID].(string)

		apiObject := expandDataTrafficPolicyRuleDoc(tfMap)

		apiObjects[id] = apiObject
	}

	return apiObjects
}

func expandDataTrafficPolicyFailOverDoc(tfList []interface{}) *trafficPolicyFailoverRule {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, _ := tfList[0].(map[string]interface{})

	apiObject := &trafficPolicyFailoverRule{}

	if v, ok := tfMap["endpoint_reference"]; ok && v.(string) != "" {
		apiObject.EndpointReference = v.(string)
	}
	if v, ok := tfMap["rule_reference"]; ok && v.(string) != "" {
		apiObject.RuleReference = v.(string)
	}
	if v, ok := tfMap["evaluate_target_health"]; ok && v.(bool) {
		apiObject.EvaluateTargetHealth = aws.Bool(v.(bool))
	}
	if v, ok := tfMap[names.AttrHealthCheck]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyLocationDoc(tfMap map[string]interface{}) *trafficPolicyGeolocationRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &trafficPolicyGeolocationRule{}

	if v, ok := tfMap["endpoint_reference"]; ok && v.(string) != "" {
		apiObject.EndpointReference = v.(string)
	}
	if v, ok := tfMap["rule_reference"]; ok && v.(string) != "" {
		apiObject.RuleReference = v.(string)
	}
	if v, ok := tfMap["is_default"]; ok && v.(bool) {
		apiObject.IsDefault = aws.Bool(v.(bool))
	}
	if v, ok := tfMap["continent"]; ok && v.(string) != "" {
		apiObject.Continent = v.(string)
	}
	if v, ok := tfMap["country"]; ok && v.(string) != "" {
		apiObject.Country = v.(string)
	}
	if v, ok := tfMap["subdivision"]; ok && v.(string) != "" {
		apiObject.Subdivision = v.(string)
	}
	if v, ok := tfMap["evaluate_target_health"]; ok && v.(bool) {
		apiObject.EvaluateTargetHealth = aws.Bool(v.(bool))
	}
	if v, ok := tfMap[names.AttrHealthCheck]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyLocationsDoc(tfList []interface{}) []*trafficPolicyGeolocationRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*trafficPolicyGeolocationRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDataTrafficPolicyLocationDoc(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDataTrafficPolicyProximityDoc(tfMap map[string]interface{}) *trafficPolicyGeoproximityRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &trafficPolicyGeoproximityRule{}

	if v, ok := tfMap["endpoint_reference"]; ok && v.(string) != "" {
		apiObject.EndpointReference = v.(string)
	}
	if v, ok := tfMap["rule_reference"]; ok && v.(string) != "" {
		apiObject.RuleReference = v.(string)
	}
	if v, ok := tfMap[names.AttrRegion]; ok && v.(string) != "" {
		apiObject.Region = v.(string)
	}
	if v, ok := tfMap["latitude"]; ok && v.(string) != "" {
		apiObject.Latitude = v.(string)
	}
	if v, ok := tfMap["longitude"]; ok && v.(string) != "" {
		apiObject.Longitude = v.(string)
	}
	if v, ok := tfMap["bias"]; ok && v.(string) != "" {
		apiObject.Bias = v.(string)
	}
	if v, ok := tfMap["evaluate_target_health"]; ok && v.(bool) {
		apiObject.EvaluateTargetHealth = aws.Bool(v.(bool))
	}
	if v, ok := tfMap[names.AttrHealthCheck]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyProximitiesDoc(tfList []interface{}) []*trafficPolicyGeoproximityRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*trafficPolicyGeoproximityRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDataTrafficPolicyProximityDoc(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDataTrafficPolicyRegionDoc(tfMap map[string]interface{}) *trafficPolicyLatencyRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &trafficPolicyLatencyRule{}

	if v, ok := tfMap["endpoint_reference"]; ok && v.(string) != "" {
		apiObject.EndpointReference = v.(string)
	}
	if v, ok := tfMap["rule_reference"]; ok && v.(string) != "" {
		apiObject.RuleReference = v.(string)
	}
	if v, ok := tfMap[names.AttrRegion]; ok && v.(string) != "" {
		apiObject.Region = v.(string)
	}
	if v, ok := tfMap["evaluate_target_health"]; ok && v.(bool) {
		apiObject.EvaluateTargetHealth = aws.Bool(v.(bool))
	}
	if v, ok := tfMap[names.AttrHealthCheck]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyRegionsDoc(tfList []interface{}) []*trafficPolicyLatencyRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*trafficPolicyLatencyRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDataTrafficPolicyRegionDoc(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDataTrafficPolicyItemDoc(tfMap map[string]interface{}) *trafficPolicyMultiValueAnswerRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &trafficPolicyMultiValueAnswerRule{}

	if v, ok := tfMap["endpoint_reference"]; ok && v.(string) != "" {
		apiObject.EndpointReference = v.(string)
	}
	if v, ok := tfMap[names.AttrHealthCheck]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyItemsDoc(tfList []interface{}) []*trafficPolicyMultiValueAnswerRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*trafficPolicyMultiValueAnswerRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDataTrafficPolicyItemDoc(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
