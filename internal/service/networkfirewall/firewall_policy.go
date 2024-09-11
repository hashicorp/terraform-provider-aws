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

// @SDKResource("aws_networkfirewall_firewall_policy", name="Firewall Policy")
// @Tags(identifierAttribute="id")
func resourceFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallPolicyCreate,
		ReadWithoutTimeout:   resourceFirewallPolicyRead,
		UpdateWithoutTimeout: resourceFirewallPolicyUpdate,
		DeleteWithoutTimeout: resourceFirewallPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrEncryptionConfiguration: encryptionConfigurationSchema(),
				"firewall_policy": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"policy_variables": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"rule_variables": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
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
							"stateful_default_actions": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"stateful_engine_options": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"rule_order": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[awstypes.RuleOrder](),
										},
										"stream_exception_policy": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[awstypes.StreamExceptionPolicy](),
										},
									},
								},
							},
							"stateful_rule_group_reference": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"override": {
											Type:     schema.TypeList,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrAction: {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[awstypes.OverrideAction](),
													},
												},
											},
										},
										names.AttrPriority: {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
										names.AttrResourceARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
							"stateless_custom_action": customActionSchema(),
							"stateless_default_actions": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"stateless_fragment_default_actions": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"stateless_rule_group_reference": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrPriority: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
										names.AttrResourceARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
							"tls_inspection_configuration_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"update_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},

		CustomizeDiff: customdiff.Sequence(
			// The stateful rule_order default action can be explicitly or implicitly set,
			// so ignore spurious diffs if toggling between the two.
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				return forceNewIfNotRuleOrderDefault("firewall_policy.0.stateful_engine_options.0.rule_order", d)
			},
			verify.SetTagsDiff,
		),
	}
}

func resourceFirewallPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &networkfirewall.CreateFirewallPolicyInput{
		FirewallPolicy:     expandFirewallPolicy(d.Get("firewall_policy").([]interface{})),
		FirewallPolicyName: aws.String(d.Get(names.AttrName).(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEncryptionConfiguration); ok {
		input.EncryptionConfiguration = expandEncryptionConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateFirewallPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating NetworkFirewall Firewall Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.FirewallPolicyResponse.FirewallPolicyArn))

	return append(diags, resourceFirewallPolicyRead(ctx, d, meta)...)
}

func resourceFirewallPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	output, err := findFirewallPolicyByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] NetworkFirewall Firewall Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall Policy (%s): %s", d.Id(), err)
	}

	response := output.FirewallPolicyResponse
	d.Set(names.AttrARN, response.FirewallPolicyArn)
	d.Set(names.AttrDescription, response.Description)
	if err := d.Set(names.AttrEncryptionConfiguration, flattenEncryptionConfiguration(response.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	if err := d.Set("firewall_policy", flattenFirewallPolicy(output.FirewallPolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting firewall_policy: %s", err)
	}
	d.Set(names.AttrName, response.FirewallPolicyName)
	d.Set("update_token", output.UpdateToken)

	setTagsOut(ctx, response.Tags)

	return diags
}

func resourceFirewallPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	if d.HasChanges(names.AttrDescription, names.AttrEncryptionConfiguration, "firewall_policy") {
		input := &networkfirewall.UpdateFirewallPolicyInput{
			EncryptionConfiguration: expandEncryptionConfiguration(d.Get(names.AttrEncryptionConfiguration).([]interface{})),
			FirewallPolicy:          expandFirewallPolicy(d.Get("firewall_policy").([]interface{})),
			FirewallPolicyArn:       aws.String(d.Id()),
			UpdateToken:             aws.String(d.Get("update_token").(string)),
		}

		// Only pass non-empty description values, else API request returns an InternalServiceError.
		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateFirewallPolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceFirewallPolicyRead(ctx, d, meta)...)
}

func resourceFirewallPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	log.Printf("[DEBUG] Deleting NetworkFirewall Firewall Policy: %s", d.Id())
	const (
		timeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidOperationException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteFirewallPolicy(ctx, &networkfirewall.DeleteFirewallPolicyInput{
			FirewallPolicyArn: aws.String(d.Id()),
		})
	}, "Unable to delete the object because it is still in use")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting NetworkFirewall Firewall Policy (%s): %s", d.Id(), err)
	}

	if _, err := waitFirewallPolicyDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Firewall Policy (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFirewallPolicy(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.DescribeFirewallPolicyInput) (*networkfirewall.DescribeFirewallPolicyOutput, error) {
	output, err := conn.DescribeFirewallPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FirewallPolicyResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findFirewallPolicyByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeFirewallPolicyOutput, error) {
	input := &networkfirewall.DescribeFirewallPolicyInput{
		FirewallPolicyArn: aws.String(arn),
	}

	return findFirewallPolicy(ctx, conn, input)
}

func statusFirewallPolicy(ctx context.Context, conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFirewallPolicyByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FirewallPolicyResponse.FirewallPolicyStatus), nil
	}
}

