package aws

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
	"strings"
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
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(12, 12),
							},
						},
					},
				},
			},

			"include_map": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(12, 12),
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
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"managed_service_data": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Required: true,
										Type:     schema.TypeString,
									},
									"rule_groups": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"override_action": {
													Type:     schema.TypeMap,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"type": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice([]string{"COUNT", "NONE"}, false),
															},
														},
													},
												},
											},
										},
									},
									"default_action": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Required: true,
													Type:     schema.TypeString,
												},
											},
										},
									},
								},
							},
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"WAF", "ADVANCED_SHIELD"}, false),
						},
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
		fmsPolicy.SecurityServicePolicyData = expandAwsFmsManagedSecurityData(v.(*schema.Set))
	}

	if rTags, tagsOk := d.GetOk("resource_tags"); tagsOk {
		fmsPolicy.ResourceTags = constructResourceTags(rTags)
	}

	if v, ok := d.GetOk("include_map"); ok {
		fmsPolicy.IncludeMap = expandAccountList(v.(*schema.Set))
	}

	if v, ok := d.GetOk("exclude_map"); ok {
		fmsPolicy.ExcludeMap = expandAccountList(v.(*schema.Set))
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
	d.Set("exclude_map", flattenFMSAccountMap(resp.Policy.ExcludeMap))
	d.Set("include_map", flattenFMSAccountMap(resp.Policy.IncludeMap))
	d.Set("remediation_enabled", aws.BoolValue(resp.Policy.RemediationEnabled))
	d.Set("resource_type_list", resp.Policy.ResourceTypeList)
	d.Set("policy_update_token", aws.StringValue(resp.Policy.PolicyUpdateToken))
	d.Set("resource_tags", flattenFMSResourceTags(resp.Policy.ResourceTags))
	d.Set("security_service_policy_data", flattenFmsSecurityServicePolicyData(resp.Policy.SecurityServicePolicyData))

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
		fmsPolicy.ExcludeMap = expandAccountList(d.Get("exclude_map").(*schema.Set))
		requestUpdate = true
	}

	if d.HasChange("include_map") {
		fmsPolicy.ExcludeMap = expandAccountList(d.Get("include_map").(*schema.Set))
		requestUpdate = true
	}

	if d.HasChange("resource_tags") {
		fmsPolicy.ResourceTags = constructResourceTags(d.Get("resource_tags"))
		requestUpdate = true
	}

	if requestUpdate {
		fmsPolicy.SecurityServicePolicyData = expandAwsFmsManagedSecurityData(d.Get("security_service_policy_data").(*schema.Set))

		params := &fms.PutPolicyInput{Policy: fmsPolicy}
		_, err := conn.PutPolicy(params)

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

func expandAccountList(set *schema.Set) map[string][]*string {
	var accountList = make(map[string][]*string)

	for _, account := range set.List() {
		l := account.(map[string]interface{})
		y := l["account"].([]interface{})

		for _, a := range y {
			accountList["ACCOUNT"] = append(accountList["ACCOUNT"], aws.String(a.(string)))
		}
	}

	return accountList
}

func constructManagedServiceData(m []interface{}) map[string]interface{} {
	var msd map[string]interface{}

	for _, data := range m {
		m := data.(map[string]interface{})

		rgl := m["rule_groups"].(*schema.Set).List()
		rgs := constructRuleGroupsList(rgl)

		msd = map[string]interface{}{
			"type":          m["type"].(string),
			"defaultAction": m["default_action"].(map[string]interface{}),
			"ruleGroups":    rgs,
		}
	}
	return msd
}

func constructRuleGroupsList(rgs []interface{}) []map[string]interface{} {
	ruleGroup := []map[string]interface{}{}

	for _, rg := range rgs {
		log.Printf("[DEBUG] Rule_Group Keys: %s", rg)

		m := rg.(map[string]interface{})

		ruleId := m["id"].(string)
		overrideAction := m["override_action"].(map[string]interface{})

		rule := map[string]interface{}{
			"id":             ruleId,
			"overrideAction": overrideAction,
		}

		ruleGroup = append(ruleGroup, rule)
	}
	return ruleGroup
}

func expandAwsFmsManagedSecurityData(set *schema.Set) *fms.SecurityServicePolicyData {
	spd := set.List()

	securityServicePolicyData := &fms.SecurityServicePolicyData{}

	for _, t := range spd {
		spdMap := t.(map[string]interface{})
		spdType := spdMap["type"].(string)

		securityServicePolicyData.Type = aws.String(spdType)

		switch spdType {
		case "WAF":
			if v, ok := spdMap["managed_service_data"]; !ok {
				log.Printf("[DEBUG] Error Looking up Managed Service Data: %s", v)
			} else {
				spdPolicy := constructManagedServiceData(v.(*schema.Set).List())

				js, err := json.Marshal(spdPolicy)
				if err != nil {
					log.Printf("[DEBUG] JSON Error: %s", err)
				}

				securityServicePolicyData.ManagedServiceData = aws.String(string(js))
			}
		}
	}

	return securityServicePolicyData
}

func flattenFMSAccountMap(accountMap map[string][]*string) *schema.Set {
	eMap := map[string]interface{}{}

	if _, ok := eMap["account"]; ok {
		for _, v := range accountMap["ACCOUNT"] {
			eMap["account"] = append(eMap["account"].([]*string), v)
		}
	}

	s := schema.NewSet(fmsPolicyDataHash, []interface{}{})
	s.Add(eMap)

	return s
}

func flattenFMSResourceTags(resourceTags []*fms.ResourceTag) map[string]interface{} {
	resTags := map[string]interface{}{}

	for _, v := range resourceTags {
		resTags[*v.Key] = v.Value
	}
	return resTags
}

func flattenFmsManagedServiceData(sspdMsd map[string]interface{}) *schema.Set {
	msdSS := schema.NewSet(fmsPolicyDataHash, []interface{}{})

	msdData := map[string]interface{}{
		"type": sspdMsd["type"].(string),
	}

	if sspdMsd["defaultAction"] != nil {
		msdData["default_action"] = sspdMsd["defaultAction"]
	}

	msdData["rule_groups"] = flattenFmsMsdRuleGroupsList(sspdMsd)

	msdSS.Add(msdData)

	return msdSS
}

func flattenFmsMsdRuleGroupsList(sspdMsd map[string]interface{}) *schema.Set {
	ruleGroupsSet := schema.NewSet(fmsPolicyDataHash, []interface{}{})
	if sspdMsd["ruleGroups"] != nil {
		for _, v := range sspdMsd["ruleGroups"].([]interface{}) {

			rg := v.(map[string]interface{})

			rule := map[string]interface{}{
				"id":              rg["id"].(string),
				"override_action": rg["overrideAction"].(map[string]interface{}),
			}

			ruleGroupsSet.Add(rule)
		}
	}
	return ruleGroupsSet
}

func flattenFmsSecurityServicePolicyData(spd *fms.SecurityServicePolicyData) *schema.Set {
	s := schema.NewSet(fmsPolicyDataHash, []interface{}{})

	sspd := map[string]interface{}{
		"type": aws.StringValue(spd.Type),
	}

	var policy map[string]interface{}

	if spd.ManagedServiceData != nil {

		msd := []byte(aws.StringValue(spd.ManagedServiceData))

		if err := json.Unmarshal(msd, &policy); err != nil {
			panic(err)
		}

		sspd["managed_service_data"] = flattenFmsManagedServiceData(policy)
	}

	s.Add(sspd)

	return s
}

func constructResourceTags(rTags interface{}) []*fms.ResourceTag {
	var rTagList []*fms.ResourceTag

	tags := rTags.(map[string]interface{})
	for k, v := range tags {
		rTagList = append(rTagList, &fms.ResourceTag{Key: aws.String(k), Value: aws.String(v.(string))})
	}

	return rTagList
}

func fmsPolicyDataHash(v interface{}) int {
	var buf strings.Builder

	m := v.(map[string]interface{})

	if _, ok := m["Id"]; ok {
		buf.WriteString(fmt.Sprintf("%s", m["Id"]))
	} else {
		buf.WriteString(fmt.Sprintf("%s-", m))
	}

	return hashcode.String(buf.String())
}
