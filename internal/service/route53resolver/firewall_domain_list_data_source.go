package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceFirewallDomainList() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallDomainListRead,

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

func dataSourceFirewallDomainListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	id := d.Get("firewall_domain_list_id").(string)
	firewallDomainList, err := FindFirewallDomainListByID(ctx, conn, id)

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Firewall Domain List (%s): %s", id, err)
	}

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
