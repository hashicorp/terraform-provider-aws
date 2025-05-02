// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_firewall_rule", name="Firewall Rule")
func resourceFirewallRule() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Action](),
			},
			"block_override_dns_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.BlockOverrideDnsType](),
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.BlockResponse](),
			},
			"firewall_domain_list_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"firewall_domain_redirection_action": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.FirewallDomainRedirectionActionInspectRedirectionDomain,
				ValidateDiagFunc: enum.Validate[awstypes.FirewallDomainRedirectionAction](),
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

func resourceFirewallRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	firewallDomainListID := d.Get("firewall_domain_list_id").(string)
	firewallRuleGroupID := d.Get("firewall_rule_group_id").(string)
	ruleID := firewallRuleCreateResourceID(firewallRuleGroupID, firewallDomainListID)
	name := d.Get(names.AttrName).(string)
	input := &route53resolver.CreateFirewallRuleInput{
		Action:                          awstypes.Action(d.Get(names.AttrAction).(string)),
		CreatorRequestId:                aws.String(id.PrefixedUniqueId("tf-r53-resolver-firewall-rule-")),
		FirewallRuleGroupId:             aws.String(firewallRuleGroupID),
		FirewallDomainListId:            aws.String(firewallDomainListID),
		FirewallDomainRedirectionAction: awstypes.FirewallDomainRedirectionAction(d.Get("firewall_domain_redirection_action").(string)),
		Name:                            aws.String(name),
		Priority:                        aws.Int32(int32(d.Get(names.AttrPriority).(int))),
	}

	if v, ok := d.GetOk("block_override_dns_type"); ok {
		input.BlockOverrideDnsType = awstypes.BlockOverrideDnsType(v.(string))
	}

	if v, ok := d.GetOk("block_override_domain"); ok {
		input.BlockOverrideDomain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("block_override_ttl"); ok {
		input.BlockOverrideTtl = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("block_response"); ok {
		input.BlockResponse = awstypes.BlockResponse(v.(string))
	}

	if v, ok := d.GetOk("q_type"); ok {
		input.Qtype = aws.String(v.(string))
	}

	_, err := conn.CreateFirewallRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Rule (%s): %s", name, err)
	}

	d.SetId(ruleID)

	return append(diags, resourceFirewallRuleRead(ctx, d, meta)...)
}

func resourceFirewallRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	firewallRuleGroupID, firewallDomainListID, err := firewallRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	firewallRule, err := findFirewallRuleByTwoPartKey(ctx, conn, firewallRuleGroupID, firewallDomainListID)

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

func resourceFirewallRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	firewallRuleGroupID, firewallDomainListID, err := firewallRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &route53resolver.UpdateFirewallRuleInput{
		Action:               awstypes.Action(d.Get(names.AttrAction).(string)),
		FirewallDomainListId: aws.String(firewallDomainListID),
		FirewallRuleGroupId:  aws.String(firewallRuleGroupID),
		Name:                 aws.String(d.Get(names.AttrName).(string)),
		Priority:             aws.Int32(int32(d.Get(names.AttrPriority).(int))),
	}

	if v, ok := d.GetOk("block_override_dns_type"); ok {
		input.BlockOverrideDnsType = awstypes.BlockOverrideDnsType(v.(string))
	}

	if v, ok := d.GetOk("block_override_domain"); ok {
		input.BlockOverrideDomain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("block_override_ttl"); ok {
		input.BlockOverrideTtl = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("block_response"); ok {
		input.BlockResponse = awstypes.BlockResponse(v.(string))
	}

	if v, ok := d.GetOk("firewall_domain_redirection_action"); ok {
		input.FirewallDomainRedirectionAction = awstypes.FirewallDomainRedirectionAction(v.(string))
	}

	if v, ok := d.GetOk("q_type"); ok {
		input.Qtype = aws.String(v.(string))
	}

	_, err = conn.UpdateFirewallRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Firewall Rule (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFirewallRuleRead(ctx, d, meta)...)
}

func resourceFirewallRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	firewallRuleGroupID, firewallDomainListID, err := firewallRuleParseResourceID(d.Id())

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
	_, err = conn.DeleteFirewallRule(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Firewall Rule (%s): %s", d.Id(), err)
	}

	return diags
}

const firewallRuleIDSeparator = ":"

func firewallRuleCreateResourceID(firewallRuleGroupID, firewallDomainListID string) string {
	parts := []string{firewallRuleGroupID, firewallDomainListID}
	id := strings.Join(parts, firewallRuleIDSeparator)

	return id
}

func firewallRuleParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, firewallRuleIDSeparator, 2)

	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected firewall_rule_group_id%[2]sfirewall_domain_list_id", id, firewallRuleIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findFirewallRuleByTwoPartKey(ctx context.Context, conn *route53resolver.Client, firewallRuleGroupID, firewallDomainListID string) (*awstypes.FirewallRule, error) {
	output, err := findFirewallRules(ctx, conn, firewallRuleGroupID, func(rule awstypes.FirewallRule) bool {
		return aws.ToString(rule.FirewallDomainListId) == firewallDomainListID
	})

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(firewallRuleGroupID)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, firewallRuleGroupID)
	}

	return &output[0], nil
}

func findFirewallRules(ctx context.Context, conn *route53resolver.Client, firewallRuleGroupID string, f func(awstypes.FirewallRule) bool) ([]awstypes.FirewallRule, error) {
	input := &route53resolver.ListFirewallRulesInput{
		FirewallRuleGroupId: aws.String(firewallRuleGroupID),
	}
	var output []awstypes.FirewallRule

	pages := route53resolver.NewListFirewallRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.FirewallRules {
			if f(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
