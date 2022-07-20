package networkmanager

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
						"condition_logic": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"and",
								"or",
							}, false),
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"rule_number": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
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
										ValidateFunc: validation.StringInSlice([]string{
											"tag",
											"constant",
										}, false),
									},
									"segment": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][A-Za-z0-9]{0,63}$`),
											"must begin with a letter and contain only alphanumeric characters"),
									},
									"tag_value_of_key": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"require_acceptance": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
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
						"asn_ranges": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"vpn_ecmp_support": {
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
						},
						"inside_cidr_blocks": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
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
										// Not all regions are valid but we will not maintain a hardcoded list
										ValidateFunc: verify.ValidRegionName,
									},
									"asn": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"inside_cidr_blocks": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
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
						"allow_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][A-Za-z0-9]{0,63}$`),
									"must begin with a letter and contain only alphanumeric characters"),
							},
						},
						"deny_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][A-Za-z0-9]{0,63}$`),
									"must begin with a letter and contain only alphanumeric characters"),
							},
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][A-Za-z0-9]{0,63}$`),
								"must begin with a letter and contain only alphanumeric characters"),
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"edge_locations": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidRegionName,
							},
						},
						"isolate_attachments": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"require_attachment_acceptance": {
							Type:     schema.TypeBool,
							Default:  true,
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

						"destinations": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.Any(
									validation.StringInSlice([]string{
										"blackhole",
									}, false),
									validation.StringMatch(regexp.MustCompile(`^attachment-[0-9a-f]{17}$`),
										"must be a list of valid attachment ids or a list with only the word \"blackhole\"."),
								),
							},
						},
						"destination_cidr_blocks": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.Any(
									verify.ValidIPv4CIDRNetworkAddress,
									verify.ValidIPv6CIDRNetworkAddress,
								),
							},
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"attachment-route",
							}, false),
						},
						"segment": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][A-Za-z0-9]{0,63}$`),
								"must begin with a letter and contain only alphanumeric characters"),
						},
						"share_with":        setOfString,
						"share_with_except": setOfString,
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

	mergedDoc := &CoreNetworkPolicyDoc{
		Version: d.Get("version").(string),
	}

	// CoreNetworkConfiguration
	networkConfiguration, err := expandDataCoreNetworkPolicyNetworkConfiguration(d.Get("core_network_configuration").([]interface{}))
	if err != nil {
		return err
	}
	mergedDoc.CoreNetworkConfiguration = networkConfiguration

	// AttachmentPolicies
	attachmentPolicies, err := expandDataCoreNetworkPolicyAttachmentPolicies(d.Get("attachment_policies").([]interface{}))
	if err != nil {
		return err
	}
	mergedDoc.AttachmentPolicies = attachmentPolicies

	// SegmentActions
	segment_actions, err := expandDataCoreNetworkPolicySegmentActions(d.Get("segment_actions").([]interface{}))
	if err != nil {
		return err
	}
	mergedDoc.SegmentActions = segment_actions

	// Segments
	segments, err := expandDataCoreNetworkPolicySegments(d.Get("segments").([]interface{}))
	if err != nil {
		return err
	}
	mergedDoc.Segments = segments

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

