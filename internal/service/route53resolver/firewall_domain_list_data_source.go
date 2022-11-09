package route53resolver

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceFirewallDomainList() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFirewallDomainListRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creator_request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"firewall_domain_list_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_owner_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modification_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFirewallDomainListRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.GetFirewallDomainListInput{
		FirewallDomainListId: aws.String(d.Get("firewall_domain_list_id").(string)),
	}

	output, err := conn.GetFirewallDomainList(input)

	if err != nil {
		return fmt.Errorf("error getting Route53 Firewall Domain List: %w", err)
	}

	if output == nil {
		return fmt.Errorf("no Route53 Firewall Domain List found matching criteria; try different search")
	}

	firewallDomainList := output.FirewallDomainList
	d.SetId(aws.StringValue(firewallDomainList.Id))
	d.Set("arn", firewallDomainList.Arn)
	d.Set("creation_time", firewallDomainList.CreationTime)
	d.Set("creator_request_id", firewallDomainList.CreatorRequestId)
	d.Set("domain_count", firewallDomainList.DomainCount)
	d.Set("firewall_domain_list_id", firewallDomainList.Id)
	d.Set("name", firewallDomainList.Name)
	d.Set("managed_owner_name", firewallDomainList.ManagedOwnerName)
	d.Set("modification_time", firewallDomainList.ModificationTime)
	d.Set("status", firewallDomainList.Status)
	d.Set("status_message", firewallDomainList.StatusMessage)

	return nil
}
