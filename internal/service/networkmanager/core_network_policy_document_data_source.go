package networkmanager

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

func DataSourceCoreNetworkPolicyDocument() *schema.Resource {
	setOfString := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return &schema.Resource{
		Read: dataSourceCoreNetworkPolicyDocumentRead,
		Schema: map[string]*schema.Schema{
			"attachment_policies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"rule_number": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"condition_logic": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"and",
								"or",
							}, false),
						},
						"conditions": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{

									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"account-id",
											"any",
											"tag-value",
											"tag-exists",
											"resource-id",
											"region",
											"attachment-type",
										}, false),
									},
									"operator": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											"equals",
											"not-equals",
											"contains",
											"begins-with",
										}, false),
									},
									"key": {
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
						"action": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"association_method": {
										Type:     schema.TypeString,
										Required: true,
										// AtLeastOneOf: []string{
										// 	"attachment_policies.#.actions.0.segment",
										// 	"attachment_policies.#.actions.0.tag_value_of_key",
										// },
										ValidateFunc: validation.StringInSlice([]string{
											"tag",
											"constant",
										}, false),
									},
									"segment": {
										Type:     schema.TypeString,
										Optional: true,
										// ConflictsWith: "tag_value_of_key",
										//"^[a-zA-Z][A-Za-z0-9]{0,63}$"
									},
									"tag_value_of_key": {
										Type:     schema.TypeString,
										Optional: true,
										// ConflictsWith: "segment",
									},
									"require_acceptance": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
										//"^[a-zA-Z][A-Za-z0-9]{0,63}$"
									},
								},
							},
						},
					},
				},
			},
			"core_network_configuration": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// TODO: required
						"asn_ranges": setOfString,
						"vpn_ecmp_support": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"inside_cidr_blocks": setOfString,
						"edge_locations": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 17,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"location": {
										Type:     schema.TypeString,
										Required: true,
										// a-z, 0-9
									},
									"asn": {
										Type:     schema.TypeInt,
										Default:  false,
										Optional: true,
										// validate asn-like
									},
									// TODO: recheck type?
									"inside_cidr_blocks": {
										Type:     schema.TypeList,
										Optional: true,
										// validate either ipv4 or 6?
										Elem: &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"segments": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_filter": setOfString,
						"deny_filter":  setOfString,
						"name": {
							Type:     schema.TypeString,
							Required: true,
							// a-z, 0-9
							// ValidateFunc: validation.StringInSlice([]string{"Allow", "Deny"}, false),
						},
						"edge_locations": setOfString,
						"isolate_attachments": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"require_attachment_acceptance": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
					},
				},
			},
			"segment_actions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"share",
								"create-route",
							}, false),
						},
						/* 2 types of segment actions
						action = share: mode = "attachment-route", share-with (list), segment (source of what is sharing) 1 to many segment to many segments

						action = create-route: destination_cidr_blocks, destination (array), segment (no mode, )

						*/
						"destination": setOfString,
						// can be either a list of attachments or ["blackhole"]

						"destination_cidr_blocks": setOfString,
						// list of cidrs ipv4 or ipv6 or a mixture of 4/

						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							//"^attachment\\-route$"
							ValidateFunc: validation.StringInSlice([]string{
								"attachment-route",
							}, false),
						},
						"segment": {
							Type:     schema.TypeString,
							Required: true,
							//"^[a-zA-Z][A-Za-z0-9]{0,63}$"
						},
						/*
							can be array or string or object
							share_with = ["segment-ids", "..."] # subset of all segments
							share_with = "*"                    # all segments
							share_with = {                      # setsubtraction of all segments
								except = ["segment-ids", "..."]
							}
						*/
						"share_with": {
							Type:     schema.TypeList,
							Required: true,
							//"^[a-zA-Z][A-Za-z0-9]{0,63}$"
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2021.12",
				ValidateFunc: validation.StringInSlice([]string{
					"2021.12",
				}, false),
			},
		},
	}
}

func dataSourceCoreNetworkPolicyDocumentRead(d *schema.ResourceData, meta interface{}) error {
	mergedDoc := &CoreNetworkPolicyDoc{}

	doc := &CoreNetworkPolicyDoc{
		Version: d.Get("version").(string),
	}

	// CoreNetworkConfiguration
	networkConfiguration, err := expandDataCoreNetworkPolicyNetworkConfiguration(d)
	if err != nil {
		return err
	}
	mergedDoc.CoreNetworkConfiguration = networkConfiguration

	// AttachmentPolicies
	attachmentPolicies, err := expandDataCoreNetworkPolicyAttachmentPolicies(d)
	if err != nil {
		return err
	}
	mergedDoc.AttachmentPolicies = attachmentPolicies

	// Segments
	segments, err := expandDataCoreNetworkPolicySegments(d)
	if err != nil {
		return err
	}
	doc.Segments = segments

	mergedDoc.Merge(doc)
	jsonDoc, err := json.MarshalIndent(mergedDoc, "", "  ")
	if err != nil {
		// should never happen if the above code is correct
		return err
	}
	jsonString := string(jsonDoc)

	d.Set("json", jsonString)
	d.SetId(strconv.Itoa(create.StringHashcode(jsonString)))

	return nil
}