func expandDataCoreNetworkPolicySegmentActions(cfgSegmentActionsIntf []interface{}) ([]*CoreNetworkPolicySegmentAction, error) {
	sgmtActions := make([]*CoreNetworkPolicySegmentAction, len(cfgSegmentActionsIntf))
	for i, sgmtActionI := range cfgSegmentActionsIntf {
		cfgSA := sgmtActionI.(map[string]interface{})
		sgmtAction := &CoreNetworkPolicySegmentAction{}
		action := cfgSA["action"].(string)
		sgmtAction.Action = action
		var shareWith, shareWithExcept interface{}

		if action == "share" {
			if mode, ok := cfgSA["mode"]; ok {
				sgmtAction.Mode = mode.(string)
			}

			if sgmt, ok := cfgSA["segment"]; ok {
				sgmtAction.Segment = sgmt.(string)
			}

			if sW := cfgSA["share_with"].(*schema.Set).List(); len(sW) > 0 {
				shareWith = CoreNetworkPolicyDecodeConfigStringList(sW)
				sgmtAction.ShareWith = shareWith
			}

			if sWE := cfgSA["share_with_except"].(*schema.Set).List(); len(sWE) > 0 {
				shareWithExcept = CoreNetworkPolicyDecodeConfigStringList(sWE)
				sgmtAction.ShareWithExcept = shareWithExcept
			}

			if (shareWith != nil && shareWithExcept != nil) || (shareWith == nil && shareWithExcept == nil) {
				return nil, fmt.Errorf("You must specify only 1 of \"share_with\" or \"share_with_except\". See segment_actions[%s].", strconv.Itoa(i))
			}
		}

		if action == "create-route" {
			if mode := cfgSA["mode"]; mode != "" {
				return nil, fmt.Errorf("Cannot specify \"mode\" if action = \"create-route\". See segment_actions[%s].", strconv.Itoa(i))
			}

			if dest := cfgSA["destinations"].(*schema.Set).List(); len(dest) > 0 {
				sgmtAction.Destinations = CoreNetworkPolicyDecodeConfigStringList(dest)
			}

			if destCidrB := cfgSA["destination_cidr_blocks"].(*schema.Set).List(); len(destCidrB) > 0 {
				sgmtAction.DestinationCidrBlocks = CoreNetworkPolicyDecodeConfigStringList(destCidrB)
			}
		}

		if sgmt, ok := cfgSA["segment"]; ok {
			sgmtAction.Segment = sgmt.(string)
		}

		sgmtActions[i] = sgmtAction

	}
	return sgmtActions, nil
}

