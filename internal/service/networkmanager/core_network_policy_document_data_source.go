// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkmanager_core_network_policy_document")
func DataSourceCoreNetworkPolicyDocument() *schema.Resource {
	setOfString := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCoreNetworkPolicyDocumentRead,

		// Order attributes to match model structures and documentation:
		// https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-policies-json.html.
		Schema: map[string]*schema.Schema{
			names.AttrVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2021.12",
				ValidateFunc: validation.StringInSlice([]string{
					"2021.12",
				}, false),
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
						"inside_cidr_blocks": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"vpn_ecmp_support": {
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
						},
						"edge_locations": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrLocation: {
										Type:     schema.TypeString,
										Required: true,
										// Not all regions are valid but we will not maintain a hardcoded list
										ValidateFunc: verify.ValidRegionName,
									},
									"asn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.Valid4ByteASN,
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
			"segments": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[A-Za-z][0-9A-Za-z]{0,63}$`),
								"must begin with a letter and contain only alphanumeric characters"),
						},
						names.AttrDescription: {
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
						"deny_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[A-Za-z][0-9A-Za-z]{0,63}$`),
									"must begin with a letter and contain only alphanumeric characters"),
							},
						},
						"allow_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[A-Za-z][0-9A-Za-z]{0,63}$`),
									"must begin with a letter and contain only alphanumeric characters"),
							},
						},
					},
				},
			},
			// TODO network_function_groups
			"segment_actions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAction: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"share",
								"create-route",
								"send-via",
								"send-to",
							}, false),
						},
						"segment": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[A-Za-z][0-9A-Za-z]{0,63}$`),
								"must begin with a letter and contain only alphanumeric characters"),
						},
						names.AttrMode: {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"attachment-route",
								"single-hop",
								"dual-hop",
							}, false),
						},
						"share_with":        setOfString,
						"share_with_except": setOfString,
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
						"destinations": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.Any(
									validation.StringInSlice([]string{
										"blackhole",
									}, false),
									validation.StringMatch(regexache.MustCompile(`^attachment-[0-9a-f]{17}$`),
										"must be a list of valid attachment ids or a list with only the word \"blackhole\"."),
								),
							},
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"when_sent_to": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"segments": setOfString,
								},
							},
						},
						"via": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_function_groups": setOfString,
									"with_edge_override": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"edge_sets": setOfString,
												"use_edge": {
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
			},
			"attachment_policies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_number": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
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
									names.AttrType: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"account-id",
											"any",
											"tag-value",
											"tag-exists",
											"resource-id",
											names.AttrRegion,
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
									names.AttrKey: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrAction: {
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
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[A-Za-z][0-9A-Za-z]{0,63}$`),
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
									"add_to_network_function_group": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrJSON: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCoreNetworkPolicyDocumentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	mergedDoc := &coreNetworkPolicyDocument{
		Version: d.Get(names.AttrVersion).(string),
	}

	// CoreNetworkConfiguration
	networkConfiguration, err := expandDataCoreNetworkPolicyNetworkConfiguration(d.Get("core_network_configuration").([]interface{}))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "writing Network Manager Core Network Policy Document: %s", err)
	}
	mergedDoc.CoreNetworkConfiguration = networkConfiguration

	// AttachmentPolicies
	attachmentPolicies, err := expandDataCoreNetworkPolicyAttachmentPolicies(d.Get("attachment_policies").([]interface{}))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "writing Network Manager Core Network Policy Document: %s", err)
	}
	mergedDoc.AttachmentPolicies = attachmentPolicies

	// SegmentActions
	segment_actions, err := expandDataCoreNetworkPolicySegmentActions(d.Get("segment_actions").([]interface{}))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "writing Network Manager Core Network Policy Document: %s", err)
	}
	mergedDoc.SegmentActions = segment_actions

	// Segments
	segments, err := expandDataCoreNetworkPolicySegments(d.Get("segments").([]interface{}))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "writing Network Manager Core Network Policy Document: %s", err)
	}
	mergedDoc.Segments = segments

	jsonDoc, err := json.MarshalIndent(mergedDoc, "", "  ")
	if err != nil {
		// should never happen if the above code is correct
		return sdkdiag.AppendErrorf(diags, "writing Network Manager Core Network Policy Document: formatting JSON: %s", err)
	}
	jsonString := string(jsonDoc)

	d.Set(names.AttrJSON, jsonString)
	d.SetId(strconv.Itoa(create.StringHashcode(jsonString)))

	return diags
}

