package route53resolver

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallRuleCreate,
		Read:   resourceFirewallRuleRead,
		Update: resourceFirewallRuleUpdate,
		Delete: resourceFirewallRuleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"action": {
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
			"firewall_rule_group_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validResolverName,
			},
			"priority": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	firewallDomainListID := d.Get("firewall_domain_list_id").(string)
	firewallRuleGroupID := d.Get("firewall_rule_group_id").(string)
	id := firewallRuleCreateResourceID(firewallRuleGroupID, firewallDomainListID)
	name := d.Get("name").(string)
	input := &route53resolver.CreateFirewallRuleInput{
		Action:               aws.String(d.Get("action").(string)),
		CreatorRequestId:     aws.String(resource.PrefixedUniqueId("tf-r53-resolver-firewall-rule-")),
		FirewallRuleGroupId:  aws.String(firewallRuleGroupID),
		FirewallDomainListId: aws.String(firewallDomainListID),
		Name:                 aws.String(name),
		Priority:             aws.Int64(int64(d.Get("priority").(int))),
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

	_, err := conn.CreateFirewallRule(input)

	if err != nil {
		return fmt.Errorf("creating Route53 Resolver Firewall Rule (%s): %w", name, err)
	}

	d.SetId(id)

	return resourceFirewallRuleRead(d, meta)
}

func resourceFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	firewallRuleGroupID, firewallDomainListID, err := FirewallRuleParseResourceID(d.Id())

	if err != nil {
		return err
	}

	firewallRule, err := FindFirewallRuleByTwoPartKey(conn, firewallRuleGroupID, firewallDomainListID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Route53 Resolver Firewall Rule (%s): %w", d.Id(), err)
	}

	d.Set("action", firewallRule.Action)
	d.Set("block_override_dns_type", firewallRule.BlockOverrideDnsType)
	d.Set("block_override_domain", firewallRule.BlockOverrideDomain)
	d.Set("block_override_ttl", firewallRule.BlockOverrideTtl)
	d.Set("block_response", firewallRule.BlockResponse)
	d.Set("firewall_rule_group_id", firewallRule.FirewallRuleGroupId)
	d.Set("firewall_domain_list_id", firewallRule.FirewallDomainListId)
	d.Set("name", firewallRule.Name)
	d.Set("priority", firewallRule.Priority)

	return nil
}

func resourceFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	firewallRuleGroupID, firewallDomainListID, err := FirewallRuleParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &route53resolver.UpdateFirewallRuleInput{
		Action:               aws.String(d.Get("action").(string)),
		FirewallDomainListId: aws.String(firewallDomainListID),
		FirewallRuleGroupId:  aws.String(firewallRuleGroupID),
		Name:                 aws.String(d.Get("name").(string)),
		Priority:             aws.Int64(int64(d.Get("priority").(int))),
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

	_, err = conn.UpdateFirewallRule(input)

	if err != nil {
		return fmt.Errorf("updating Route53 Resolver Firewall Rule (%s): %w", d.Id(), err)
	}

	return resourceFirewallRuleRead(d, meta)
}

func resourceFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	firewallRuleGroupID, firewallDomainListID, err := FirewallRuleParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Rule: %s", d.Id())
	_, err = conn.DeleteFirewallRule(&route53resolver.DeleteFirewallRuleInput{
		FirewallDomainListId: aws.String(firewallDomainListID),
		FirewallRuleGroupId:  aws.String(firewallRuleGroupID),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Route53 Resolver Firewall Rule (%s): %w", d.Id(), err)
	}

	return nil
}

const firewallRuleIDSeparator = ":"

func firewallRuleCreateResourceID(firewallRuleGroupID, firewallDomainListID string) string {
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

func FindFirewallRuleByTwoPartKey(conn *route53resolver.Route53Resolver, firewallRuleGroupID, firewallDomainListID string) (*route53resolver.FirewallRule, error) {
	input := &route53resolver.ListFirewallRulesInput{
		FirewallRuleGroupId: aws.String(firewallRuleGroupID),
	}
	var output *route53resolver.FirewallRule

	err := conn.ListFirewallRulesPages(input, func(page *route53resolver.ListFirewallRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FirewallRules {
			if aws.StringValue(v.FirewallDomainListId) == firewallDomainListID {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{LastRequest: input}
	}

	return output, nil
}
