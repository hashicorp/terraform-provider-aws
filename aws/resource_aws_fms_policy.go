package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mitchellh/mapstructure"
)

func resourceAwsFmsPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFmsPolicyCreate,
		Read:   resourceAwsFmsPolicyRead,
		Update: resourceAwsFmsPolicyUpdate,
		Delete: resourceAwsFmsPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"delete_all_policy_resources": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"exclude_resource_tags": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},

			"exclude_map": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"orgunit": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},

			"include_map": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"orgunit": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},

			"remediation_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},

			"resource_type_list": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"AWS::ApiGateway::Stage", "AWS::ElasticLoadBalancingV2::LoadBalancer", "AWS::CloudFront::Distribution"}, false),
				},
				Set: schema.HashString,
			},

			"policy_update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_tags": tagsSchema(),

			"security_service_policy_data": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"waf":   wafSchema(),
						"wafv2": wafV2Schema(),
						"shield_advanced": {
							Type:         schema.TypeBool,
							ExactlyOneOf: []string{"security_service_policy_data.0.waf", "security_service_policy_data.0.wafv2", "security_service_policy_data.0.shield_advanced", "security_service_policy_data.0.security_groups_common", "security_service_policy_data.0.security_groups_content_audit", "security_service_policy_data.0.security_groups_usage_audit"},
							Optional:     true,
						},
						"security_groups_common":        securityGroupsCommon(),
						"security_groups_content_audit": securityGroupsContentAudit(),
						"security_groups_usage_audit":   securityGroupsUsageAudit(),
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func securityGroupsCommon() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeList,
		MaxItems:     1,
		ExactlyOneOf: []string{"security_service_policy_data.0.waf", "security_service_policy_data.0.wafv2", "security_service_policy_data.0.shield_advanced", "security_service_policy_data.0.security_groups_common", "security_service_policy_data.0.security_groups_content_audit", "security_service_policy_data.0.security_groups_usage_audit"},
		Optional:     true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"revert_manual_security_group_changes": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"exclusive_resource_security_group_management": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"security_groups": {
					Type:     schema.TypeSet,
					Required: true,
					Elem:     schema.TypeString,
				},
				"remediation_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"resource_type": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func securityGroupsContentAudit() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeList,
		MaxItems:     1,
		ExactlyOneOf: []string{"security_service_policy_data.0.waf", "security_service_policy_data.0.wafv2", "security_service_policy_data.0.shield_advanced", "security_service_policy_data.0.security_groups_common", "security_service_policy_data.0.security_groups_content_audit", "security_service_policy_data.0.security_groups_usage_audit"},
		Optional:     true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"revert_manual_security_group_changes": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"security_group_action": {
					Type:     schema.TypeString,
					Required: true,
				},
				"security_groups": {
					Type:     schema.TypeSet,
					Required: true,
					Elem:     schema.TypeString,
				},
				"remediation_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"resource_type": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func securityGroupsUsageAudit() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeList,
		MaxItems:     1,
		ExactlyOneOf: []string{"security_service_policy_data.0.waf", "security_service_policy_data.0.wafv2", "security_service_policy_data.0.shield_advanced", "security_service_policy_data.0.security_groups_common", "security_service_policy_data.0.security_groups_content_audit", "security_service_policy_data.0.security_groups_usage_audit"},
		Optional:     true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"revert_manual_security_group_changes": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"security_group_action": {
					Type:     schema.TypeString,
					Required: true,
				},
				"security_groups": {
					Type:     schema.TypeSet,
					Required: true,
					Elem:     schema.TypeString,
				},
				"remediation_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"resource_type": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func wafSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeList,
		MaxItems:     1,
		ExactlyOneOf: []string{"security_service_policy_data.0.waf", "security_service_policy_data.0.wafv2", "security_service_policy_data.0.shield_advanced", "security_service_policy_data.0.security_groups_common", "security_service_policy_data.0.security_groups_content_audit", "security_service_policy_data.0.security_groups_usage_audit"},
		Optional:     true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"rule_groups": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:     schema.TypeString,
								Required: true,
							},
							"override_action": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"type": {
											Type:     schema.TypeString,
											Required: true,
										},
										"default_action": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func wafV2Schema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeList,
		MaxItems:     1,
		ExactlyOneOf: []string{"security_service_policy_data.0.waf", "security_service_policy_data.0.wafv2", "security_service_policy_data.0.shield_advanced", "security_service_policy_data.0.security_groups_common", "security_service_policy_data.0.security_groups_content_audit", "security_service_policy_data.0.security_groups_usage_audit"},
		Optional:     true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"preprocess_rule_groups": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"rule_group_arn": {
								Type:     schema.TypeString,
								Required: true,
							},
							"override_action": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"managed_rule_group_identifier": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"version": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"vendor_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"managed_rule_group_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"rule_group_type": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"exclude_rule_groups": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     schema.TypeString,
				},
				"post_process_rule_groups": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     schema.TypeString,
				},
				"default_action": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"override_customer_web_acl_association": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"logging_configuration": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"log_destination_configs": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     schema.TypeString,
							},
							"redacted_fields": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"redacted_field_type": {
											Type:     schema.TypeString,
											Required: true,
										},
										"redacted_field_value": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Shared structs