func expandDataCoreNetworkPolicySegmentActions(tfList []interface{}) ([]*coreNetworkPolicySegmentAction, error) {
	apiObjects := make([]*coreNetworkPolicySegmentAction, len(tfList))

	for i, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObject := &coreNetworkPolicySegmentAction{}

		action := tfMap[names.AttrAction].(string)
		apiObject.Action = action
		switch action {
		case "share":
			if v, ok := tfMap["segment"]; ok {
				apiObject.Segment = v.(string)
			}

			if v, ok := tfMap[names.AttrMode]; ok {
				apiObject.Mode = v.(string)
			}

			var shareWith, shareWithExcept interface{}

			if v := tfMap["share_with"].(*schema.Set).List(); len(v) > 0 {
				shareWith = coreNetworkPolicyExpandStringList(v)
				apiObject.ShareWith = shareWith
			}

			if v := tfMap["share_with_except"].(*schema.Set).List(); len(v) > 0 {
				shareWithExcept = coreNetworkPolicyExpandStringList(v)
				apiObject.ShareWithExcept = shareWithExcept
			}

			if (shareWith != nil && shareWithExcept != nil) || (shareWith == nil && shareWithExcept == nil) {
				return nil, fmt.Errorf(`you must specify only 1 of "share_with" or "share_with_except". See segment_actions[%d]`, i)
			}

		case "create-route":
			if v, ok := tfMap["segment"]; ok {
				apiObject.Segment = v.(string)
			}

			if v := tfMap[names.AttrMode]; v != "" {
				return nil, fmt.Errorf(`you cannot specify "mode" if action = "create-route". See segment_actions[%d]`, i)
			}

			if v := tfMap["destination_cidr_blocks"].(*schema.Set).List(); len(v) > 0 {
				apiObject.DestinationCidrBlocks = coreNetworkPolicyExpandStringList(v)
			}

			if v := tfMap["destinations"].(*schema.Set).List(); len(v) > 0 {
				apiObject.Destinations = coreNetworkPolicyExpandStringList(v)
			}

			if v, ok := tfMap[names.AttrDescription]; ok {
				apiObject.Description = v.(string)
			}

		case "send-via", "send-to":
			if v, ok := tfMap["segment"]; ok {
				apiObject.Segment = v.(string)
			}

			if v, ok := tfMap[names.AttrMode]; ok {
				apiObject.Mode = v.(string)
			}

			if v, ok := tfMap["when_sent_to"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				apiObject.WhenSentTo = &coreNetworkPolicySegmentActionWhenSentTo{}

				tfMap := v[0].(map[string]interface{})

				if v := tfMap["segments"].(*schema.Set).List(); len(v) > 0 {
					apiObject.WhenSentTo.Segments = coreNetworkPolicyExpandStringList(v)
				}
			}

			if v, ok := tfMap["via"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				apiObject.Via = &coreNetworkPolicySegmentActionVia{}

				tfMap := v[0].(map[string]interface{})

				if v := tfMap["network_function_groups"].(*schema.Set).List(); len(v) > 0 {
					apiObject.Via.NetworkFunctionGroups = coreNetworkPolicyExpandStringList(v)
				}

				if v, ok := tfMap["with_edge_override"].([]interface{}); ok && len(v) > 0 {
					apiObjects := []*coreNetworkPolicySegmentActionViaEdgeOverride{}

					for _, tfMapRaw := range v {
						tfMap := tfMapRaw.(map[string]interface{})
						apiObject := &coreNetworkPolicySegmentActionViaEdgeOverride{}

						if v := tfMap["edge_sets"].(*schema.Set).List(); len(v) > 0 {
							apiObject.EdgeSets = coreNetworkPolicyExpandStringList(v)
						}

						if v, ok := tfMap["use_edge"]; ok {
							apiObject.UseEdge = v.(string)
						}

						apiObjects = append(apiObjects, apiObject)
					}

					apiObject.Via.WithEdgeOverrides = apiObjects
				}
			}
		}

		apiObjects[i] = apiObject
	}

	return apiObjects, nil
}

