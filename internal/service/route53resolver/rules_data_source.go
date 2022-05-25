package route53resolver

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRulesRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.Any(
					verify.ValidAccountID,
					// The owner of the default Internet Resolver rule.
					validation.StringInSlice([]string{"Route 53 Resolver"}, false),
				),
			},

			"resolver_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"resolver_rule_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"rule_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					route53resolver.RuleTypeOptionForward,
					route53resolver.RuleTypeOptionSystem,
					route53resolver.RuleTypeOptionRecursive,
				}, false),
			},

			"share_status": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					route53resolver.ShareStatusNotShared,
					route53resolver.ShareStatusSharedWithMe,
					route53resolver.ShareStatusSharedByMe,
				}, false),
			},
		},
	}
}

func dataSourceRulesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	req := &route53resolver.ListResolverRulesInput{}
	resolverRuleIds := []*string{}

	log.Printf("[DEBUG] Listing Route53 Resolver rules: %s", req)
	err := conn.ListResolverRulesPages(req, func(page *route53resolver.ListResolverRulesOutput, lastPage bool) bool {
		for _, rule := range page.ResolverRules {
			if v, ok := d.GetOk("name_regex"); ok && !regexp.MustCompile(v.(string)).MatchString(aws.StringValue(rule.Name)) {
				continue
			}
			if v, ok := d.GetOk("owner_id"); ok && aws.StringValue(rule.OwnerId) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("resolver_endpoint_id"); ok && aws.StringValue(rule.ResolverEndpointId) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("rule_type"); ok && aws.StringValue(rule.RuleType) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("share_status"); ok && aws.StringValue(rule.ShareStatus) != v.(string) {
				continue
			}

			resolverRuleIds = append(resolverRuleIds, rule.Id)
		}
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("error getting Route53 Resolver rules: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	err = d.Set("resolver_rule_ids", flex.FlattenStringSet(resolverRuleIds))
	if err != nil {
		return fmt.Errorf("error setting resolver_rule_ids: %w", err)
	}

	return nil
}
