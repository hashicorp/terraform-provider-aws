// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkfirewall_firewall_policy", name="Firewall Policy")
// @Tags(identifierAttribute="id")
func ResourceFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallPolicyCreate,
		ReadWithoutTimeout:   resourceFirewallPolicyRead,
		UpdateWithoutTimeout: resourceFirewallPolicyUpdate,
		DeleteWithoutTimeout: resourceFirewallPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encryption_configuration": encryptionConfigurationSchema(),
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
												"key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 32),
														validation.StringMatch(regexp.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
														validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_]+$`), "must contain only alphanumeric and underscore characters"),
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(networkfirewall.RuleOrder_Values(), false),
									},
									"stream_exception_policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(networkfirewall.StreamExceptionPolicy_Values(), false),
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
												"action": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(networkfirewall.OverrideAction_Values(), false),
												},
											},
										},
									},
									"priority": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"resource_arn": {
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
									"priority": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"resource_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
			"name": {
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
	conn := meta.(*conns.AWSClient).NetworkFirewallConn(ctx)

	name := d.Get("name").(string)
	input := &networkfirewall.CreateFirewallPolicyInput{
		FirewallPolicy:     expandFirewallPolicy(d.Get("firewall_policy").([]interface{})),
		FirewallPolicyName: aws.String(d.Get("name").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("encryption_configuration"); ok {
		input.EncryptionConfiguration = expandEncryptionConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateFirewallPolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating NetworkFirewall Firewall Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.FirewallPolicyResponse.FirewallPolicyArn))

	return resourceFirewallPolicyRead(ctx, d, meta)
}

func resourceFirewallPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn(ctx)

	output, err := FindFirewallPolicyByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] NetworkFirewall Firewall Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading NetworkFirewall Firewall Policy (%s): %s", d.Id(), err)
	}

	response := output.FirewallPolicyResponse
	d.Set("arn", response.FirewallPolicyArn)
	d.Set("description", response.Description)
	d.Set("encryption_configuration", flattenEncryptionConfiguration(response.EncryptionConfiguration))
	if err := d.Set("firewall_policy", flattenFirewallPolicy(output.FirewallPolicy)); err != nil {
		return diag.Errorf("setting firewall_policy: %s", err)
	}
	d.Set("name", response.FirewallPolicyName)
	d.Set("update_token", output.UpdateToken)

	setTagsOut(ctx, response.Tags)

	return nil
}

func resourceFirewallPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn(ctx)

	if d.HasChanges("description", "encryption_configuration", "firewall_policy") {
		input := &networkfirewall.UpdateFirewallPolicyInput{
			EncryptionConfiguration: expandEncryptionConfiguration(d.Get("encryption_configuration").([]interface{})),
			FirewallPolicy:          expandFirewallPolicy(d.Get("firewall_policy").([]interface{})),
			FirewallPolicyArn:       aws.String(d.Id()),
			UpdateToken:             aws.String(d.Get("update_token").(string)),
		}

		// Only pass non-empty description values, else API request returns an InternalServiceError
		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateFirewallPolicyWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating NetworkFirewall Firewall Policy (%s): %s", d.Id(), err)
		}
	}

	return resourceFirewallPolicyRead(ctx, d, meta)
}

func resourceFirewallPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const (
		timeout = 10 * time.Minute
	)
	conn := meta.(*conns.AWSClient).NetworkFirewallConn(ctx)

	log.Printf("[DEBUG] Deleting NetworkFirewall Firewall Policy: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, timeout, func() (interface{}, error) {
		return conn.DeleteFirewallPolicyWithContext(ctx, &networkfirewall.DeleteFirewallPolicyInput{
			FirewallPolicyArn: aws.String(d.Id()),
		})
	}, networkfirewall.ErrCodeInvalidOperationException, "Unable to delete the object because it is still in use")

	if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting NetworkFirewall Firewall Policy (%s): %s", d.Id(), err)
	}

	if _, err := waitFirewallPolicyDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return diag.Errorf("waiting for NetworkFirewall Firewall Policy (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindFirewallPolicyByARN(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeFirewallPolicyOutput, error) {
	input := &networkfirewall.DescribeFirewallPolicyInput{
		FirewallPolicyArn: aws.String(arn),
	}

	output, err := conn.DescribeFirewallPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
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

func statusFirewallPolicy(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFirewallPolicyByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.FirewallPolicyResponse.FirewallPolicyStatus), nil
	}
}

func waitFirewallPolicyDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string, timeout time.Duration) (*networkfirewall.DescribeFirewallPolicyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkfirewall.ResourceStatusDeleting},
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

func expandPolicyVariables(tfMap map[string]interface{}) *networkfirewall.PolicyVariables {
	if tfMap == nil {
		return nil
	}

	policyVariables := &networkfirewall.PolicyVariables{}

	if rvMap, ok := tfMap["rule_variables"].(*schema.Set); ok && rvMap.Len() > 0 {
		policyVariables.RuleVariables = expandIPSets(rvMap.List())
	}

	return policyVariables
}

func expandStatefulEngineOptions(l []interface{}) *networkfirewall.StatefulEngineOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	options := &networkfirewall.StatefulEngineOptions{}

	m := l[0].(map[string]interface{})
	if v, ok := m["rule_order"].(string); ok && v != "" {
		options.RuleOrder = aws.String(v)
	}
	if v, ok := m["stream_exception_policy"].(string); ok && v != "" {
		options.StreamExceptionPolicy = aws.String(v)
	}

	return options
}

func expandStatefulRuleGroupOverride(l []interface{}) *networkfirewall.StatefulRuleGroupOverride {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	lRaw := l[0].(map[string]interface{})
	override := &networkfirewall.StatefulRuleGroupOverride{}

	if v, ok := lRaw["action"].(string); ok && v != "" {
		override.SetAction(v)
	}

	return override
}

func expandStatefulRuleGroupReferences(l []interface{}) []*networkfirewall.StatefulRuleGroupReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	references := make([]*networkfirewall.StatefulRuleGroupReference, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		reference := &networkfirewall.StatefulRuleGroupReference{}
		if v, ok := tfMap["priority"].(int); ok && v > 0 {
			reference.Priority = aws.Int64(int64(v))
		}
		if v, ok := tfMap["resource_arn"].(string); ok && v != "" {
			reference.ResourceArn = aws.String(v)
		}

		if v, ok := tfMap["override"].([]interface{}); ok && len(v) > 0 {
			reference.Override = expandStatefulRuleGroupOverride(v)
		}

		references = append(references, reference)
	}

	return references
}

func expandStatelessRuleGroupReferences(l []interface{}) []*networkfirewall.StatelessRuleGroupReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	references := make([]*networkfirewall.StatelessRuleGroupReference, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		reference := &networkfirewall.StatelessRuleGroupReference{}
		if v, ok := tfMap["priority"].(int); ok && v > 0 {
			reference.Priority = aws.Int64(int64(v))
		}
		if v, ok := tfMap["resource_arn"].(string); ok && v != "" {
			reference.ResourceArn = aws.String(v)
		}
		references = append(references, reference)
	}
	return references
}

func expandFirewallPolicy(l []interface{}) *networkfirewall.FirewallPolicy {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	lRaw := l[0].(map[string]interface{})
	policy := &networkfirewall.FirewallPolicy{
		StatelessDefaultActions:         flex.ExpandStringSet(lRaw["stateless_default_actions"].(*schema.Set)),
		StatelessFragmentDefaultActions: flex.ExpandStringSet(lRaw["stateless_fragment_default_actions"].(*schema.Set)),
	}

	if v, ok := lRaw["policy_variables"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		policy.PolicyVariables = expandPolicyVariables(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := lRaw["stateful_default_actions"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatefulDefaultActions = flex.ExpandStringSet(v)
	}

	if v, ok := lRaw["stateful_engine_options"].([]interface{}); ok && len(v) > 0 {
		policy.StatefulEngineOptions = expandStatefulEngineOptions(v)
	}

	if v, ok := lRaw["stateful_rule_group_reference"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatefulRuleGroupReferences = expandStatefulRuleGroupReferences(v.List())
	}

	if v, ok := lRaw["stateless_custom_action"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatelessCustomActions = expandCustomActions(v.List())
	}

	if v, ok := lRaw["stateless_rule_group_reference"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatelessRuleGroupReferences = expandStatelessRuleGroupReferences(v.List())
	}

	return policy
}

func flattenFirewallPolicy(policy *networkfirewall.FirewallPolicy) []interface{} {
	if policy == nil {
		return []interface{}{}
	}
	p := map[string]interface{}{}
	if policy.PolicyVariables != nil {
		p["policy_variables"] = flattenPolicyVariables(policy.PolicyVariables)
	}
	if policy.StatefulDefaultActions != nil {
		p["stateful_default_actions"] = flex.FlattenStringSet(policy.StatefulDefaultActions)
	}
	if policy.StatefulEngineOptions != nil {
		p["stateful_engine_options"] = flattenStatefulEngineOptions(policy.StatefulEngineOptions)
	}
	if policy.StatefulRuleGroupReferences != nil {
		p["stateful_rule_group_reference"] = flattenPolicyStatefulRuleGroupReference(policy.StatefulRuleGroupReferences)
	}
	if policy.StatelessCustomActions != nil {
		p["stateless_custom_action"] = flattenCustomActions(policy.StatelessCustomActions)
	}
	if policy.StatelessDefaultActions != nil {
		p["stateless_default_actions"] = flex.FlattenStringSet(policy.StatelessDefaultActions)
	}
	if policy.StatelessFragmentDefaultActions != nil {
		p["stateless_fragment_default_actions"] = flex.FlattenStringSet(policy.StatelessFragmentDefaultActions)
	}
	if policy.StatelessRuleGroupReferences != nil {
		p["stateless_rule_group_reference"] = flattenPolicyStatelessRuleGroupReference(policy.StatelessRuleGroupReferences)
	}

	return []interface{}{p}
}

func flattenPolicyVariables(variables *networkfirewall.PolicyVariables) []interface{} {
	if variables == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"rule_variables": flattenIPSets(variables.RuleVariables),
	}

	return []interface{}{m}
}

func flattenStatefulEngineOptions(options *networkfirewall.StatefulEngineOptions) []interface{} {
	if options == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}
	if options.RuleOrder != nil {
		m["rule_order"] = aws.StringValue(options.RuleOrder)
	}
	if options.StreamExceptionPolicy != nil {
		m["stream_exception_policy"] = aws.StringValue(options.StreamExceptionPolicy)
	}

	return []interface{}{m}
}

func flattenStatefulRuleGroupOverride(override *networkfirewall.StatefulRuleGroupOverride) []interface{} {
	if override == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"action": aws.StringValue(override.Action),
	}

	return []interface{}{m}
}

func flattenPolicyStatefulRuleGroupReference(l []*networkfirewall.StatefulRuleGroupReference) []interface{} {
	references := make([]interface{}, 0, len(l))
	for _, ref := range l {
		reference := map[string]interface{}{
			"resource_arn": aws.StringValue(ref.ResourceArn),
		}
		if ref.Priority != nil {
			reference["priority"] = int(aws.Int64Value(ref.Priority))
		}
		if ref.Override != nil {
			reference["override"] = flattenStatefulRuleGroupOverride(ref.Override)
		}

		references = append(references, reference)
	}

	return references
}

func flattenPolicyStatelessRuleGroupReference(l []*networkfirewall.StatelessRuleGroupReference) []interface{} {
	references := make([]interface{}, 0, len(l))
	for _, ref := range l {
		reference := map[string]interface{}{
			"priority":     int(aws.Int64Value(ref.Priority)),
			"resource_arn": aws.StringValue(ref.ResourceArn),
		}
		references = append(references, reference)
	}
	return references
}