func expandDataCoreNetworkPolicyAttachmentPolicies(cfgAttachmentPolicyIntf []interface{}) ([]*CoreNetworkAttachmentPolicy, error) {
	aPolicies := make([]*CoreNetworkAttachmentPolicy, len(cfgAttachmentPolicyIntf))
	ruleMap := make(map[string]struct{})

	for i, polI := range cfgAttachmentPolicyIntf {
		cfgPol := polI.(map[string]interface{})
		policy := &CoreNetworkAttachmentPolicy{}

		rule := cfgPol["rule_number"].(int)
		ruleStr := strconv.Itoa(rule)
		if _, ok := ruleMap[ruleStr]; ok {
			return nil, fmt.Errorf("duplicate Rule Number (%s). Remove the Rule Number or ensure the Rule Number is unique.", ruleStr)
		}
		policy.RuleNumber = rule
		ruleMap[ruleStr] = struct{}{}

		if desc, ok := cfgPol["description"]; ok {
			policy.Description = desc.(string)
		}
		if cL, ok := cfgPol["condition_logic"]; ok {
			policy.ConditionLogic = cL.(string)
		}

		action, err := expandDataCoreNetworkPolicyAttachmentPoliciesAction(cfgPol["action"].([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("Problem with attachment policy rule number (%s). See attachment_policy[%s].action: %q", ruleStr, strconv.Itoa(i), err)
		}
		policy.Action = action

		conditions, err := expandDataCoreNetworkPolicyAttachmentPoliciesConditions(cfgPol["conditions"].([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("Problem with attachment policy rule number (%s). See attachment_policy[%s].conditions %q", ruleStr, strconv.Itoa(i), err)
		}
		policy.Conditions = conditions

		aPolicies[i] = policy
	}

	// adjust
	return aPolicies, nil

}

func expandDataCoreNetworkPolicyAttachmentPoliciesConditions(tfList []interface{}) ([]*CoreNetworkAttachmentPolicyCondition, error) {
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

		if o := cfgCond["operator"]; o != "" {
			k["operator"] = true
			condition.Operator = o.(string)
		}
		if key := cfgCond["key"]; key != "" {
			k["key"] = true
			condition.Key = key.(string)
		}
		if v := cfgCond["value"]; v != "" {
			k["value"] = true
			condition.Value = v.(string)
		}

		if t == "any" {
			for _, key := range k {
				if key {
					return nil, fmt.Errorf("Conditions %s: You cannot set \"operator\", \"key\", or \"value\" if type = \"any\".", strconv.Itoa(i))
				}
			}
		}
		if t == "tag-exists" {
			if !k["key"] || k["operator"] || k["value"] {
				return nil, fmt.Errorf("Conditions %s: You must set \"key\" and cannot set \"operator\", or \"value\" if type = \"tag-exists\".", strconv.Itoa(i))
			}
		}
		if t == "tag-value" {
			if !k["key"] || !k["operator"] || !k["value"] {
				return nil, fmt.Errorf("Conditions %s: You must set \"key\", \"operator\", and \"value\" if type = \"tag-value\".", strconv.Itoa(i))
			}
		}
		if t == "region" || t == "resource-id" || t == "account-id" {
			if k["key"] || !k["operator"] || !k["value"] {
				return nil, fmt.Errorf("Conditions %s: You must set \"value\" and \"operator\" and cannot set \"key\" if type = \"region\", \"resource-id\", or \"account-id\".", strconv.Itoa(i))
			}
		}
		if t == "attachment-type" {
			if k["key"] || !k["value"] || cfgCond["operator"].(string) != "equals" {
				return nil, fmt.Errorf("Conditions %s: You must set \"value\", cannot set \"key\" and \"operator\" must be \"equals\" if type = \"attachment-type\".", strconv.Itoa(i))
			}
		}
		conditions[i] = condition
	}
	return conditions, nil
}

func expandDataCoreNetworkPolicyAttachmentPoliciesAction(tfList []interface{}) (*CoreNetworkAttachmentPolicyAction, error) {
	cfgAP := tfList[0].(map[string]interface{})
	assocMethod := cfgAP["association_method"].(string)
	aP := &CoreNetworkAttachmentPolicyAction{
		AssociationMethod: assocMethod,
	}

	if segment := cfgAP["segment"]; segment != "" {
		if assocMethod == "tag" {
			return nil, fmt.Errorf("Cannot set \"segment\" argument if association_method = \"tag\".")
		}
		aP.Segment = segment.(string)
	}
	if tag := cfgAP["tag_value_of_key"]; tag != "" {
		if assocMethod == "constant" {
			return nil, fmt.Errorf("Cannot set \"tag_value_of_key\" argument if association_method = \"constant\".")
		}
		aP.TagValueOfKey = tag.(string)
	}
	if acceptance, ok := cfgAP["require_acceptance"]; ok {
		aP.RequireAcceptance = acceptance.(bool)
	}
	return aP, nil
}

func expandDataCoreNetworkPolicySegments(cfgSgmtIntf []interface{}) ([]*CoreNetworkPolicySegment, error) {
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
		if description, ok := cfgSgmt["description"]; ok {
			sgmt.Description = description.(string)
		}
		if actions := cfgSgmt["allow_filter"].(*schema.Set).List(); len(actions) > 0 {
			sgmt.AllowFilter = CoreNetworkPolicyDecodeConfigStringList(actions)
		}
		if actions := cfgSgmt["deny_filter"].(*schema.Set).List(); len(actions) > 0 {
			sgmt.DenyFilter = CoreNetworkPolicyDecodeConfigStringList(actions)
		}
		if edgeLocations := cfgSgmt["edge_locations"].(*schema.Set).List(); len(edgeLocations) > 0 {
			sgmt.EdgeLocations = CoreNetworkPolicyDecodeConfigStringList(edgeLocations)
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

func expandDataCoreNetworkPolicyNetworkConfiguration(networkCfgIntf []interface{}) (*CoreNetworkPolicyCoreNetworkConfiguration, error) {
	m := networkCfgIntf[0].(map[string]interface{})

	nc := &CoreNetworkPolicyCoreNetworkConfiguration{}

	nc.AsnRanges = CoreNetworkPolicyDecodeConfigStringList(m["asn_ranges"].(*schema.Set).List())

	if cidrs := m["inside_cidr_blocks"].(*schema.Set).List(); len(cidrs) > 0 {
		nc.InsideCidrBlocks = CoreNetworkPolicyDecodeConfigStringList(cidrs)
	}

	nc.VpnEcmpSupport = m["vpn_ecmp_support"].(bool)

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

		if cidrs := cfgEdgeLocation["inside_cidr_blocks"].([]interface{}); len(cidrs) > 0 {
			edgeLocation.InsideCidrBlocks = CoreNetworkPolicyDecodeConfigStringList(cidrs)
		}

		edgeLocations[i] = edgeLocation
	}
	return edgeLocations, nil
}
