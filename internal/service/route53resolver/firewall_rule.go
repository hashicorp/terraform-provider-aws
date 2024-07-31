// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_firewall_rule")
func ResourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallRuleCreate,
		ReadWithoutTimeout:   resourceFirewallRuleRead,
		UpdateWithoutTimeout: resourceFirewallRuleUpdate,
		DeleteWithoutTimeout: resourceFirewallRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAction: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.Action_Values(), false),
			},
			"block_override_dns_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.BlockOverrideDnsType_Values(), false),
			},
			"block_override_domain": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"block_override_ttl": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 604800),
			},
			"block_response": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.BlockResponse_Values(), false),
			},
			"firewall_domain_list_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"firewall_domain_redirection_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      route53resolver.FirewallDomainRedirectionActionInspectRedirectionDomain,
				ValidateFunc: validation.StringInSlice(route53resolver.FirewallDomainRedirectionAction_Values(), false),
			},
			"firewall_rule_group_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validResolverName,
			},
			names.AttrPriority: {
				Type:     schema.TypeInt,
				Required: true,
			},
			"q_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceFirewallRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	firewallDomainListID := d.Get("firewall_domain_list_id").(string)
	firewallRuleGroupID := d.Get("firewall_rule_group_id").(string)
	ruleID := FirewallRuleCreateResourceID(firewallRuleGroupID, firewallDomainListID)
	name := d.Get(names.AttrName).(string)
	input := &route53resolver.CreateFirewallRuleInput{
		Action:                          aws.String(d.Get(names.AttrAction).(string)),
		CreatorRequestId:                aws.String(id.PrefixedUniqueId("tf-r53-resolver-firewall-rule-")),
		FirewallRuleGroupId:             aws.String(firewallRuleGroupID),
		FirewallDomainListId:            aws.String(firewallDomainListID),
		FirewallDomainRedirectionAction: aws.String(d.Get("firewall_domain_redirection_action").(string)),
		Name:                            aws.String(name),
		Priority:                        aws.Int64(int64(d.Get(names.AttrPriority).(int))),
	}

	if v, ok := d.GetOk("block_override_dns_type"); ok {
		input.BlockOverrideDnsType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("block_override_domain"); ok {
		input.BlockOverrideDomain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("block_override_ttl"); ok {
		input.BlockOverrideTtl = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("block_response"); ok {
		input.BlockResponse = aws.String(v.(string))
	}

	if v, ok := d.GetOk("q_type"); ok {
		input.Qtype = aws.String(v.(string))
	}

	_, err := conn.CreateFirewallRuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Rule (%s): %s", name, err)
	}

	d.SetId(ruleID)

	return append(diags, resourceFirewallRuleRead(ctx, d, meta)...)
}

func resourceFirewallRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	firewallRuleGroupID, firewallDomainListID, err := FirewallRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	firewallRule, err := FindFirewallRuleByTwoPartKey(ctx, conn, firewallRuleGroupID, firewallDomainListID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Rule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAction, firewallRule.Action)
	d.Set("block_override_dns_type", firewallRule.BlockOverrideDnsType)
	d.Set("block_override_domain", firewallRule.BlockOverrideDomain)
	d.Set("block_override_ttl", firewallRule.BlockOverrideTtl)
	d.Set("block_response", firewallRule.BlockResponse)
	d.Set("firewall_rule_group_id", firewallRule.FirewallRuleGroupId)
	d.Set("firewall_domain_list_id", firewallRule.FirewallDomainListId)
	d.Set("firewall_domain_redirection_action", firewallRule.FirewallDomainRedirectionAction)
	d.Set(names.AttrName, firewallRule.Name)
	d.Set(names.AttrPriority, firewallRule.Priority)
	d.Set("q_type", firewallRule.Qtype)

	return diags
}

func resourceFirewallRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	firewallRuleGroupID, firewallDomainListID, err := FirewallRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &route53resolver.UpdateFirewallRuleInput{
		Action:               aws.String(d.Get(names.AttrAction).(string)),
		FirewallDomainListId: aws.String(firewallDomainListID),
		FirewallRuleGroupId:  aws.String(firewallRuleGroupID),
		Name:                 aws.String(d.Get(names.AttrName).(string)),
		Priority:             aws.Int64(int64(d.Get(names.AttrPriority).(int))),
	}

	if v, ok := d.GetOk("block_override_dns_type"); ok {
		input.BlockOverrideDnsType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("block_override_domain"); ok {
		input.BlockOverrideDomain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("block_override_ttl"); ok {
		input.BlockOverrideTtl = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("block_response"); ok {
		input.BlockResponse = aws.String(v.(string))
	}

	if v, ok := d.GetOk("firewall_domain_redirection_action"); ok {
		input.FirewallDomainRedirectionAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk("q_type"); ok {
		input.Qtype = aws.String(v.(string))
	}

	_, err = conn.UpdateFirewallRuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Firewall Rule (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFirewallRuleRead(ctx, d, meta)...)
}

func resourceFirewallRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	firewallRuleGroupID, firewallDomainListID, err := FirewallRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &route53resolver.DeleteFirewallRuleInput{
		FirewallDomainListId: aws.String(firewallDomainListID),
		FirewallRuleGroupId:  aws.String(firewallRuleGroupID),
	}

	if v, ok := d.GetOk("q_type"); ok {
		input.Qtype = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Rule: %s", d.Id())
	_, err = conn.DeleteFirewallRuleWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Firewall Rule (%s): %s", d.Id(), err)
	}

	return diags
}

const firewallRuleIDSeparator = ":"

func FirewallRuleCreateResourceID(firewallRuleGroupID, firewallDomainListID string) string {
	parts := []string{firewallRuleGroupID, firewallDomainListID}
	id := strings.Join(parts, firewallRuleIDSeparator)

	return id
}

func FirewallRuleParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, firewallRuleIDSeparator, 2)

	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected firewall_rule_group_id%[2]sfirewall_domain_list_id", id, firewallRuleIDSeparator)
	}

	return parts[0], parts[1], nil
}

func FindFirewallRuleByTwoPartKey(ctx context.Context, conn *route53resolver.Route53Resolver, firewallRuleGroupID, firewallDomainListID string) (*route53resolver.FirewallRule, error) {
	output, err := findFirewallRules(ctx, conn, firewallRuleGroupID, func(rule *route53resolver.FirewallRule) bool {
		return aws.StringValue(rule.FirewallDomainListId) == firewallDomainListID
	})

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(firewallRuleGroupID)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, firewallRuleGroupID)
	}

	return output[0], nil
}

func findFirewallRules(ctx context.Context, conn *route53resolver.Route53Resolver, firewallRuleGroupID string, f func(*route53resolver.FirewallRule) bool) ([]*route53resolver.FirewallRule, error) {
	input := &route53resolver.ListFirewallRulesInput{
		FirewallRuleGroupId: aws.String(firewallRuleGroupID),
	}
	var output []*route53resolver.FirewallRule

	err := conn.ListFirewallRulesPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FirewallRules {
			if f(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