type fmsPolicyBasicType struct {
	Type string `json:"type" mapstructure:"type"`
}

// WAF structs
type fmsPolicyManagedServiceDataWAF struct {
	Type       string               `json:"type" mapstructure:"type"`
	RuleGroups []fmsPolicyRuleGroup `json:"ruleGroups" mapstructure:"rule_groups"`
}

type fmsPolicyRuleGroup struct {
	ID              string             `json:"id" mapstructure:"id"`
	OverrrideAction fmsPolicyBasicType `json:"overrideAction" mapstructure:"override_action"`
}

// WAFv2 structs
type fmsPolicyManagedServiceDataWAFV2 struct {
	Type                              string                        `json:"type" mapstructure:"type"`
	PreProcessRuleGroups              []fmsPolicyProcessRuleGroup   `json:"preProcessRuleGroups" mapstructure:"preprocess_rule_groups"`
	PostProcessRuleGroups             []fmsPolicyProcessRuleGroup   `json:"postProcessRuleGroups" mapstructure:"postprocess_rule_groups"`
	DefaultAction                     fmsPolicyBasicType            `json:"defaultAction" mapstructure:"default_action"`
	OverrideCustomerWebACLAssociation bool                          `json:"overrideCustomerWebACLAssociation" mapstructure:"override_customer_web_acl_association"`
	LoggingConfiguration              fmsPolicyLoggingConfiguration `json:"loggingConfiguration" mapstructure:"logging_configuration"`
}

type fmsPolicyLoggingConfiguration struct {
	LogDestinationConfigs []string                 `json:"logDestinationConfigs" mapstructure:"log_destination_configs"`
	RedactedFields        []fmsPolicyRedactedField `json:"redactedFields" mapstructure:"redacted_fields"`
}

type fmsPolicyRedactedField struct {
	RedactedFieldType  string `json:"redactedFieldType" mapstructure:"redacted_field_type"`
	RedactedFieldValue string `json:"redactedFieldValue" mapstructure:"redacted_field_value"`
}

type fmsPolicyRuleGroupIdentifier struct {
	Version              string `json:"version" mapstructure:"version"`
	VendorName           string `json:"vendorName" mapstructure:"vendor_name"`
	ManagedRuleGroupName string `json:"managedRuleGroupName" mapstructure:"managed_rule_group_name"`
}

type fmsPolicyProcessRuleGroup struct {
	RuleGroupARN               string                       `json:"ruleGroupArn" mapstructure:"rule_group_arn"`
	OverrideAction             fmsPolicyBasicType           `json:"overrideAction" mapstructure:"override_action"`
	ManagedRuleGroupIdentifier fmsPolicyRuleGroupIdentifier `json:"managedRuleGroupIdentifier" mapstructure:"managed_rule_group_identifier"`
	RuleGroupType              string                       `json:"ruleGroupType" mapstructure:"rule_group_type"`
	ExcludeRules               []string                     `json:"excludeRules" mapstructure:"excluded_rules"`
}

// SECURITY_GROUPS_COMMON structs
type fmsPolicyManagedServiceDataSecurityGroupsCommon struct {
	Type                                     string               `json:"type" mapstructure:"type"`
	RevertManualSecurityGroupChanges         bool                 `json:"revertManualSecurityGroupChanges" mapstructure:"revert_manual_security_group_changes"`
	ExclusiveResourceSecurityGroupManagement bool                 `json:"exclusiveResourceSecurityGroupManagement" mapstructure:"exclusive_resource_security_group_management"`
	SecurityGroups                           []fmsPolicyRuleGroup `json:"ruleGroups" mapstructure:"security_groups"`
}

