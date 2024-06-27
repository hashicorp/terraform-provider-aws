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
			"network_function_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"require_attachment_acceptance": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
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
										Optional: true,
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
	networkConfiguration, err := expandCoreNetworkPolicyCoreNetworkConfiguration(d.Get("core_network_configuration").([]interface{}))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.CoreNetworkConfiguration = networkConfiguration

	// Segments
	segments, err := expandCoreNetworkPolicySegments(d.Get("segments").([]interface{}))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.Segments = segments

	// NetworkFunctionGroups
	networkFunctionGroups, err := expandCoreNetworkPolicyNetworkFunctionGroups(d.Get("network_function_groups").([]interface{}))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.NetworkFunctionGroups = networkFunctionGroups

	// SegmentActions
	segment_actions, err := expandCoreNetworkPolicySegmentActions(d.Get("segment_actions").([]interface{}))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.SegmentActions = segment_actions

	// AttachmentPolicies
	attachmentPolicies, err := expandCoreNetworkPolicyAttachmentPolicies(d.Get("attachment_policies").([]interface{}))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.AttachmentPolicies = attachmentPolicies

	jsonDoc, err := json.MarshalIndent(mergedDoc, "", "  ")
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	jsonString := string(jsonDoc)

	d.Set(names.AttrJSON, jsonString)
	d.SetId(strconv.Itoa(create.StringHashcode(jsonString)))

	return diags
}

