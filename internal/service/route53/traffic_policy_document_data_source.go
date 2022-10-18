package route53

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceTrafficPolicyDocument() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTrafficPolicyDocumentRead,

		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(TrafficPolicyDocEndpointType_values(), false),
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"record_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
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
									"health_check": {
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
									"health_check": {
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
						"location": {
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
									"health_check": {
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
									"health_check": {
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
									"region": {
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
						"region": {
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
									"health_check": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"region": {
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
						"items": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"endpoint_reference": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"health_check": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
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
			"version": {
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
	trafficDoc := &Route53TrafficPolicyDoc{}

	if v, ok := d.GetOk("endpoint"); ok {
		trafficDoc.Endpoints = expandDataTrafficPolicyEndpointsDoc(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("record_type"); ok {
		trafficDoc.RecordType = v.(string)
	}
	if v, ok := d.GetOk("rule"); ok {
		trafficDoc.Rules = expandDataTrafficPolicyRulesDoc(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("start_endpoint"); ok {
		trafficDoc.StartEndpoint = v.(string)
	}
	if v, ok := d.GetOk("start_rule"); ok {
		trafficDoc.StartRule = v.(string)
	}
	if v, ok := d.GetOk("version"); ok {
		trafficDoc.AWSPolicyFormatVersion = v.(string)
	}

	jsonDoc, err := json.Marshal(trafficDoc)
	if err != nil {
		return diag.FromErr(err)
	}
	jsonString := string(jsonDoc)

	d.Set("json", jsonString)

	d.SetId(strconv.Itoa(schema.HashString(jsonString)))

	return nil
}

func expandDataTrafficPolicyEndpointDoc(tfMap map[string]interface{}) *TrafficPolicyEndpoint {
	if tfMap == nil {
		return nil
	}

	apiObject := &TrafficPolicyEndpoint{}

	if v, ok := tfMap["type"]; ok && v.(string) != "" {
		apiObject.Type = v.(string)
	}
	if v, ok := tfMap["region"]; ok && v.(string) != "" {
		apiObject.Region = v.(string)
	}
	if v, ok := tfMap["value"]; ok && v.(string) != "" {
		apiObject.Value = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyEndpointsDoc(tfList []interface{}) map[string]*TrafficPolicyEndpoint {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]*TrafficPolicyEndpoint)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		id := tfMap["id"].(string)

		apiObject := expandDataTrafficPolicyEndpointDoc(tfMap)

		apiObjects[id] = apiObject
	}

	return apiObjects
}

func expandDataTrafficPolicyRuleDoc(tfMap map[string]interface{}) *TrafficPolicyRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &TrafficPolicyRule{}

	if v, ok := tfMap["type"]; ok && v.(string) != "" {
		apiObject.RuleType = v.(string)
	}
	if v, ok := tfMap["primary"]; ok && len(v.([]interface{})) > 0 {
		apiObject.Primary = expandDataTrafficPolicyFailOverDoc(v.([]interface{}))
	}
	if v, ok := tfMap["secondary"]; ok && len(v.([]interface{})) > 0 {
		apiObject.Secondary = expandDataTrafficPolicyFailOverDoc(v.([]interface{}))
	}
	if v, ok := tfMap["location"]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.Locations = expandDataTrafficPolicyLocationsDoc(v.(*schema.Set).List())
	}
	if v, ok := tfMap["geo_proximity_location"]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.GeoProximityLocations = expandDataTrafficPolicyProximitiesDoc(v.(*schema.Set).List())
	}
	if v, ok := tfMap["region"]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.Regions = expandDataTrafficPolicyRegionsDoc(v.(*schema.Set).List())
	}
	if v, ok := tfMap["items"]; ok && len(v.(*schema.Set).List()) > 0 {
		apiObject.Items = expandDataTrafficPolicyItemsDoc(v.(*schema.Set).List())
	}

	return apiObject
}

func expandDataTrafficPolicyRulesDoc(tfList []interface{}) map[string]*TrafficPolicyRule {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]*TrafficPolicyRule)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		id := tfMap["id"].(string)

		apiObject := expandDataTrafficPolicyRuleDoc(tfMap)

		apiObjects[id] = apiObject
	}

	return apiObjects
}

func expandDataTrafficPolicyFailOverDoc(tfList []interface{}) *TrafficPolicyFailoverRule {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, _ := tfList[0].(map[string]interface{})

	apiObject := &TrafficPolicyFailoverRule{}

	if v, ok := tfMap["endpoint_reference"]; ok && v.(string) != "" {
		apiObject.EndpointReference = v.(string)
	}
	if v, ok := tfMap["rule_reference"]; ok && v.(string) != "" {
		apiObject.RuleReference = v.(string)
	}
	if v, ok := tfMap["evaluate_target_health"]; ok && v.(bool) {
		apiObject.EvaluateTargetHealth = aws.Bool(v.(bool))
	}
	if v, ok := tfMap["health_check"]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyLocationDoc(tfMap map[string]interface{}) *TrafficPolicyGeolocationRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &TrafficPolicyGeolocationRule{}

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
	if v, ok := tfMap["health_check"]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyLocationsDoc(tfList []interface{}) []*TrafficPolicyGeolocationRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*TrafficPolicyGeolocationRule

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

func expandDataTrafficPolicyProximityDoc(tfMap map[string]interface{}) *TrafficPolicyGeoproximityRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &TrafficPolicyGeoproximityRule{}

	if v, ok := tfMap["endpoint_reference"]; ok && v.(string) != "" {
		apiObject.EndpointReference = v.(string)
	}
	if v, ok := tfMap["rule_reference"]; ok && v.(string) != "" {
		apiObject.RuleReference = v.(string)
	}
	if v, ok := tfMap["region"]; ok && v.(string) != "" {
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
	if v, ok := tfMap["health_check"]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyProximitiesDoc(tfList []interface{}) []*TrafficPolicyGeoproximityRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*TrafficPolicyGeoproximityRule

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

func expandDataTrafficPolicyRegionDoc(tfMap map[string]interface{}) *TrafficPolicyLatencyRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &TrafficPolicyLatencyRule{}

	if v, ok := tfMap["endpoint_reference"]; ok && v.(string) != "" {
		apiObject.EndpointReference = v.(string)
	}
	if v, ok := tfMap["rule_reference"]; ok && v.(string) != "" {
		apiObject.RuleReference = v.(string)
	}
	if v, ok := tfMap["region"]; ok && v.(string) != "" {
		apiObject.Region = v.(string)
	}
	if v, ok := tfMap["evaluate_target_health"]; ok && v.(bool) {
		apiObject.EvaluateTargetHealth = aws.Bool(v.(bool))
	}
	if v, ok := tfMap["health_check"]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyRegionsDoc(tfList []interface{}) []*TrafficPolicyLatencyRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*TrafficPolicyLatencyRule

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

func expandDataTrafficPolicyItemDoc(tfMap map[string]interface{}) *TrafficPolicyMultiValueAnswerRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &TrafficPolicyMultiValueAnswerRule{}

	if v, ok := tfMap["endpoint_reference"]; ok && v.(string) != "" {
		apiObject.EndpointReference = v.(string)
	}
	if v, ok := tfMap["health_check"]; ok && v.(string) != "" {
		apiObject.HealthCheck = v.(string)
	}

	return apiObject
}

func expandDataTrafficPolicyItemsDoc(tfList []interface{}) []*TrafficPolicyMultiValueAnswerRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*TrafficPolicyMultiValueAnswerRule

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