func waitFirewallPolicyDeleted(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeFirewallPolicyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusDeleting),
		Target:  []string{},
		Refresh: statusFirewallPolicy(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeFirewallPolicyOutput); ok {
		return output, err
	}

	return nil, err
}

func expandPolicyVariables(tfMap map[string]interface{}) *awstypes.PolicyVariables {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PolicyVariables{}

	if v, ok := tfMap["rule_variables"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.RuleVariables = expandIPSets(v.List())
	}

	return apiObject
}

func expandStatefulEngineOptions(tfList []interface{}) *awstypes.StatefulEngineOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.StatefulEngineOptions{}

	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["rule_order"].(string); ok && v != "" {
		apiObject.RuleOrder = awstypes.RuleOrder(v)
	}
	if v, ok := tfMap["stream_exception_policy"].(string); ok && v != "" {
		apiObject.StreamExceptionPolicy = awstypes.StreamExceptionPolicy(v)
	}

	return apiObject
}

func expandStatefulRuleGroupOverride(tfList []interface{}) *awstypes.StatefulRuleGroupOverride {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.StatefulRuleGroupOverride{}

	if v, ok := tfMap[names.AttrAction].(string); ok && v != "" {
		apiObject.Action = awstypes.OverrideAction(v)
	}

	return apiObject
}

