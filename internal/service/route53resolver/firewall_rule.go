// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
			"confidence_threshold": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ConfidenceThreshold](),
				ConflictsWith:    []string{"firewall_domain_list_id"},
			},
			"dns_threat_protection": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DnsThreatProtection](),
				ConflictsWith:    []string{"firewall_domain_list_id"},
			},
			"firewall_domain_list_id": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ValidateFunc:  validation.StringLenBetween(1, 64),
				ConflictsWith: []string{"dns_threat_protection", "confidence_threshold"},
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
			"firewall_threat_protection_id": {
				Type:     schema.TypeString,
				Computed: true,
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
		CustomizeDiff: firewallRuleCustomizeDiff,
	}
}

func resourceFirewallRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	firewallRuleGroupID := d.Get("firewall_rule_group_id").(string)
	name := d.Get(names.AttrName).(string)
	input := &route53resolver.CreateFirewallRuleInput{
		Action:              awstypes.Action(d.Get(names.AttrAction).(string)),
		CreatorRequestId:    aws.String(sdkid.PrefixedUniqueId("tf-r53-resolver-firewall-rule-")),
		FirewallRuleGroupId: aws.String(firewallRuleGroupID),
		Name:                aws.String(name),
		Priority:            aws.Int32(int32(d.Get(names.AttrPriority).(int))),
	}

	// Standard rule (domain list-based)
	if v, ok := d.GetOk("firewall_domain_list_id"); ok {
		input.FirewallDomainListId = aws.String(v.(string))
		input.FirewallDomainRedirectionAction = awstypes.FirewallDomainRedirectionAction(d.Get("firewall_domain_redirection_action").(string))
	}

	// Advanced rule (DNS threat protection)
	if v, ok := d.GetOk("dns_threat_protection"); ok {
		input.DnsThreatProtection = awstypes.DnsThreatProtection(v.(string))
	}

	if v, ok := d.GetOk("confidence_threshold"); ok {
		input.ConfidenceThreshold = awstypes.ConfidenceThreshold(v.(string))
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

	output, err := conn.CreateFirewallRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Rule (%s): %s", name, err)
	}

	// Set the resource ID based on rule type
	if output.FirewallRule.FirewallThreatProtectionId != nil {
		// Advanced rule
		d.SetId(firewallRuleCreateResourceID(firewallRuleGroupID, aws.ToString(output.FirewallRule.FirewallThreatProtectionId)))
	} else {
		// Standard rule
		d.SetId(firewallRuleCreateResourceID(firewallRuleGroupID, aws.ToString(output.FirewallRule.FirewallDomainListId)))
	}

	return append(diags, resourceFirewallRuleRead(ctx, d, meta)...)
}

func resourceFirewallRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	firewallRuleGroupID, ruleIdentifier, err := firewallRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	firewallRule, err := findFirewallRuleByTwoPartKey(ctx, conn, firewallRuleGroupID, ruleIdentifier)

	if !d.IsNewResource() && retry.NotFound(err) {
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
	d.Set("confidence_threshold", firewallRule.ConfidenceThreshold)
	d.Set("dns_threat_protection", firewallRule.DnsThreatProtection)
	d.Set("firewall_domain_list_id", firewallRule.FirewallDomainListId)
	d.Set("firewall_domain_redirection_action", firewallRule.FirewallDomainRedirectionAction)
	d.Set("firewall_rule_group_id", firewallRule.FirewallRuleGroupId)
	d.Set("firewall_threat_protection_id", firewallRule.FirewallThreatProtectionId)
	d.Set(names.AttrName, firewallRule.Name)
	d.Set(names.AttrPriority, firewallRule.Priority)
	d.Set("q_type", firewallRule.Qtype)

	return diags
}

func resourceFirewallRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	firewallRuleGroupID, ruleIdentifier, err := firewallRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &route53resolver.UpdateFirewallRuleInput{
		Action:              awstypes.Action(d.Get(names.AttrAction).(string)),
		FirewallRuleGroupId: aws.String(firewallRuleGroupID),
		Name:                aws.String(d.Get(names.AttrName).(string)),
		Priority:            aws.Int32(int32(d.Get(names.AttrPriority).(int))),
	}

	// Standard rule (domain list-based)
	if v, ok := d.GetOk("firewall_domain_list_id"); ok {
		input.FirewallDomainListId = aws.String(v.(string))
	}

	// Advanced rule (DNS threat protection)
	if v, ok := d.GetOk("firewall_threat_protection_id"); ok {
		input.FirewallThreatProtectionId = aws.String(v.(string))
	} else if _, ok := d.GetOk("dns_threat_protection"); ok {
		// Use the identifier from the resource ID for advanced rules
		input.FirewallThreatProtectionId = aws.String(ruleIdentifier)
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

	firewallRuleGroupID, ruleIdentifier, err := firewallRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &route53resolver.DeleteFirewallRuleInput{
		FirewallRuleGroupId: aws.String(firewallRuleGroupID),
	}

	// Standard rule (domain list-based)
	if v, ok := d.GetOk("firewall_domain_list_id"); ok {
		input.FirewallDomainListId = aws.String(v.(string))
	}

	// Advanced rule (DNS threat protection)
	if v, ok := d.GetOk("firewall_threat_protection_id"); ok {
		input.FirewallThreatProtectionId = aws.String(v.(string))
	} else if _, ok := d.GetOk("dns_threat_protection"); ok {
		// Use the identifier from the resource ID for advanced rules
		input.FirewallThreatProtectionId = aws.String(ruleIdentifier)
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

func findFirewallRuleByTwoPartKey(ctx context.Context, conn *route53resolver.Client, firewallRuleGroupID, ruleIdentifier string) (*awstypes.FirewallRule, error) {
	input := route53resolver.ListFirewallRulesInput{
		FirewallRuleGroupId: aws.String(firewallRuleGroupID),
	}
	output, err := findFirewallRules(ctx, conn, &input, func(v *awstypes.FirewallRule) bool {
		// Match standard rules by firewall_domain_list_id
		if aws.ToString(v.FirewallDomainListId) == ruleIdentifier {
			return true
		}
		// Match advanced rules by firewall_threat_protection_id
		if aws.ToString(v.FirewallThreatProtectionId) == ruleIdentifier {
			return true
		}
		return false
	})

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFirewallRules(ctx context.Context, conn *route53resolver.Client, input *route53resolver.ListFirewallRulesInput, f tfslices.Predicate[*awstypes.FirewallRule]) ([]awstypes.FirewallRule, error) {
	var output []awstypes.FirewallRule

	pages := route53resolver.NewListFirewallRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.FirewallRules {
			if f(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
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
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected firewall_rule_group_id%[2]srule_identifier", id, firewallRuleIDSeparator)
	}

	return parts[0], parts[1], nil
}

var firewallRuleCustomizeDiff = customdiff.All(
	func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		// Use GetRawConfig to check if attributes are set in the configuration,
		// regardless of whether their values are known yet (e.g., references to other resources)
		rawConfig := d.GetRawConfig()
		domainListVal := rawConfig.GetAttr("firewall_domain_list_id")
		threatProtectionVal := rawConfig.GetAttr("dns_threat_protection")
		confidenceThresholdVal := rawConfig.GetAttr("confidence_threshold")

		hasDomainList := !domainListVal.IsNull()
		hasThreatProtection := !threatProtectionVal.IsNull()
		hasConfidenceThreshold := !confidenceThresholdVal.IsNull()

		// Must have either firewall_domain_list_id OR (dns_threat_protection AND confidence_threshold)
		if !hasDomainList && !hasThreatProtection {
			return fmt.Errorf("one of firewall_domain_list_id or dns_threat_protection must be specified")
		}

		// For advanced rules, both dns_threat_protection and confidence_threshold are required
		if hasThreatProtection && !hasConfidenceThreshold {
			return fmt.Errorf("confidence_threshold is required when dns_threat_protection is specified")
		}

		if hasConfidenceThreshold && !hasThreatProtection {
			return fmt.Errorf("dns_threat_protection is required when confidence_threshold is specified")
		}

		return nil
	},
)
