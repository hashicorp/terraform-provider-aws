// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkfirewall_rule_group", name="Rule Group")
// @Tags(identifierAttribute="id")
func resourceRuleGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleGroupCreate,
		ReadWithoutTimeout:   resourceRuleGroupRead,
		UpdateWithoutTimeout: resourceRuleGroupUpdate,
		DeleteWithoutTimeout: resourceRuleGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"capacity": {
					Type:     schema.TypeInt,
					Required: true,
					ForceNew: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrEncryptionConfiguration: encryptionConfigurationSchema(),
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"rule_group": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"reference_sets": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"ip_set_references": {
											Type:     schema.TypeSet,
											Optional: true,
											MaxItems: 5,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"ip_set_reference": {
														Type:     schema.TypeList,
														Required: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"reference_arn": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: verify.ValidARN,
																},
															},
														},
													},
													names.AttrKey: {
														Type:     schema.TypeString,
														Required: true,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 32),
															validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
															validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_]+$`), "must contain only alphanumeric and underscore characters"),
														),
													},
												},
											},
										},
									},
								},
							},
							"rules_source": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"rules_source_list": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"generated_rules_type": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.GeneratedRulesType](),
													},
													"target_types": {
														Type:     schema.TypeSet,
														Required: true,
														Elem: &schema.Schema{
															Type:             schema.TypeString,
															ValidateDiagFunc: enum.Validate[awstypes.TargetType](),
														},
													},
													"targets": {
														Type:     schema.TypeSet,
														Required: true,
														Elem:     &schema.Schema{Type: schema.TypeString},
													},
												},
											},
										},
										"rules_string": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"stateful_rule": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrAction: {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.StatefulAction](),
													},
													names.AttrHeader: {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrDestination: {
																	Type:     schema.TypeString,
																	Required: true,
																},
																"destination_port": {
																	Type:     schema.TypeString,
																	Required: true,
																},
																"direction": {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[awstypes.StatefulRuleDirection](),
																},
																names.AttrProtocol: {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[awstypes.StatefulRuleProtocol](),
																},
																names.AttrSource: {
																	Type:     schema.TypeString,
																	Required: true,
																},
																"source_port": {
																	Type:     schema.TypeString,
																	Required: true,
																},
															},
														},
													},
													"rule_option": {
														Type:     schema.TypeSet,
														Required: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"keyword": {
																	Type:     schema.TypeString,
																	Required: true,
																},
																"settings": {
																	Type:     schema.TypeSet,
																	Optional: true,
																	Elem:     &schema.Schema{Type: schema.TypeString},
																},
															},
														},
													},
												},
											},
										},
										"stateless_rules_and_custom_actions": {
											Type:     schema.TypeList,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"custom_action": customActionSchema(),
													"stateless_rule": {
														Type:     schema.TypeSet,
														Required: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrPriority: {
																	Type:     schema.TypeInt,
																	Required: true,
																},
																"rule_definition": {
																	Type:     schema.TypeList,
																	MaxItems: 1,
																	Required: true,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			names.AttrActions: {
																				Type:     schema.TypeSet,
																				Required: true,
																				Elem:     &schema.Schema{Type: schema.TypeString},
																			},
																			"match_attributes": {
																				Type:     schema.TypeList,
																				MaxItems: 1,
																				Required: true,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						names.AttrDestination: {
																							Type:     schema.TypeSet,
																							Optional: true,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"address_definition": {
																										Type:         schema.TypeString,
																										Required:     true,
																										ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
																									},
																								},
																							},
																						},
																						"destination_port": {
																							Type:     schema.TypeSet,
																							Optional: true,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"from_port": {
																										Type:     schema.TypeInt,
																										Required: true,
																									},
																									"to_port": {
																										Type:     schema.TypeInt,
																										Optional: true,
																									},
																								},
																							},
																						},
																						"protocols": {
																							Type:     schema.TypeSet,
																							Optional: true,
																							Elem:     &schema.Schema{Type: schema.TypeInt},
																						},
																						names.AttrSource: {
																							Type:     schema.TypeSet,
																							Optional: true,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"address_definition": {
																										Type:         schema.TypeString,
																										Required:     true,
																										ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
																									},
																								},
																							},
																						},
																						"source_port": {
																							Type:     schema.TypeSet,
																							Optional: true,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"from_port": {
																										Type:     schema.TypeInt,
																										Required: true,
																									},
																									"to_port": {
																										Type:     schema.TypeInt,
																										Optional: true,
																									},
																								},
																							},
																						},
																						"tcp_flag": {
																							Type:     schema.TypeSet,
																							Optional: true,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"flags": {
																										Type:     schema.TypeSet,
																										Required: true,
																										Elem: &schema.Schema{
																											Type:             schema.TypeString,
																											ValidateDiagFunc: enum.Validate[awstypes.TCPFlag](),
																										},
																									},
																									"masks": {
																										Type:     schema.TypeSet,
																										Optional: true,
																										Elem: &schema.Schema{
																											Type:             schema.TypeString,
																											ValidateDiagFunc: enum.Validate[awstypes.TCPFlag](),
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
													},
												},
											},
										},
									},
								},
							},
							"rule_variables": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"ip_sets": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrKey: {
														Type:     schema.TypeString,
														Required: true,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 32),
															validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
															validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_]+$`), "must contain only alphanumeric and underscore characters"),
														),
													},
													"ip_set": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"definition": {
																	Type:     schema.TypeSet,
																	Required: true,
																	Elem:     &schema.Schema{Type: schema.TypeString},
																},
															},
														},
													},
												},
											},
										},
										"port_sets": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrKey: {
														Type:     schema.TypeString,
														Required: true,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 32),
															validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
															validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_]+$`), "must contain only alphanumeric and underscore characters"),
														),
													},
													"port_set": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"definition": {
																	Type:     schema.TypeSet,
																	Required: true,
																	Elem:     &schema.Schema{Type: schema.TypeString},
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
							"stateful_rule_options": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"rule_order": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.RuleOrder](),
										},
									},
								},
							},
						},
					},
				},
				"rules": {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				names.AttrType: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.RuleGroupType](),
				},
				"update_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},

		CustomizeDiff: customdiff.Sequence(
			// The stateful rule_order default action can be explicitly or implicitly set,
			// so ignore spurious diffs if toggling between the two.
			func(_ context.Context, d *schema.ResourceDiff, meta any) error {
				return forceNewIfNotRuleOrderDefault("rule_group.0.stateful_rule_options.0.rule_order", d)
			},
		),
	}
}

func resourceRuleGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &networkfirewall.CreateRuleGroupInput{
		Capacity:      aws.Int32(int32(d.Get("capacity").(int))),
		RuleGroupName: aws.String(name),
		Tags:          getTagsIn(ctx),
		Type:          awstypes.RuleGroupType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEncryptionConfiguration); ok {
		input.EncryptionConfiguration = expandEncryptionConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("rule_group"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.RuleGroup = expandRuleGroup(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("rules"); ok {
		input.Rules = aws.String(v.(string))
	}

	output, err := conn.CreateRuleGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating NetworkFirewall Rule Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.RuleGroupResponse.RuleGroupArn))

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	output, err := findRuleGroupByARN(ctx, conn, d.Id())

	if err == nil && output.RuleGroup == nil {
		err = tfresource.NewEmptyResultError(d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] NetworkFirewall Rule Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Rule Group (%s): %s", d.Id(), err)
	}

	response := output.RuleGroupResponse
	d.Set(names.AttrARN, response.RuleGroupArn)
	d.Set("capacity", response.Capacity)
	d.Set(names.AttrDescription, response.Description)
	if err := d.Set(names.AttrEncryptionConfiguration, flattenEncryptionConfiguration(response.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	d.Set(names.AttrName, response.RuleGroupName)
	if err := d.Set("rule_group", flattenRuleGroup(output.RuleGroup)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule_group: %s", err)
	}
	d.Set(names.AttrType, response.Type)
	d.Set("update_token", output.UpdateToken)

	setTagsOut(ctx, response.Tags)

	return diags
}

func resourceRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	if d.HasChanges(names.AttrDescription, names.AttrEncryptionConfiguration, "rule_group", "rules", names.AttrType) {
		input := &networkfirewall.UpdateRuleGroupInput{
			EncryptionConfiguration: expandEncryptionConfiguration(d.Get(names.AttrEncryptionConfiguration).([]any)),
			RuleGroupArn:            aws.String(d.Id()),
			Type:                    awstypes.RuleGroupType(d.Get(names.AttrType).(string)),
			UpdateToken:             aws.String(d.Get("update_token").(string)),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		// Network Firewall UpdateRuleGroup API method only allows one of Rules or RuleGroup
		// else, request returns "InvalidRequestException: Exactly one of Rules or RuleGroup must be set";
		// Here, "rules" takes precedence as "rule_group" is Computed from "rules" when configured
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19414
		if d.HasChange("rules") {
			input.Rules = aws.String(d.Get("rules").(string))
		} else if d.HasChange("rule_group") {
			if v, ok := d.GetOk("rule_group"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.RuleGroup = expandRuleGroup(v.([]any)[0].(map[string]any))
			}
		}

		// If neither "rules" or "rule_group" are set at this point, neither have changed but
		// at least one must still be sent to allow other attributes (ex. description) to update.
		// Give precedence again to "rules", as documented above.
		if input.Rules == nil && input.RuleGroup == nil {
			if v, ok := d.GetOk("rules"); ok {
				input.Rules = aws.String(v.(string))
			} else if v, ok := d.GetOk("rule_group"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.RuleGroup = expandRuleGroup(v.([]any)[0].(map[string]any))
			}
		}

		_, err := conn.UpdateRuleGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Rule Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	log.Printf("[DEBUG] Deleting NetworkFirewall Rule Group: %s", d.Id())
	const (
		timeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidOperationException](ctx, timeout, func() (any, error) {
		return conn.DeleteRuleGroup(ctx, &networkfirewall.DeleteRuleGroupInput{
			RuleGroupArn: aws.String(d.Id()),
		})
	}, "Unable to delete the object because it is still in use")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting NetworkFirewall Rule Group (%s): %s", d.Id(), err)
	}

	if _, err := waitRuleGroupDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Rule Group (%s) delete : %s", d.Id(), err)
	}

	return diags
}

func findRuleGroupByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeRuleGroupOutput, error) {
	input := &networkfirewall.DescribeRuleGroupInput{
		RuleGroupArn: aws.String(arn),
	}

	output, err := conn.DescribeRuleGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RuleGroupResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusRuleGroup(ctx context.Context, conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findRuleGroupByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.RuleGroup, string(output.RuleGroupResponse.RuleGroupStatus), nil
	}
}

func waitRuleGroupDeleted(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeRuleGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusDeleting),
		Target:  []string{},
		Refresh: statusRuleGroup(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeRuleGroupOutput); ok {
		return output, err
	}

	return nil, err
}

func expandStatefulRuleHeader(tfList []any) *awstypes.Header {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.Header{}

	if v, ok := tfMap[names.AttrDestination].(string); ok && v != "" {
		apiObject.Destination = aws.String(v)
	}
	if v, ok := tfMap["destination_port"].(string); ok && v != "" {
		apiObject.DestinationPort = aws.String(v)
	}
	if v, ok := tfMap["direction"].(string); ok && v != "" {
		apiObject.Direction = awstypes.StatefulRuleDirection(v)
	}
	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = awstypes.StatefulRuleProtocol(v)
	}
	if v, ok := tfMap[names.AttrSource].(string); ok && v != "" {
		apiObject.Source = aws.String(v)
	}
	if v, ok := tfMap["source_port"].(string); ok && v != "" {
		apiObject.SourcePort = aws.String(v)
	}

	return apiObject
}

func expandStatefulRuleOptions(tfList []any) []awstypes.RuleOption {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.RuleOption, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.RuleOption{
			Keyword: aws.String(tfMap["keyword"].(string)),
		}

		if v, ok := tfMap["settings"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.Settings = flex.ExpandStringValueSet(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandRulesSourceList(tfList []any) *awstypes.RulesSourceList {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RulesSourceList{}

	if v, ok := tfMap["generated_rules_type"].(string); ok && v != "" {
		apiObject.GeneratedRulesType = awstypes.GeneratedRulesType(v)
	}
	if v, ok := tfMap["target_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TargetTypes = flex.ExpandStringyValueSet[awstypes.TargetType](v)
	}
	if v, ok := tfMap["targets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Targets = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandStatefulRules(tfList []any) []awstypes.StatefulRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.StatefulRule, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.StatefulRule{}

		if v, ok := tfMap[names.AttrAction].(string); ok && v != "" {
			apiObject.Action = awstypes.StatefulAction(v)
		}
		if v, ok := tfMap[names.AttrHeader].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.Header = expandStatefulRuleHeader(v)
		}
		if v, ok := tfMap["rule_option"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.RuleOptions = expandStatefulRuleOptions(v.List())
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandRuleGroup(tfMap map[string]any) *awstypes.RuleGroup {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RuleGroup{}

	if v, ok := tfMap["reference_sets"].([]any); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]any); ok {
			referenceSets := &awstypes.ReferenceSets{}

			if v, ok := tfMap["ip_set_references"].(*schema.Set); ok && v.Len() > 0 {
				referenceSets.IPSetReferences = expandIPSetReferences(v.List())
			}

			apiObject.ReferenceSets = referenceSets
		}
	}

	if v, ok := tfMap["rule_variables"].([]any); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]any); ok {
			ruleVariables := &awstypes.RuleVariables{}

			if v, ok := tfMap["ip_sets"].(*schema.Set); ok && v.Len() > 0 {
				ruleVariables.IPSets = expandIPSets(v.List())
			}
			if v, ok := tfMap["port_sets"].(*schema.Set); ok && v.Len() > 0 {
				ruleVariables.PortSets = expandPortSets(v.List())
			}

			apiObject.RuleVariables = ruleVariables
		}
	}

	if v, ok := tfMap["rules_source"].([]any); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]any); ok {
			rulesSource := &awstypes.RulesSource{}

			if v, ok := tfMap["rules_source_list"].([]any); ok && len(v) > 0 && v[0] != nil {
				rulesSource.RulesSourceList = expandRulesSourceList(v)
			}
			if v, ok := tfMap["rules_string"].(string); ok && v != "" {
				rulesSource.RulesString = aws.String(v)
			}
			if v, ok := tfMap["stateful_rule"].([]any); ok && len(v) > 0 {
				rulesSource.StatefulRules = expandStatefulRules(v)
			}
			if v, ok := tfMap["stateless_rules_and_custom_actions"].([]any); ok && len(v) > 0 && v[0] != nil {
				rulesSource.StatelessRulesAndCustomActions = expandStatelessRulesAndCustomActions(v)
			}

			apiObject.RulesSource = rulesSource
		}
	}

	if v, ok := tfMap["stateful_rule_options"].([]any); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]any); ok {
			statefulRuleOptions := &awstypes.StatefulRuleOptions{}

			if v, ok := tfMap["rule_order"].(string); ok && v != "" {
				statefulRuleOptions.RuleOrder = awstypes.RuleOrder(v)
			}

			apiObject.StatefulRuleOptions = statefulRuleOptions
		}
	}

	return apiObject
}

func expandIPSetReferences(tfList []any) map[string]awstypes.IPSetReference {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := make(map[string]awstypes.IPSetReference)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		if k, ok := tfMap[names.AttrKey].(string); ok && k != "" {
			if v, ok := tfMap["ip_set_reference"].([]any); ok && len(v) > 0 && v[0] != nil {
				if tfMap, ok := v[0].(map[string]any); ok {
					if v, ok := tfMap["reference_arn"].(string); ok && v != "" {
						apiObject[k] = awstypes.IPSetReference{
							ReferenceArn: aws.String(v),
						}
					}
				}
			}
		}
	}

	return apiObject
}
func expandPortSets(tfList []any) map[string]awstypes.PortSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := make(map[string]awstypes.PortSet)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		if k, ok := tfMap[names.AttrKey].(string); ok && k != "" {
			if v, ok := tfMap["port_set"].([]any); ok && len(v) > 0 && v[0] != nil {
				if tfMap, ok := v[0].(map[string]any); ok {
					if v, ok := tfMap["definition"].(*schema.Set); ok && v.Len() > 0 {
						apiObject[k] = awstypes.PortSet{
							Definition: flex.ExpandStringValueSet(v),
						}
					}
				}
			}
		}
	}

	return apiObject
}

func expandAddresses(tfList []any) []awstypes.Address {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.Address, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.Address{}

		if v, ok := tfMap["address_definition"].(string); ok && v != "" {
			apiObject.AddressDefinition = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandPortRanges(tfList []any) []awstypes.PortRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.PortRange, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.PortRange{}

		if v, ok := tfMap["from_port"].(int); ok {
			apiObject.FromPort = int32(v)
		}
		if v, ok := tfMap["to_port"].(int); ok {
			apiObject.ToPort = int32(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandTCPFlags(tfList []any) []awstypes.TCPFlagField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.TCPFlagField, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.TCPFlagField{}

		if v, ok := tfMap["flags"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.Flags = flex.ExpandStringyValueSet[awstypes.TCPFlag](v)
		}
		if v, ok := tfMap["masks"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.Masks = flex.ExpandStringyValueSet[awstypes.TCPFlag](v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandMatchAttributes(tfList []any) *awstypes.MatchAttributes {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.MatchAttributes{}

	if v, ok := tfMap[names.AttrDestination].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Destinations = expandAddresses(v.List())
	}
	if v, ok := tfMap["destination_port"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.DestinationPorts = expandPortRanges(v.List())
	}
	if v, ok := tfMap["protocols"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Protocols = flex.ExpandInt32ValueSet(v)
	}
	if v, ok := tfMap[names.AttrSource].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Sources = expandAddresses(v.List())
	}
	if v, ok := tfMap["source_port"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SourcePorts = expandPortRanges(v.List())
	}
	if v, ok := tfMap["tcp_flag"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TCPFlags = expandTCPFlags(v.List())
	}

	return apiObject
}

func expandRuleDefinition(tfList []any) *awstypes.RuleDefinition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RuleDefinition{}

	if v, ok := tfMap[names.AttrActions].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Actions = flex.ExpandStringValueSet(v)
	}
	if v, ok := tfMap["match_attributes"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.MatchAttributes = expandMatchAttributes(v)
	}

	return apiObject
}

func expandStatelessRules(tfList []any) []awstypes.StatelessRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.StatelessRule, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.StatelessRule{}

		if v, ok := tfMap[names.AttrPriority].(int); ok && v > 0 {
			apiObject.Priority = aws.Int32(int32(v))
		}
		if v, ok := tfMap["rule_definition"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.RuleDefinition = expandRuleDefinition(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandStatelessRulesAndCustomActions(tfList []any) *awstypes.StatelessRulesAndCustomActions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.StatelessRulesAndCustomActions{}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	if v, ok := tfMap["custom_action"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CustomActions = expandCustomActions(v.List())
	}
	if v, ok := tfMap["stateless_rule"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.StatelessRules = expandStatelessRules(v.List())
	}

	return apiObject
}

func flattenRuleGroup(apiObject *awstypes.RuleGroup) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"reference_sets":        flattenReferenceSets(apiObject.ReferenceSets),
		"rule_variables":        flattenRuleVariables(apiObject.RuleVariables),
		"rules_source":          flattenRulesSource(apiObject.RulesSource),
		"stateful_rule_options": flattenStatefulRulesOptions(apiObject.StatefulRuleOptions),
	}

	return []any{tfMap}
}

func flattenReferenceSets(apiObject *awstypes.ReferenceSets) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"ip_set_references": flattenIPSetReferences(apiObject.IPSetReferences),
	}

	return []any{tfMap}
}

func flattenIPSetReferences(apiObject map[string]awstypes.IPSetReference) []any {
	if apiObject == nil {
		return []any{}
	}

	tfList := make([]any, 0, len(apiObject))

	for k, v := range apiObject {
		tfMap := map[string]any{
			"ip_set_reference": flattenIPSetReference(&v),
			names.AttrKey:      k,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenIPSetReference(apiObject *awstypes.IPSetReference) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"reference_arn": aws.ToString(apiObject.ReferenceArn),
	}

	return []any{tfMap}
}

func flattenRuleVariables(apiObject *awstypes.RuleVariables) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"ip_sets":   flattenIPSets(apiObject.IPSets),
		"port_sets": flattenPortSets(apiObject.PortSets),
	}

	return []any{tfMap}
}

func flattenPortSets(apiObject map[string]awstypes.PortSet) []any {
	if apiObject == nil {
		return []any{}
	}

	tfList := make([]any, 0, len(apiObject))

	for k, v := range apiObject {
		tfMap := map[string]any{
			names.AttrKey: k,
			"port_set":    flattenPortSet(&v),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPortSet(apiObject *awstypes.PortSet) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"definition": apiObject.Definition,
	}

	return []any{tfMap}
}

func flattenRulesSource(apiObject *awstypes.RulesSource) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"rules_source_list":                  flattenRulesSourceList(apiObject.RulesSourceList),
		"rules_string":                       aws.ToString(apiObject.RulesString),
		"stateful_rule":                      flattenStatefulRules(apiObject.StatefulRules),
		"stateless_rules_and_custom_actions": flattenStatelessRulesAndCustomActions(apiObject.StatelessRulesAndCustomActions),
	}

	return []any{tfMap}
}

func flattenRulesSourceList(apiObject *awstypes.RulesSourceList) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"generated_rules_type": apiObject.GeneratedRulesType,
		"target_types":         apiObject.TargetTypes,
		"targets":              apiObject.Targets,
	}

	return []any{tfMap}
}

func flattenStatefulRules(apiObjects []awstypes.StatefulRule) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		m := map[string]any{
			names.AttrAction: apiObject.Action,
			names.AttrHeader: flattenHeader(apiObject.Header),
			"rule_option":    flattenRuleOptions(apiObject.RuleOptions),
		}

		tfList = append(tfList, m)
	}

	return tfList
}

func flattenHeader(apiObject *awstypes.Header) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrDestination: aws.ToString(apiObject.Destination),
		"destination_port":    aws.ToString(apiObject.DestinationPort),
		"direction":           apiObject.Direction,
		names.AttrProtocol:    apiObject.Protocol,
		names.AttrSource:      aws.ToString(apiObject.Source),
		"source_port":         aws.ToString(apiObject.SourcePort),
	}

	return []any{tfMap}
}

func flattenRuleOptions(apiObjects []awstypes.RuleOption) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"keyword":  aws.ToString(apiObject.Keyword),
			"settings": apiObject.Settings,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenStatelessRulesAndCustomActions(apiObject *awstypes.StatelessRulesAndCustomActions) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"custom_action":  flattenCustomActions(apiObject.CustomActions),
		"stateless_rule": flattenStatelessRules(apiObject.StatelessRules),
	}

	return []any{tfMap}
}

func flattenStatelessRules(apiObjects []awstypes.StatelessRule) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrPriority: aws.ToInt32(apiObject.Priority),
			"rule_definition":  flattenRuleDefinition(apiObject.RuleDefinition),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRuleDefinition(apiObject *awstypes.RuleDefinition) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrActions:  apiObject.Actions,
		"match_attributes": flattenMatchAttributes(apiObject.MatchAttributes),
	}

	return []any{tfMap}
}

func flattenMatchAttributes(apiObject *awstypes.MatchAttributes) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrDestination: flattenAddresses(apiObject.Destinations),
		"destination_port":    flattenPortRanges(apiObject.DestinationPorts),
		"protocols":           apiObject.Protocols,
		names.AttrSource:      flattenAddresses(apiObject.Sources),
		"source_port":         flattenPortRanges(apiObject.SourcePorts),
		"tcp_flag":            flattenTCPFlags(apiObject.TCPFlags),
	}

	return []any{tfMap}
}

func flattenAddresses(apiObjects []awstypes.Address) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"address_definition": aws.ToString(apiObject.AddressDefinition),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPortRanges(apiObjects []awstypes.PortRange) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"from_port": apiObject.FromPort,
			"to_port":   apiObject.ToPort,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTCPFlags(apiObjects []awstypes.TCPFlagField) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		m := map[string]any{
			"flags": apiObject.Flags,
			"masks": apiObject.Masks,
		}

		tfList = append(tfList, m)
	}

	return tfList
}

func flattenStatefulRulesOptions(apiObject *awstypes.StatefulRuleOptions) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"rule_order": apiObject.RuleOrder,
	}

	return []any{tfMap}
}