func expandStatefulRuleGroupReferences(tfList []interface{}) []awstypes.StatefulRuleGroupReference {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.StatefulRuleGroupReference, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.StatefulRuleGroupReference{}

		if v, ok := tfMap["override"].([]interface{}); ok && len(v) > 0 {
			apiObject.Override = expandStatefulRuleGroupOverride(v)
		}
		if v, ok := tfMap[names.AttrPriority].(int); ok && v > 0 {
			apiObject.Priority = aws.Int32(int32(v))
		}
		if v, ok := tfMap[names.AttrResourceARN].(string); ok && v != "" {
			apiObject.ResourceArn = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandStatelessRuleGroupReferences(tfList []interface{}) []awstypes.StatelessRuleGroupReference {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.StatelessRuleGroupReference, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.StatelessRuleGroupReference{}

		if v, ok := tfMap[names.AttrPriority].(int); ok && v > 0 {
			apiObject.Priority = aws.Int32(int32(v))
		}
		if v, ok := tfMap[names.AttrResourceARN].(string); ok && v != "" {
			apiObject.ResourceArn = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandFirewallPolicy(tfList []interface{}) *awstypes.FirewallPolicy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.FirewallPolicy{
		StatelessDefaultActions:         flex.ExpandStringValueSet(tfMap["stateless_default_actions"].(*schema.Set)),
		StatelessFragmentDefaultActions: flex.ExpandStringValueSet(tfMap["stateless_fragment_default_actions"].(*schema.Set)),
	}

	if v, ok := tfMap["policy_variables"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.PolicyVariables = expandPolicyVariables(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["stateful_default_actions"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.StatefulDefaultActions = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["stateful_engine_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.StatefulEngineOptions = expandStatefulEngineOptions(v)
	}

	if v, ok := tfMap["stateful_rule_group_reference"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.StatefulRuleGroupReferences = expandStatefulRuleGroupReferences(v.List())
	}

	if v, ok := tfMap["stateless_custom_action"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.StatelessCustomActions = expandCustomActions(v.List())
	}

	if v, ok := tfMap["stateless_rule_group_reference"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.StatelessRuleGroupReferences = expandStatelessRuleGroupReferences(v.List())
	}

	if v, ok := tfMap["tls_inspection_configuration_arn"].(string); ok && v != "" {
		apiObject.TLSInspectionConfigurationArn = aws.String(v)
	}

	return apiObject
}

func flattenFirewallPolicy(apiObject *awstypes.FirewallPolicy) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if apiObject.PolicyVariables != nil {
		tfMap["policy_variables"] = flattenPolicyVariables(apiObject.PolicyVariables)
	}
	if apiObject.StatefulDefaultActions != nil {
		tfMap["stateful_default_actions"] = apiObject.StatefulDefaultActions
	}
	if apiObject.StatefulEngineOptions != nil {
		tfMap["stateful_engine_options"] = flattenStatefulEngineOptions(apiObject.StatefulEngineOptions)
	}
	if apiObject.StatefulRuleGroupReferences != nil {
		tfMap["stateful_rule_group_reference"] = flattenPolicyStatefulRuleGroupReferences(apiObject.StatefulRuleGroupReferences)
	}
	if apiObject.StatelessCustomActions != nil {
		tfMap["stateless_custom_action"] = flattenCustomActions(apiObject.StatelessCustomActions)
	}
	if apiObject.StatelessDefaultActions != nil {
		tfMap["stateless_default_actions"] = apiObject.StatelessDefaultActions
	}
	if apiObject.StatelessFragmentDefaultActions != nil {
		tfMap["stateless_fragment_default_actions"] = apiObject.StatelessFragmentDefaultActions
	}
	if apiObject.StatelessRuleGroupReferences != nil {
		tfMap["stateless_rule_group_reference"] = flattenPolicyStatelessRuleGroupReferences(apiObject.StatelessRuleGroupReferences)
	}
	if apiObject.TLSInspectionConfigurationArn != nil {
		tfMap["tls_inspection_configuration_arn"] = aws.ToString(apiObject.TLSInspectionConfigurationArn)
	}

	return []interface{}{tfMap}
}

func flattenPolicyVariables(apiObject *awstypes.PolicyVariables) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"rule_variables": flattenIPSets(apiObject.RuleVariables),
	}

	return []interface{}{tfMap}
}

func flattenStatefulEngineOptions(apiObject *awstypes.StatefulEngineOptions) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"rule_order":              apiObject.RuleOrder,
		"stream_exception_policy": apiObject.StreamExceptionPolicy,
	}

	return []interface{}{tfMap}
}

func flattenStatefulRuleGroupOverride(apiObject *awstypes.StatefulRuleGroupOverride) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrAction: apiObject.Action,
	}

	return []interface{}{tfMap}
}

func flattenPolicyStatefulRuleGroupReferences(apiObjects []awstypes.StatefulRuleGroupReference) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrResourceARN: aws.ToString(apiObject.ResourceArn),
		}

		if apiObject.Override != nil {
			tfMap["override"] = flattenStatefulRuleGroupOverride(apiObject.Override)
		}
		if apiObject.Priority != nil {
			tfMap[names.AttrPriority] = aws.ToInt32(apiObject.Priority)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPolicyStatelessRuleGroupReferences(apiObjects []awstypes.StatelessRuleGroupReference) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrPriority:    aws.ToInt32(apiObject.Priority),
			names.AttrResourceARN: aws.ToString(apiObject.ResourceArn),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