func expandDataCoreNetworkPolicyAttachmentPolicies(d *schema.ResourceData) ([]*CoreNetworkAttachmentPolicy, error) {
	var cfgAttachmentPolicyInterface = d.Get("attachment_policies").([]interface{})
	aPolicies := make([]*CoreNetworkAttachmentPolicy, len(cfgAttachmentPolicyInterface))
	ruleMap := make(map[string]struct{})

	for i, polI := range cfgAttachmentPolicyInterface {
		cfgPol := polI.(map[string]interface{})
		policy := &CoreNetworkAttachmentPolicy{}

		rule := cfgPol["rule_number"]
		ruleStr := strconv.Itoa(rule.(int))
		if _, ok := ruleMap[ruleStr]; ok {
			return nil, fmt.Errorf("duplicate Rule Number (%s). Remove the Rule Number or ensure the Rule Number is unique.", ruleStr)
		}
		policy.RuleNumber = rule.(int)
		ruleMap[ruleStr] = struct{}{}

		if desc, ok := cfgPol["description"]; ok {
			policy.Description = desc.(string)
		}
		if cL, ok := cfgPol["condition_logic"]; ok {
			policy.ConditionLogic = cL.(string)
		}

		action := expandDataCoreNetworkPolicyAttachmentPoliciesAction(cfgPol["action"].([]interface{}))
		policy.Action = action

		conditions, err := expandDataCoreNetworkPolicyAttachmentPoliciesConditions(cfgPol["conditions"].([]interface{}))
		if err != nil {
			return nil, err
		}
		policy.Conditions = conditions

		aPolicies[i] = policy
	}

	// adjust
	return aPolicies, nil

}

func expandDataCoreNetworkPolicyAttachmentPoliciesConditions(tfList []interface{}) ([]*CoreNetworkAttachmentPolicyCondition, error) {
	/* 5 situations by type
	any:        no other fields allowed (op, key, value)
	tag-exists: only key allowed. no operator, value
	tag-value:  all fields required

	{region,resource-id,account-id}: operator, value required. no key
	attachment-type:                 operator must be "equals". value is required. key is not allowed
	*/

	conditions := make([]*CoreNetworkAttachmentPolicyCondition, len(tfList))

	for i, condI := range tfList {
		cfgCond := condI.(map[string]interface{})
		condition := &CoreNetworkAttachmentPolicyCondition{}
		k := map[string]bool{
			"operator": false,
			"key":      false,
			"value":    false,
		}

		t := cfgCond["type"].(string)
		condition.Type = t

		if o, _ := cfgCond["operator"]; o != "" {
			k["operator"] = true
			condition.Operator = o.(string)
		}
		if key, _ := cfgCond["key"]; key != "" {
			k["key"] = true
			condition.Key = key.(string)
		}
		if v, _ := cfgCond["value"]; v != "" {
			k["value"] = true
			condition.Value = v.(string)
		}

		if t == "any" {
			for _, key := range k {
				if key {
					return nil, fmt.Errorf("You cannot set \"operator\", \"key\", or \"value\" if type = \"any\".")
				}
			}
		}
		if t == "tag-exists" {
			if k["key"] == false || k["operator"] || k["value"] {
				return nil, fmt.Errorf("You must set \"key\" and cannot set \"operator\", or \"value\" if type = \"tag-exists\".")
			}
		}
		if t == "tag-value" {
			if !k["key"] || !k["operator"] || !k["value"] {
				return nil, fmt.Errorf("You must set \"key\", \"operator\", and \"value\" if type = \"tag-value\".")
			}
		}
		if t == "region" || t == "resource-id" || t == "account-id" {
			if !k["key"] || k["operator"] || k["value"] {
				return nil, fmt.Errorf("You must set \"key\" and \"operator\" and cannot set \"value\" if type = \"region\", \"resource-id\", or \"account-id\".")
			}
		}
		if t == "attachment-type" {
			if k["key"] || !k["value"] || cfgCond["operator"].(string) != "equals" {
				return nil, fmt.Errorf("You must set \"value\", cannot set \"key\" and \"operator\" must be \"equals\" if type = \"attachment-type\".")
			}
		}
		conditions[i] = condition
	}
	return conditions, nil
}