func expandDataCoreNetworkPolicyAttachmentPolicies(cfgAttachmentPolicyIntf []interface{}) ([]*coreNetworkAttachmentPolicy, error) {
	aPolicies := make([]*coreNetworkAttachmentPolicy, len(cfgAttachmentPolicyIntf))
	ruleMap := make(map[string]struct{})

	for i, polI := range cfgAttachmentPolicyIntf {
		cfgPol := polI.(map[string]interface{})
		policy := &coreNetworkAttachmentPolicy{}

		rule := cfgPol["rule_number"].(int)
		ruleStr := strconv.Itoa(rule)
		if _, ok := ruleMap[ruleStr]; ok {
			return nil, fmt.Errorf("duplicate Rule Number (%s). Remove the Rule Number or ensure the Rule Number is unique.", ruleStr)
		}
		policy.RuleNumber = rule
		ruleMap[ruleStr] = struct{}{}

		if desc, ok := cfgPol[names.AttrDescription]; ok {
			policy.Description = desc.(string)
		}
		if cL, ok := cfgPol["condition_logic"]; ok {
			policy.ConditionLogic = cL.(string)
		}

		action, err := expandDataCoreNetworkPolicyAttachmentPoliciesAction(cfgPol[names.AttrAction].([]interface{}))
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

func expandDataCoreNetworkPolicyAttachmentPoliciesConditions(tfList []interface{}) ([]*coreNetworkAttachmentPolicyCondition, error) {
	conditions := make([]*coreNetworkAttachmentPolicyCondition, len(tfList))

	for i, condI := range tfList {
		cfgCond := condI.(map[string]interface{})
		condition := &coreNetworkAttachmentPolicyCondition{}
		k := map[string]bool{
			"operator":      false,
			names.AttrKey:   false,
			names.AttrValue: false,
		}

		t := cfgCond[names.AttrType].(string)
		condition.Type = t

		if o := cfgCond["operator"]; o != "" {
			k["operator"] = true
			condition.Operator = o.(string)
		}
		if key := cfgCond[names.AttrKey]; key != "" {
			k[names.AttrKey] = true
			condition.Key = key.(string)
		}
		if v := cfgCond[names.AttrValue]; v != "" {
			k[names.AttrValue] = true
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
			if !k[names.AttrKey] || k["operator"] || k[names.AttrValue] {
				return nil, fmt.Errorf("Conditions %s: You must set \"key\" and cannot set \"operator\", or \"value\" if type = \"tag-exists\".", strconv.Itoa(i))
			}
		}
		if t == "tag-value" {
			if !k[names.AttrKey] || !k["operator"] || !k[names.AttrValue] {
				return nil, fmt.Errorf("Conditions %s: You must set \"key\", \"operator\", and \"value\" if type = \"tag-value\".", strconv.Itoa(i))
			}
		}
		if t == names.AttrRegion || t == "resource-id" || t == "account-id" {
			if k[names.AttrKey] || !k["operator"] || !k[names.AttrValue] {
				return nil, fmt.Errorf("Conditions %s: You must set \"value\" and \"operator\" and cannot set \"key\" if type = \"region\", \"resource-id\", or \"account-id\".", strconv.Itoa(i))
			}
		}
		if t == "attachment-type" {
			if k[names.AttrKey] || !k[names.AttrValue] || cfgCond["operator"].(string) != "equals" {
				return nil, fmt.Errorf("Conditions %s: You must set \"value\", cannot set \"key\" and \"operator\" must be \"equals\" if type = \"attachment-type\".", strconv.Itoa(i))
			}
		}
		conditions[i] = condition
	}
	return conditions, nil
}

func expandDataCoreNetworkPolicyAttachmentPoliciesAction(tfList []interface{}) (*coreNetworkAttachmentPolicyAction, error) {
	tfMap := tfList[0].(map[string]interface{})

	associationMethod := tfMap["association_method"].(string)
	apiObject := &coreNetworkAttachmentPolicyAction{
		AssociationMethod: associationMethod,
	}

	if v := tfMap["segment"]; v != "" {
		if associationMethod == "tag" {
			return nil, fmt.Errorf(`cannot set "segment" argument if association_method = "tag"`)
		}
		apiObject.Segment = v.(string)
	}

	if v := tfMap["tag_value_of_key"]; v != "" {
		if associationMethod == "constant" {
			return nil, fmt.Errorf(`cannot set "tag_value_of_key" argument if association_method = "constant"`)
		}
		apiObject.TagValueOfKey = v.(string)
	}

	if v, ok := tfMap["require_acceptance"]; ok {
		apiObject.RequireAcceptance = v.(bool)
	}

	if v := tfMap["add_to_network_function_group"]; v != "" {
		apiObject.AddToNetworkFunctionGroup = v.(string)
	}

	return apiObject, nil
}

func expandDataCoreNetworkPolicySegments(cfgSgmtIntf []interface{}) ([]*coreNetworkPolicySegment, error) {
	Sgmts := make([]*coreNetworkPolicySegment, len(cfgSgmtIntf))
	nameMap := make(map[string]struct{})

	for i, sgmtI := range cfgSgmtIntf {
		cfgSgmt := sgmtI.(map[string]interface{})
		sgmt := &coreNetworkPolicySegment{}

		if name, ok := cfgSgmt[names.AttrName]; ok {
			if _, ok := nameMap[name.(string)]; ok {
				return nil, fmt.Errorf("duplicate Name (%s). Remove the Name or ensure the Name is unique.", name.(string))
			}
			sgmt.Name = name.(string)
			if len(sgmt.Name) > 0 {
				nameMap[sgmt.Name] = struct{}{}
			}
		}
		if description, ok := cfgSgmt[names.AttrDescription]; ok {
			sgmt.Description = description.(string)
		}
		if actions := cfgSgmt["allow_filter"].(*schema.Set).List(); len(actions) > 0 {
			sgmt.AllowFilter = coreNetworkPolicyExpandStringList(actions)
		}
		if actions := cfgSgmt["deny_filter"].(*schema.Set).List(); len(actions) > 0 {
			sgmt.DenyFilter = coreNetworkPolicyExpandStringList(actions)
		}
		if edgeLocations := cfgSgmt["edge_locations"].(*schema.Set).List(); len(edgeLocations) > 0 {
			sgmt.EdgeLocations = coreNetworkPolicyExpandStringList(edgeLocations)
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

func expandDataCoreNetworkPolicyNetworkConfiguration(networkCfgIntf []interface{}) (*coreNetworkPolicyCoreNetworkConfiguration, error) {
	m := networkCfgIntf[0].(map[string]interface{})

	nc := &coreNetworkPolicyCoreNetworkConfiguration{}

	nc.AsnRanges = coreNetworkPolicyExpandStringList(m["asn_ranges"].(*schema.Set).List())

	if cidrs := m["inside_cidr_blocks"].(*schema.Set).List(); len(cidrs) > 0 {
		nc.InsideCidrBlocks = coreNetworkPolicyExpandStringList(cidrs)
	}

	nc.VpnEcmpSupport = m["vpn_ecmp_support"].(bool)

	el, err := expandDataCoreNetworkPolicyNetworkConfigurationEdgeLocations(m["edge_locations"].([]interface{}))

	if err != nil {
		return nil, err
	}
	nc.EdgeLocations = el

	return nc, nil
}

func expandDataCoreNetworkPolicyNetworkConfigurationEdgeLocations(tfList []interface{}) ([]*coreNetworkPolicyCoreNetworkEdgeLocation, error) {
	edgeLocations := make([]*coreNetworkPolicyCoreNetworkEdgeLocation, len(tfList))
	locMap := make(map[string]struct{})

	for i, edgeLocationsRaw := range tfList {
		cfgEdgeLocation, ok := edgeLocationsRaw.(map[string]interface{})
		edgeLocation := &coreNetworkPolicyCoreNetworkEdgeLocation{}

		if !ok {
			continue
		}

		location := cfgEdgeLocation[names.AttrLocation].(string)

		if _, ok := locMap[location]; ok {
			return nil, fmt.Errorf("duplicate Location (%s). Remove the Location or ensure the Location is unique.", location)
		}
		edgeLocation.Location = location
		if len(edgeLocation.Location) > 0 {
			locMap[edgeLocation.Location] = struct{}{}
		}

		if v, ok := cfgEdgeLocation["asn"].(string); ok && v != "" {
			v, err := strconv.ParseInt(v, 10, 64)

			if err != nil {
				return nil, err
			}

			edgeLocation.Asn = v
		}

		if cidrs := cfgEdgeLocation["inside_cidr_blocks"].([]interface{}); len(cidrs) > 0 {
			edgeLocation.InsideCidrBlocks = coreNetworkPolicyExpandStringList(cidrs)
		}

		edgeLocations[i] = edgeLocation
	}
	return edgeLocations, nil
}