func expandCoreNetworkPolicySegmentActions(tfList []interface{}) ([]*coreNetworkPolicySegmentAction, error) {
	apiObjects := make([]*coreNetworkPolicySegmentAction, 0)

	for i, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

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

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandCoreNetworkPolicyAttachmentPolicies(tfList []interface{}) ([]*coreNetworkPolicyAttachmentPolicy, error) {
	apiObjects := make([]*coreNetworkPolicyAttachmentPolicy, 0)
	ruleMap := make(map[int]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicyAttachmentPolicy{}

		if v, ok := tfMap["rule_number"].(int); ok {
			if _, ok := ruleMap[v]; ok {
				return nil, fmt.Errorf("duplicate Rule Number (%d). Remove the Rule Number or ensure the Rule Number is unique", v)
			}
			apiObject.RuleNumber = v
			ruleMap[apiObject.RuleNumber] = struct{}{}
		}

		if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
			apiObject.Description = v
		}

		if v, ok := tfMap["condition_logic"].(string); ok && v != "" {
			apiObject.ConditionLogic = v
		}

		action, err := expandDataCoreNetworkPolicyAttachmentPoliciesAction(tfMap[names.AttrAction].([]interface{}))
		if err != nil {
			return nil, err
		}
		apiObject.Action = action

		conditions, err := expandDataCoreNetworkPolicyAttachmentPoliciesConditions(tfMap["conditions"].([]interface{}))
		if err != nil {
			return nil, err
		}
		apiObject.Conditions = conditions

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandDataCoreNetworkPolicyAttachmentPoliciesConditions(tfList []interface{}) ([]*coreNetworkPolicyAttachmentPolicyCondition, error) {
	apiObjects := make([]*coreNetworkPolicyAttachmentPolicyCondition, 0)

	for i, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicyAttachmentPolicyCondition{}
		k := map[string]bool{
			"operator":      false,
			names.AttrKey:   false,
			names.AttrValue: false,
		}

		typ := tfMap[names.AttrType].(string)
		apiObject.Type = typ

		if v, ok := tfMap["operator"].(string); ok && v != "" {
			k["operator"] = true
			apiObject.Operator = v
		}

		if v := tfMap[names.AttrKey].(string); ok && v != "" {
			k[names.AttrKey] = true
			apiObject.Key = v
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			k[names.AttrValue] = true
			apiObject.Value = v
		}

		switch typ {
		case "any":
			for _, key := range k {
				if key {
					return nil, fmt.Errorf("Conditions %d: You cannot set \"operator\", \"key\", or \"value\" if type = \"any\".", i)
				}
			}

		case "tag-exists":
			if !k[names.AttrKey] || k["operator"] || k[names.AttrValue] {
				return nil, fmt.Errorf("Conditions %d: You must set \"key\" and cannot set \"operator\", or \"value\" if type = \"tag-exists\".", i)
			}

		case "tag-value":
			if !k[names.AttrKey] || !k["operator"] || !k[names.AttrValue] {
				return nil, fmt.Errorf("Conditions %d: You must set \"key\", \"operator\", and \"value\" if type = \"tag-value\".", i)
			}

		case names.AttrRegion, "resource-id", "account-id":
			if k[names.AttrKey] || !k["operator"] || !k[names.AttrValue] {
				return nil, fmt.Errorf("Conditions %d: You must set \"value\" and \"operator\" and cannot set \"key\" if type = \"region\", \"resource-id\", or \"account-id\".", i)
			}

		case "attachment-type":
			if k[names.AttrKey] || !k[names.AttrValue] || tfMap["operator"].(string) != "equals" {
				return nil, fmt.Errorf("Conditions %d: You must set \"value\", cannot set \"key\" and \"operator\" must be \"equals\" if type = \"attachment-type\".", i)
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandDataCoreNetworkPolicyAttachmentPoliciesAction(tfList []interface{}) (*coreNetworkPolicyAttachmentPolicyAction, error) {
	tfMap := tfList[0].(map[string]interface{})

	associationMethod := tfMap["association_method"].(string)
	apiObject := &coreNetworkPolicyAttachmentPolicyAction{
		AssociationMethod: associationMethod,
	}

	if v, ok := tfMap["segment"].(string); ok && v != "" {
		if associationMethod == "tag" {
			return nil, fmt.Errorf(`cannot set "segment" argument if association_method = "tag"`)
		}
		apiObject.Segment = v
	}

	if v, ok := tfMap["tag_value_of_key"].(string); ok && v != "" {
		if associationMethod == "constant" {
			return nil, fmt.Errorf(`cannot set "tag_value_of_key" argument if association_method = "constant"`)
		}
		apiObject.TagValueOfKey = v
	}

	if v, ok := tfMap["require_acceptance"].(bool); ok {
		apiObject.RequireAcceptance = v
	}

	if v, ok := tfMap["add_to_network_function_group"].(string); ok && v != "" {
		apiObject.AddToNetworkFunctionGroup = v
	}

	return apiObject, nil
}

func expandCoreNetworkPolicySegments(tfList []interface{}) ([]*coreNetworkPolicySegment, error) {
	apiObjects := make([]*coreNetworkPolicySegment, 0)
	nameMap := make(map[string]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicySegment{}

		if v, ok := tfMap[names.AttrName].(string); ok {
			if _, ok := nameMap[v]; ok {
				return nil, fmt.Errorf("duplicate Name (%s). Remove the Name or ensure the Name is unique", v)
			}
			apiObject.Name = v
			if len(apiObject.Name) > 0 {
				nameMap[apiObject.Name] = struct{}{}
			}
		}

		if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
			apiObject.Description = v
		}

		if v := tfMap["allow_filter"].(*schema.Set).List(); len(v) > 0 {
			apiObject.AllowFilter = coreNetworkPolicyExpandStringList(v)
		}

		if v := tfMap["deny_filter"].(*schema.Set).List(); len(v) > 0 {
			apiObject.DenyFilter = coreNetworkPolicyExpandStringList(v)
		}

		if v := tfMap["edge_locations"].(*schema.Set).List(); len(v) > 0 {
			apiObject.EdgeLocations = coreNetworkPolicyExpandStringList(v)
		}

		if v, ok := tfMap["require_attachment_acceptance"].(bool); ok {
			apiObject.RequireAttachmentAcceptance = v
		}

		if v, ok := tfMap["isolate_attachments"].(bool); ok {
			apiObject.IsolateAttachments = v
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandCoreNetworkPolicyCoreNetworkConfiguration(tfList []interface{}) (*coreNetworkPolicyCoreNetworkConfiguration, error) {
	tfMap := tfList[0].(map[string]interface{})
	apiObject := &coreNetworkPolicyCoreNetworkConfiguration{}

	apiObject.AsnRanges = coreNetworkPolicyExpandStringList(tfMap["asn_ranges"].(*schema.Set).List())

	if v := tfMap["inside_cidr_blocks"].(*schema.Set).List(); len(v) > 0 {
		apiObject.InsideCidrBlocks = coreNetworkPolicyExpandStringList(v)
	}

	apiObject.VpnEcmpSupport = tfMap["vpn_ecmp_support"].(bool)

	el, err := expandDataCoreNetworkPolicyNetworkConfigurationEdgeLocations(tfMap["edge_locations"].([]interface{}))
	if err != nil {
		return nil, err
	}
	apiObject.EdgeLocations = el

	return apiObject, nil
}

func expandDataCoreNetworkPolicyNetworkConfigurationEdgeLocations(tfList []interface{}) ([]*coreNetworkPolicyCoreNetworkEdgeLocation, error) {
	apiObjects := make([]*coreNetworkPolicyCoreNetworkEdgeLocation, 0)
	locationMap := make(map[string]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicyCoreNetworkEdgeLocation{}

		if v, ok := tfMap[names.AttrLocation].(string); ok {
			if _, ok := locationMap[v]; ok {
				return nil, fmt.Errorf("duplicate Location (%s). Remove the Location or ensure the Location is unique", v)
			}
			apiObject.Location = v
			if len(apiObject.Location) > 0 {
				locationMap[apiObject.Location] = struct{}{}
			}
		}

		if v, ok := tfMap["asn"].(string); ok && v != "" {
			v, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, err
			}
			apiObject.Asn = v
		}

		if v := tfMap["inside_cidr_blocks"].([]interface{}); len(v) > 0 {
			apiObject.InsideCidrBlocks = coreNetworkPolicyExpandStringList(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandCoreNetworkPolicyNetworkFunctionGroups(tfList []interface{}) ([]*coreNetworkPolicyNetworkFunctionGroup, error) {
	apiObjects := make([]*coreNetworkPolicyNetworkFunctionGroup, 0)
	nameMap := make(map[string]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicyNetworkFunctionGroup{}

		if v, ok := tfMap[names.AttrName].(string); ok {
			if _, ok := nameMap[v]; ok {
				return nil, fmt.Errorf("duplicate Name (%s). Remove the Name or ensure the Name is unique", v)
			}
			apiObject.Name = v
			if len(apiObject.Name) > 0 {
				nameMap[apiObject.Name] = struct{}{}
			}
		}

		if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
			apiObject.Description = v
		}

		if v, ok := tfMap["require_attachment_acceptance"].(bool); ok {
			apiObject.RequireAttachmentAcceptance = v
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}