func expandDataCoreNetworkPolicyAttachmentPoliciesAction(tfList []interface{}) *CoreNetworkAttachmentPolicyAction {
	// TODO:
	/*
		if association_method = "tag", must also specify tag_value_of_key
		if association_method = "constant", can be either segment
	*/

	cfgAP := tfList[0].(map[string]interface{})
	aP := &CoreNetworkAttachmentPolicyAction{
		AssociationMethod: cfgAP["association_method"].(string),
	}

	if segment, ok := cfgAP["segment"]; ok {
		aP.Segment = segment.(string)
	}
	if tag, ok := cfgAP["tag_value_of_key"]; ok {
		aP.TagValueOfKey = tag.(string)
	}
	if acceptance, ok := cfgAP["require_acceptance"]; ok {
		aP.RequireAcceptance = acceptance.(bool)
	}
	return aP
}

func expandDataCoreNetworkPolicySegments(d *schema.ResourceData) ([]*CoreNetworkPolicySegment, error) {
	var cfgSgmtIntf = d.Get("segments").([]interface{})
	Sgmts := make([]*CoreNetworkPolicySegment, len(cfgSgmtIntf))
	nameMap := make(map[string]struct{})

	for i, sgmtI := range cfgSgmtIntf {
		cfgSgmt := sgmtI.(map[string]interface{})
		sgmt := &CoreNetworkPolicySegment{}

		if name, ok := cfgSgmt["name"]; ok {
			if _, ok := nameMap[name.(string)]; ok {
				return nil, fmt.Errorf("duplicate Name (%s). Remove the Name or ensure the Name is unique.", name.(string))
			}
			sgmt.Name = name.(string)
			if len(sgmt.Name) > 0 {
				nameMap[sgmt.Name] = struct{}{}
			}
		}
		if actions := cfgSgmt["allow_filter"].(*schema.Set).List(); len(actions) > 0 {
			sgmt.AllowFilter = CoreNetworkPolicyDecodeConfigStringList(actions)
		}
		if actions := cfgSgmt["deny_filter"].(*schema.Set).List(); len(actions) > 0 {
			sgmt.DenyFilter = CoreNetworkPolicyDecodeConfigStringList(actions)
		}
		if b, ok := cfgSgmt["require_attachment_acceptance"]; ok {
			sgmt.RequireAttachmentAcceptance = b.(bool)
		}
		if b, ok := cfgSgmt["isolate_attachments"]; ok {
			sgmt.IsolateAttachments = b.(bool)
		}
		Sgmts[i] = sgmt
	}

	return Sgmts, nil
}

func expandDataCoreNetworkPolicyNetworkConfiguration(d *schema.ResourceData) (*CoreNetworkPolicyCoreNetworkConfiguration, error) {
	var networkCfgIntf = d.Get("core_network_configuration").([]interface{})
	m := networkCfgIntf[0].(map[string]interface{})

	nc := &CoreNetworkPolicyCoreNetworkConfiguration{}

	nc.AsnRanges = CoreNetworkPolicyDecodeConfigStringList(m["asn_ranges"].(*schema.Set).List())

	if cidrs := m["inside_cidr_blocks"].(*schema.Set).List(); len(cidrs) > 0 {
		nc.InsideCidrBlocks = CoreNetworkPolicyDecodeConfigStringList(cidrs)
	}

	if vpn, ok := m["vpn_ecmp_support"]; ok {
		nc.VpnEcmpSupport = vpn.(bool)
	}

	el, err := expandDataCoreNetworkPolicyNetworkConfigurationEdgeLocations(m["edge_locations"].([]interface{}))

	if err != nil {
		return nil, err
	}
	nc.EdgeLocations = el

	return nc, nil

}

func expandDataCoreNetworkPolicyNetworkConfigurationEdgeLocations(tfList []interface{}) ([]*CoreNetworkEdgeLocation, error) {
	edgeLocations := make([]*CoreNetworkEdgeLocation, len(tfList))
	locMap := make(map[string]struct{})

	for i, edgeLocationsRaw := range tfList {

		cfgEdgeLocation, ok := edgeLocationsRaw.(map[string]interface{})
		edgeLocation := &CoreNetworkEdgeLocation{}

		if !ok {
			continue
		}

		location := cfgEdgeLocation["location"].(string)

		if _, ok := locMap[location]; ok {
			return nil, fmt.Errorf("duplicate Location (%s). Remove the Location or ensure the Location is unique.", location)
		}
		edgeLocation.Location = location
		if len(edgeLocation.Location) > 0 {
			locMap[edgeLocation.Location] = struct{}{}
		}

		if v, ok := cfgEdgeLocation["asn"]; ok {
			edgeLocation.Asn = v.(int)
		}

		edgeLocations[i] = edgeLocation
	}
	return edgeLocations, nil
}
