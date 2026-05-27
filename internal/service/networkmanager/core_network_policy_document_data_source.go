// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package networkmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkmanager_core_network_policy_document", name="Core Network Policy Document")
func dataSourceCoreNetworkPolicyDocument() *schema.Resource {
	setOfStringOptional := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
	setOfStringRequired := &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCoreNetworkPolicyDocumentRead,

		// Order attributes to match model structures and documentation:
		// https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-policies-json.html.
		// Consciously NOT sorted alphabetically.
		Schema: map[string]*schema.Schema{
			names.AttrVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2021.12",
				ValidateFunc: validation.StringInSlice([]string{
					"2021.12",
					"2025.11",
				}, false),
			},
			"core_network_configuration": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"asn_ranges":         setOfStringRequired,
						"inside_cidr_blocks": setOfStringOptional,
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
						"dns_support": {
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
						},
						"security_group_referencing_support": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
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
								"associate-routing-policy",
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
						"share_with":           setOfStringOptional,
						"share_with_except":    setOfStringOptional,
						"routing_policy_names": setOfStringOptional,
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
									"segments": setOfStringOptional,
								},
							},
						},
						"via": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_function_groups": setOfStringOptional,
									"with_edge_override": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"edge_sets": {
													Type: schema.TypeSet,
													Elem: &schema.Schema{
														Type: schema.TypeSet,
														Elem: &schema.Schema{
															Type: schema.TypeString,
														},
													},
													Optional: true,
												},
												"use_edge_location": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"use_edge": {
													Type:       schema.TypeString,
													Optional:   true,
													Deprecated: "use_edge is deprecated. Use use_edge_location instead.",
												},
											},
										},
									},
								},
							},
						},
						"edge_location_association": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"edge_location": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidRegionName,
									},
									"peer_edge_location": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidRegionName,
									},
									"routing_policy_names": setOfStringRequired,
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
											"tag-name",
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
			"attachment_routing_policy_rules": {
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
						"edge_locations": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidRegionName,
							},
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
											"routing-policy-label",
										}, false),
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
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
									"associate_routing_policies": setOfStringRequired,
								},
							},
						},
					},
				},
			},
			"routing_policies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"routing_policy_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`),
									"must contain only alphanumeric characters"),
							),
						},
						"routing_policy_description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"routing_policy_direction": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"inbound",
								"outbound",
							}, false),
						},
						"routing_policy_number": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 9999),
						},
						"routing_policy_rules": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"rule_number": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 9999),
									},
									"rule_definition": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"match_conditions": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrType: {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.StringInSlice([]string{
																	"prefix-equals",
																	"prefix-in-cidr",
																	"prefix-in-prefix-list",
																	"asn-in-as-path",
																	"community-in-list",
																	"med-equals",
																}, false),
															},
															names.AttrValue: {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
												"condition_logic": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														"and",
														"or",
													}, false),
												},
												names.AttrAction: {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrType: {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.StringInSlice([]string{
																	"drop",
																	"allow",
																	"summarize",
																	"prepend-asn-list",
																	"remove-asn-list",
																	"replace-asn-list",
																	"add-community",
																	"remove-community",
																	"set-med",
																	"set-local-preference",
																}, false),
															},
															names.AttrValue: {
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

func dataSourceCoreNetworkPolicyDocumentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	mergedDoc := &coreNetworkPolicyDocument{
		Version: d.Get(names.AttrVersion).(string),
	}

	// CoreNetworkConfiguration
	networkConfiguration, err := expandCoreNetworkPolicyCoreNetworkConfiguration(d.Get("core_network_configuration").([]any))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.CoreNetworkConfiguration = networkConfiguration

	// Segments
	segments, err := expandCoreNetworkPolicySegments(d.Get("segments").([]any))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.Segments = segments

	// RoutingPolicies
	routingPolicies, err := expandCoreNetworkPolicyRoutingPolicies(d.Get("routing_policies").([]any), mergedDoc.Version)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.RoutingPolicies = routingPolicies

	// NetworkFunctionGroups
	networkFunctionGroups, err := expandCoreNetworkPolicyNetworkFunctionGroups(d.Get("network_function_groups").([]any))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.NetworkFunctionGroups = networkFunctionGroups

	// SegmentActions
	segment_actions, err := expandCoreNetworkPolicySegmentActions(d.Get("segment_actions").([]any))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.SegmentActions = segment_actions

	// AttachmentPolicies
	attachmentPolicies, err := expandCoreNetworkPolicyAttachmentPolicies(d.Get("attachment_policies").([]any))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.AttachmentPolicies = attachmentPolicies

	// AttachmentRoutingPolicyRules
	attachmentRoutingPolicyRules, err := expandCoreNetworkPolicyAttachmentRoutingPolicyRules(d.Get("attachment_routing_policy_rules").([]any), mergedDoc.Version)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	mergedDoc.AttachmentRoutingPolicyRules = attachmentRoutingPolicyRules

	jsonDoc, err := json.MarshalIndent(mergedDoc, "", "  ")
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	jsonString := string(jsonDoc)

	// Debug: Print the generated JSON policy document
	log.Printf("[DEBUG] Generated Core Network Policy Document JSON:\n%s", jsonString)

	d.Set(names.AttrJSON, jsonString)
	d.SetId(strconv.Itoa(create.StringHashcode(jsonString)))

	return diags
}

func expandCoreNetworkPolicySegmentActions(tfList []any) ([]*coreNetworkPolicySegmentAction, error) {
	apiObjects := make([]*coreNetworkPolicySegmentAction, 0)

	for i, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
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

			var shareWith, shareWithExcept any

			if v := tfMap["share_with"].(*schema.Set).List(); len(v) > 0 {
				shareWith = coreNetworkPolicyExpandStringList(v)
				apiObject.ShareWith = shareWith
			}

			if v := tfMap["share_with_except"].(*schema.Set).List(); len(v) > 0 {
				shareWithExcept = coreNetworkPolicyExpandStringList(v)
				apiObject.ShareWithExcept = shareWithExcept
			}

			if v := tfMap["routing_policy_names"].(*schema.Set).List(); len(v) > 0 {
				apiObject.RoutingPolicyNames = coreNetworkPolicyExpandStringList(v)
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

		case "associate-routing-policy":
			if v, ok := tfMap["segment"]; ok {
				apiObject.Segment = v.(string)
			}

			if v, ok := tfMap["edge_location_association"].([]any); ok && len(v) > 0 && v[0] != nil {
				apiObject.EdgeLocationAssociation = &coreNetworkPolicySegmentActionEdgeLocationAssociation{}

				tfMap := v[0].(map[string]any)

				if v, ok := tfMap["edge_location"].(string); ok && v != "" {
					apiObject.EdgeLocationAssociation.EdgeLocation = v
				}

				if v, ok := tfMap["peer_edge_location"].(string); ok && v != "" {
					apiObject.EdgeLocationAssociation.PeerEdgeLocation = v
				}

				if v := tfMap["routing_policy_names"].(*schema.Set).List(); len(v) > 0 {
					apiObject.EdgeLocationAssociation.RoutingPolicyNames = coreNetworkPolicyExpandStringList(v)
				}
			}

		case "send-via", "send-to":
			if v, ok := tfMap["segment"]; ok {
				apiObject.Segment = v.(string)
			}

			if v, ok := tfMap[names.AttrMode]; ok {
				apiObject.Mode = v.(string)
			}

			if v, ok := tfMap["when_sent_to"].([]any); ok && len(v) > 0 && v[0] != nil {
				apiObject.WhenSentTo = &coreNetworkPolicySegmentActionWhenSentTo{}

				tfMap := v[0].(map[string]any)

				if v := tfMap["segments"].(*schema.Set).List(); len(v) > 0 {
					apiObject.WhenSentTo.Segments = coreNetworkPolicyExpandStringList(v)
				}
			}

			if v, ok := tfMap["via"].([]any); ok && len(v) > 0 && v[0] != nil {
				apiObject.Via = &coreNetworkPolicySegmentActionVia{}

				tfMap := v[0].(map[string]any)

				if v := tfMap["network_function_groups"].(*schema.Set).List(); len(v) > 0 {
					apiObject.Via.NetworkFunctionGroups = coreNetworkPolicyExpandStringList(v)
				}

				if v, ok := tfMap["with_edge_override"].([]any); ok && len(v) > 0 {
					apiObjects := []*coreNetworkPolicySegmentActionViaEdgeOverride{}

					for _, tfMapRaw := range v {
						tfMap := tfMapRaw.(map[string]any)
						apiObject := &coreNetworkPolicySegmentActionViaEdgeOverride{}

						if v := tfMap["edge_sets"].(*schema.Set).List(); len(v) > 0 {
							var edgeSets [][]string
							for _, esRaw := range v {
								es := esRaw.(*schema.Set)
								edgeSets = append(edgeSets, flex.ExpandStringValueSet(es))
							}
							apiObject.EdgeSets = edgeSets
						}

						if v, ok := tfMap["use_edge_location"]; ok && v != "" {
							apiObject.UseEdgeLocation = v.(string)
						} else if v, ok := tfMap["use_edge"]; ok {
							apiObject.UseEdgeLocation = v.(string)
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

func expandCoreNetworkPolicyAttachmentPolicies(tfList []any) ([]*coreNetworkPolicyAttachmentPolicy, error) {
	apiObjects := make([]*coreNetworkPolicyAttachmentPolicy, 0)
	ruleMap := make(map[int]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
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

		action, err := expandDataCoreNetworkPolicyAttachmentPoliciesAction(tfMap[names.AttrAction].([]any))
		if err != nil {
			return nil, err
		}
		apiObject.Action = action

		conditions, err := expandDataCoreNetworkPolicyAttachmentPoliciesConditions(tfMap["conditions"].([]any))
		if err != nil {
			return nil, err
		}
		apiObject.Conditions = conditions

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandDataCoreNetworkPolicyAttachmentPoliciesConditions(tfList []any) ([]*coreNetworkPolicyAttachmentPolicyCondition, error) {
	apiObjects := make([]*coreNetworkPolicyAttachmentPolicyCondition, 0)

	for i, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
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
					return nil, fmt.Errorf("conditions %d: cannot set \"operator\", \"key\", or \"value\" if type = \"any\"", i)
				}
			}

		case "tag-exists", "tag-name":
			if !k[names.AttrKey] || k["operator"] || k[names.AttrValue] {
				return nil, fmt.Errorf("conditions %d: must set \"key\" and cannot set \"operator\" or \"value\" if type = \"tag-exists\" or \"tag-name\"", i)
			}

		case "tag-value":
			if !k[names.AttrKey] || !k["operator"] || !k[names.AttrValue] {
				return nil, fmt.Errorf("conditions %d: must set \"key\", \"operator\", and \"value\" if type = \"tag-value\"", i)
			}

		case names.AttrRegion, "resource-id", "account-id":
			if k[names.AttrKey] || !k["operator"] || !k[names.AttrValue] {
				return nil, fmt.Errorf("conditions %d: must set \"value\" and \"operator\" and cannot set \"key\" if type = \"region\", \"resource-id\", or \"account-id\"\n%[2]t, %[3]t, %[4]t", i, k[names.AttrKey], k["operator"], k[names.AttrValue])
			}

		case "attachment-type":
			if k[names.AttrKey] || !k[names.AttrValue] || tfMap["operator"].(string) != "equals" {
				return nil, fmt.Errorf("conditions %d: must set \"value\", cannot set \"key\" and \"operator\" must be \"equals\" if type = \"attachment-type\"", i)
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandDataCoreNetworkPolicyAttachmentPoliciesAction(tfList []any) (*coreNetworkPolicyAttachmentPolicyAction, error) {
	tfMap := tfList[0].(map[string]any)

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

func expandCoreNetworkPolicyAttachmentRoutingPolicyRules(tfList []any, version string) ([]*coreNetworkPolicyAttachmentRoutingPolicyRule, error) {
	if len(tfList) > 0 && version != "2025.11" {
		return nil, fmt.Errorf("attachment_routing_policy_rules requires version 2025.11")
	}

	apiObjects := make([]*coreNetworkPolicyAttachmentRoutingPolicyRule, 0)
	ruleMap := make(map[int]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicyAttachmentRoutingPolicyRule{}

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

		if v := tfMap["edge_locations"].(*schema.Set).List(); len(v) > 0 {
			apiObject.EdgeLocations = coreNetworkPolicyExpandStringList(v)
		}

		apiObject.Conditions = expandCoreNetworkPolicyAttachmentRoutingPolicyRulesConditions(tfMap["conditions"].([]any))

		apiObject.Action = expandCoreNetworkPolicyAttachmentRoutingPolicyRulesAction(tfMap[names.AttrAction].([]any))

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandCoreNetworkPolicyAttachmentRoutingPolicyRulesConditions(tfList []any) []*coreNetworkPolicyAttachmentRoutingPolicyRuleCondition {
	apiObjects := make([]*coreNetworkPolicyAttachmentRoutingPolicyRuleCondition, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicyAttachmentRoutingPolicyRuleCondition{}

		if v, ok := tfMap[names.AttrType].(string); ok {
			apiObject.Type = v
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = v
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCoreNetworkPolicyAttachmentRoutingPolicyRulesAction(tfList []any) *coreNetworkPolicyAttachmentRoutingPolicyRuleAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &coreNetworkPolicyAttachmentRoutingPolicyRuleAction{}

	if v := tfMap["associate_routing_policies"].(*schema.Set).List(); len(v) > 0 {
		apiObject.AssociateRoutingPolicies = coreNetworkPolicyExpandStringList(v)
	}

	return apiObject
}

func expandCoreNetworkPolicySegments(tfList []any) ([]*coreNetworkPolicySegment, error) {
	apiObjects := make([]*coreNetworkPolicySegment, 0)
	nameMap := make(map[string]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
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

func expandCoreNetworkPolicyRoutingPolicies(tfList []any, version string) ([]*coreNetworkPolicyRoutingPolicy, error) {
	if len(tfList) > 0 && version != "2025.11" {
		return nil, fmt.Errorf("routing_policies requires version 2025.11")
	}

	apiObjects := make([]*coreNetworkPolicyRoutingPolicy, 0)
	nameMap := make(map[string]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicyRoutingPolicy{}

		if v, ok := tfMap["routing_policy_name"].(string); ok {
			if _, ok := nameMap[v]; ok {
				return nil, fmt.Errorf("duplicate routing_policy_name (%s). Remove the name or ensure it is unique", v)
			}
			apiObject.RoutingPolicyName = v
			if len(apiObject.RoutingPolicyName) > 0 {
				nameMap[apiObject.RoutingPolicyName] = struct{}{}
			}
		}

		if v, ok := tfMap["routing_policy_description"].(string); ok && v != "" {
			apiObject.RoutingPolicyDescription = v
		}

		if v, ok := tfMap["routing_policy_direction"].(string); ok {
			apiObject.RoutingPolicyDirection = v
		}

		if v, ok := tfMap["routing_policy_number"].(int); ok {
			apiObject.RoutingPolicyNumber = v
		}

		rules, err := expandCoreNetworkPolicyRoutingPolicyRules(tfMap["routing_policy_rules"].([]any))
		if err != nil {
			return nil, err
		}
		apiObject.RoutingPolicyRules = rules

		// Validate that summarize action is only used with outbound policies
		if apiObject.RoutingPolicyDirection == "inbound" {
			for _, rule := range rules {
				if rule.RuleDefinition != nil && rule.RuleDefinition.Action != nil && rule.RuleDefinition.Action.Type == "summarize" {
					return nil, fmt.Errorf("summarize action cannot be used for inbound routing policies (routing_policy_name: %s, rule_number: %d)", apiObject.RoutingPolicyName, rule.RuleNumber)
				}
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandCoreNetworkPolicyRoutingPolicyRules(tfList []any) ([]*coreNetworkPolicyRoutingPolicyRule, error) {
	apiObjects := make([]*coreNetworkPolicyRoutingPolicyRule, 0)
	ruleMap := make(map[int]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicyRoutingPolicyRule{}

		if v, ok := tfMap["rule_number"].(int); ok {
			if _, ok := ruleMap[v]; ok {
				return nil, fmt.Errorf("duplicate rule_number (%d) in routing policy rules. Remove the rule number or ensure it is unique", v)
			}
			apiObject.RuleNumber = v
			ruleMap[apiObject.RuleNumber] = struct{}{}
		}

		if v, ok := tfMap["rule_definition"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.RuleDefinition = expandCoreNetworkPolicyRoutingPolicyRuleDefinition(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandCoreNetworkPolicyRoutingPolicyRuleDefinition(tfList []any) *coreNetworkPolicyRoutingPolicyRuleDefinition {
	tfMap := tfList[0].(map[string]any)

	apiObject := &coreNetworkPolicyRoutingPolicyRuleDefinition{}

	// AWS requires condition-logic to always be present, default to "and"
	if v, ok := tfMap["condition_logic"].(string); ok && v != "" {
		apiObject.ConditionLogic = v
	} else {
		apiObject.ConditionLogic = "and"
	}

	if v, ok := tfMap["match_conditions"].([]any); ok && len(v) > 0 {
		apiObject.MatchConditions = expandCoreNetworkPolicyRoutingPolicyRuleMatchConditions(v)
	}

	if v, ok := tfMap[names.AttrAction].([]any); ok && len(v) > 0 {
		apiObject.Action = expandCoreNetworkPolicyRoutingPolicyRuleAction(v)
	}

	return apiObject
}

func expandCoreNetworkPolicyRoutingPolicyRuleMatchConditions(tfList []any) []*coreNetworkPolicyRoutingPolicyRuleMatchCondition {
	apiObjects := make([]*coreNetworkPolicyRoutingPolicyRuleMatchCondition, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := &coreNetworkPolicyRoutingPolicyRuleMatchCondition{}

		if v, ok := tfMap[names.AttrType].(string); ok {
			apiObject.Type = v
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = v
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCoreNetworkPolicyRoutingPolicyRuleAction(tfList []any) *coreNetworkPolicyRoutingPolicyRuleAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &coreNetworkPolicyRoutingPolicyRuleAction{}

	if v, ok := tfMap[names.AttrType].(string); ok {
		apiObject.Type = v
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = v
	}

	return apiObject
}

func expandCoreNetworkPolicyCoreNetworkConfiguration(tfList []any) (*coreNetworkPolicyCoreNetworkConfiguration, error) {
	tfMap := tfList[0].(map[string]any)
	apiObject := &coreNetworkPolicyCoreNetworkConfiguration{}

	apiObject.AsnRanges = coreNetworkPolicyExpandStringList(tfMap["asn_ranges"].(*schema.Set).List())

	if v := tfMap["inside_cidr_blocks"].(*schema.Set).List(); len(v) > 0 {
		apiObject.InsideCidrBlocks = coreNetworkPolicyExpandStringList(v)
	}

	apiObject.VpnEcmpSupport = tfMap["vpn_ecmp_support"].(bool)

	el, err := expandDataCoreNetworkPolicyNetworkConfigurationEdgeLocations(tfMap["edge_locations"].([]any))
	if err != nil {
		return nil, err
	}
	apiObject.EdgeLocations = el

	if v, ok := tfMap["dns_support"].(bool); ok {
		apiObject.DnsSupport = v
	}

	if v, ok := tfMap["security_group_referencing_support"].(bool); ok {
		apiObject.SecurityGroupReferencingSupport = v
	}

	return apiObject, nil
}

func expandDataCoreNetworkPolicyNetworkConfigurationEdgeLocations(tfList []any) ([]*coreNetworkPolicyCoreNetworkEdgeLocation, error) {
	apiObjects := make([]*coreNetworkPolicyCoreNetworkEdgeLocation, 0)
	locationMap := make(map[string]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
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

		if v := tfMap["inside_cidr_blocks"].([]any); len(v) > 0 {
			apiObject.InsideCidrBlocks = coreNetworkPolicyExpandStringList(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func expandCoreNetworkPolicyNetworkFunctionGroups(tfList []any) ([]*coreNetworkPolicyNetworkFunctionGroup, error) {
	apiObjects := make([]*coreNetworkPolicyNetworkFunctionGroup, 0)
	nameMap := make(map[string]struct{})

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
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