// SECURITY_GROUPS_CONTENT_AUDIT structs
type fmsPolicyManagedServiceDataSecurityGroupsContentAudit struct {
	Type                string               `json:"type" mapstructure:"type"`
	SecurityGroups      []fmsPolicyRuleGroup `json:"ruleGroups" mapstructure:"security_groups"`
	SecurityGroupAction fmsPolicyBasicType   `json:"securityGroupAction" mapstructure:"security_group_action"`
}

// SECURITY_GROUPS_USAGE_AUDIT structs
type fmsPolicyManagedServiceDataSecurityGroupsUsageAudit struct {
	Type                            string `json:"type" mapstructure:"type"`
	DeleteUnusedSecurityGroups      bool   `json:"deleteUnusedSecurityGroups" mapstructure:"delete_unused_security_groups"`
	CoalesceRedundantSecurityGroups bool   `json:"coalesceRedundantSecurityGroups" mapstructure:"coalesce_redundant_security_groups"`
}

func resourceAwsFmsPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	fmsPolicy := &fms.Policy{
		PolicyName:          aws.String(d.Get("name").(string)),
		RemediationEnabled:  aws.Bool(d.Get("remediation_enabled").(bool)),
		ResourceType:        aws.String("ResourceTypeList"),
		ResourceTypeList:    expandStringList(d.Get("resource_type_list").(*schema.Set).List()),
		ExcludeResourceTags: aws.Bool(d.Get("exclude_resource_tags").(bool)),
	}

	if v, ok := d.GetOk("security_service_policy_data"); ok {
		var err error
		if fmsPolicy.SecurityServicePolicyData, err = expandSecurityServicePolicyData(v.([]interface{})[0].(map[string]interface{})); err != nil {
			return err
		}
	}

	if rTags, tagsOk := d.GetOk("resource_tags"); tagsOk {
		fmsPolicy.ResourceTags = constructResourceTags(rTags)
	}

	if v, ok := d.GetOk("include_map"); ok {
		fmsPolicy.IncludeMap = expandFMSPolicyMap(v.(*schema.Set))
	}

	if v, ok := d.GetOk("exclude_map"); ok {
		fmsPolicy.ExcludeMap = expandFMSPolicyMap(v.(*schema.Set))
	}

	params := &fms.PutPolicyInput{
		Policy: fmsPolicy,
	}

	var resp *fms.PutPolicyOutput
	var err error

	resp, err = conn.PutPolicy(params)

	if err != nil {
		return fmt.Errorf("Creating Policy Failed: %s", err.Error())
	}

	d.SetId(aws.StringValue(resp.Policy.PolicyId))

	return resourceAwsFmsPolicyRead(d, meta)
}

func resourceAwsFmsPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	var resp *fms.GetPolicyOutput
	var req = &fms.GetPolicyInput{
		PolicyId: aws.String(d.Id()),
	}

	resp, err := conn.GetPolicy(req)

	if err != nil {
		if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] FMS Policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("arn", aws.StringValue(resp.PolicyArn))

	d.Set("name", aws.StringValue(resp.Policy.PolicyName))
	d.Set("exclude_resource_tags", aws.BoolValue(resp.Policy.ExcludeResourceTags))
	if err = d.Set("exclude_map", flattenFMSPolicyMap(resp.Policy.ExcludeMap)); err != nil {
		return err
	}
	d.Set("include_map", flattenFMSPolicyMap(resp.Policy.IncludeMap))
	d.Set("remediation_enabled", aws.BoolValue(resp.Policy.RemediationEnabled))
	d.Set("resource_type_list", resp.Policy.ResourceTypeList)
	d.Set("policy_update_token", aws.StringValue(resp.Policy.PolicyUpdateToken))
	d.Set("resource_tags", flattenFMSResourceTags(resp.Policy.ResourceTags))

	securityServicePolicyData, err := fmsPolicyUnmarshalManagedServiceData(resp.Policy.SecurityServicePolicyData)
	if err != nil {
		return err
	}
	if err = d.Set("security_service_policy_data", securityServicePolicyData); err != nil {
		return err
	}

	return nil
}

func resourceAwsFmsPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	fmsPolicy := &fms.Policy{
		PolicyName:          aws.String(d.Get("name").(string)),
		PolicyId:            aws.String(d.Id()),
		PolicyUpdateToken:   aws.String(d.Get("policy_update_token").(string)),
		RemediationEnabled:  aws.Bool(d.Get("remediation_enabled").(bool)),
		ResourceType:        aws.String("ResourceTypeList"),
		ResourceTypeList:    expandStringList(d.Get("resource_type_list").(*schema.Set).List()),
		ExcludeResourceTags: aws.Bool(d.Get("exclude_resource_tags").(bool)),
	}

	requestUpdate := false

	if d.HasChange("exclude_map") {
		fmsPolicy.ExcludeMap = expandFMSPolicyMap(d.Get("exclude_map").(*schema.Set))
		requestUpdate = true
	}

	if d.HasChange("include_map") {
		fmsPolicy.IncludeMap = expandFMSPolicyMap(d.Get("include_map").(*schema.Set))
		requestUpdate = true
	}

	if d.HasChange("resource_tags") {
		fmsPolicy.ResourceTags = constructResourceTags(d.Get("resource_tags"))
		requestUpdate = true
	}

	if requestUpdate {
		var err error
		if fmsPolicy.SecurityServicePolicyData, err = expandSecurityServicePolicyData(d.Get("security_service_policy_data").(*schema.Set).List()[0].(map[string]interface{})); err != nil {
			return err
		}

		params := &fms.PutPolicyInput{Policy: fmsPolicy}
		_, err = conn.PutPolicy(params)

		if err != nil {
			return fmt.Errorf("Error modifying FMS Policy Rule: %s", err)
		}
	}

	return resourceAwsFmsPolicyRead(d, meta)
}

func resourceAwsFmsPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn
	log.Printf("[DEBUG] Delete FMS Policy: %s", d.Id())

	_, err := conn.DeletePolicy(&fms.DeletePolicyInput{
		PolicyId:                 aws.String(d.Id()),
		DeleteAllPolicyResources: aws.Bool(d.Get("delete_all_policy_resources").(bool)),
	})

	if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FMS Policy (%s): %s", d.Id(), err)
	}

	return nil
}

func terraformMapDecodeHelper() mapstructure.DecodeHookFuncType {
	return func(inType reflect.Type, outType reflect.Type, value interface{}) (interface{}, error) {
		if inType == reflect.SliceOf(outType) && reflect.ValueOf(value).Len() == 1 {
			return reflect.ValueOf(value).Index(0).Interface(), nil
		}
		return value, nil
	}
}

func fmsPolicyUnmarshalManagedServiceData(policyData *fms.SecurityServicePolicyData) (map[string]interface{}, error) {
	var policyStruct interface{}
	var securityServicePolicy map[string]interface{}
	var policyType string
	switch *policyData.Type {
	case "WAF":
		policyType = "waf"
		policyStruct = fmsPolicyManagedServiceDataWAF{}
		if err := json.Unmarshal([]byte(*policyData.ManagedServiceData), &policyStruct); err != nil {
			return nil, err
		}
	case "WAFV2":
		policyType = "wafv2"
		policyStruct = fmsPolicyManagedServiceDataWAFV2{}
		if err := json.Unmarshal([]byte(*policyData.ManagedServiceData), &policyStruct); err != nil {
			return nil, err
		}
	case "SECURITY_GROUPS_COMMON":
		policyType = "security_groups_common"
		policyStruct = fmsPolicyManagedServiceDataSecurityGroupsCommon{}
		if err := json.Unmarshal([]byte(*policyData.ManagedServiceData), &policyStruct); err != nil {
			return nil, err
		}
	case "SECURITY_CONTENT_AUDIT":
		policyType = "security_groups_content_audit"
		policyStruct = fmsPolicyManagedServiceDataSecurityGroupsContentAudit{}
		if err := json.Unmarshal([]byte(*policyData.ManagedServiceData), &policyStruct); err != nil {
			return nil, err
		}
	case "SECURITY_GROUPS_USAGE_AUDIT":
		policyType = "security_groups_usage_audit"
		policyStruct = fmsPolicyManagedServiceDataSecurityGroupsUsageAudit{}
		if err := json.Unmarshal([]byte(*policyData.ManagedServiceData), &policyStruct); err != nil {
			return nil, err
		}
	case "SHIELD_ADVANCED":
		policyType = "security_groups_usage_audit"
		policyStruct = true
	}
	var policyMap map[string]interface{}
	err := mapstructure.Decode(policyStruct, policyMap)
	securityServicePolicy[policyType] = policyMap
	return securityServicePolicy, err
}

func fmsPolicyMarshalManagedServiceData(policyMap interface{}, policyStruct interface{}) (*string, error) {
	var managedServiceData []byte
	var err error
	decoderConfig := mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &policyStruct,
		DecodeHook:       terraformMapDecodeHelper(),
	}
	weakDecoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		return nil, err
	}
	if err = weakDecoder.Decode(policyMap); err != nil {
		return nil, err
	}
	if managedServiceData, err = json.Marshal(policyStruct); err != nil {
		return nil, err
	}
	return aws.String(string(managedServiceData)), nil
}

func expandSecurityServicePolicyData(policyData map[string]interface{}) (*fms.SecurityServicePolicyData, error) {
	var managedServiceData *string
	var err error
	var SecurityPolicyType string
	switch {
	case len(policyData["waf"].([]interface{})) == 1:
		policyStruct := fmsPolicyManagedServiceDataWAF{}
		SecurityPolicyType = "WAF"
		managedServiceData, err = fmsPolicyMarshalManagedServiceData(policyData["waf"].([]interface{})[0], policyStruct)

	case len(policyData["wafv2"].([]interface{})) == 1:
		policyStruct := fmsPolicyManagedServiceDataWAFV2{}
		SecurityPolicyType = "WAFV2"
		managedServiceData, err = fmsPolicyMarshalManagedServiceData(policyData["wafv2"].([]interface{})[0], policyStruct)

	case policyData["shield_advanced"].(bool):
		SecurityPolicyType = "SHIELD_ADVANCED"
		managedServiceData = aws.String("{}")

	case len(policyData["security_groups_common"].([]interface{})) == 1:
		policyStruct := fmsPolicyManagedServiceDataSecurityGroupsCommon{}
		SecurityPolicyType = "SECURITY_GROUPS_COMMON"
		managedServiceData, err = fmsPolicyMarshalManagedServiceData(policyData["security_groups_common"].([]interface{})[0], policyStruct)

	case len(policyData["security_groups_content_audit"].([]interface{})) == 1:
		policyStruct := fmsPolicyManagedServiceDataSecurityGroupsContentAudit{}
		SecurityPolicyType = "SECURITY_GROUPS_CONTENT_AUDIT"
		managedServiceData, err = fmsPolicyMarshalManagedServiceData(policyData["security_groups_content_audit"].([]interface{})[0], policyStruct)

	case len(policyData["security_groups_usage_audit"].([]interface{})) == 1:
		policyStruct := fmsPolicyManagedServiceDataSecurityGroupsUsageAudit{}
		SecurityPolicyType = "SECURITY_GROUPS_USAGE_AUDIT"
		managedServiceData, err = fmsPolicyMarshalManagedServiceData(policyData["security_groups_usage_audit"].([]interface{})[0], policyStruct)
	}

	return &fms.SecurityServicePolicyData{
		Type:               aws.String(SecurityPolicyType),
		ManagedServiceData: managedServiceData,
	}, err
}

func expandFMSPolicyMap(set *schema.Set) map[string][]*string {
	fmsPolicyMap := map[string][]*string{}
	if set.Len() > 0 {
		for key, listValue := range set.List()[0].(map[string]interface{}) {
			for _, value := range listValue.([]interface{}) {
				fmsPolicyMap[key] = append(fmsPolicyMap[key], aws.String(value.(string)))
			}
		}
	}
	return fmsPolicyMap
}

func flattenFMSPolicyMap(fmsPolicyMap map[string][]*string) []interface{} {
	flatPolicyMap := map[string]interface{}{}

	for key, value := range fmsPolicyMap {
		switch key {
		case "account":
			flatPolicyMap["account"] = value
		case "orgunit":
			flatPolicyMap["orgunit"] = value
		default:
			log.Printf("[WARNING] Unexpected key (%q) found in FMS policy", key)
		}
	}

	return []interface{}{flatPolicyMap}
}

func flattenFMSResourceTags(resourceTags []*fms.ResourceTag) map[string]interface{} {
	resTags := map[string]interface{}{}

	for _, v := range resourceTags {
		resTags[*v.Key] = v.Value
	}
	return resTags
}

func constructResourceTags(rTags interface{}) []*fms.ResourceTag {
	var rTagList []*fms.ResourceTag

	tags := rTags.(map[string]interface{})
	for k, v := range tags {
		rTagList = append(rTagList, &fms.ResourceTag{Key: aws.String(k), Value: aws.String(v.(string))})
	}

	return rTagList
}
