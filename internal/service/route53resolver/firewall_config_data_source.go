package route53resolver

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceFirewallConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFirewallConfigRead,

		Schema: map[string]*schema.Schema{
			"firewall_fail_open": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceFirewallConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.GetFirewallConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	output, err := conn.GetFirewallConfig(input)

	if err != nil {
		return fmt.Errorf("error getting Route53 Firewall Config: %w", err)
	}

	if output == nil {
		return fmt.Errorf("no  Route53 Firewall Config found matching criteria; try different search")
	}

	d.SetId(aws.StringValue(output.FirewallConfig.Id))
	d.Set("firewall_fail_open", output.FirewallConfig.FirewallFailOpen)
	d.Set("owner_id", output.FirewallConfig.OwnerId)
	d.Set("resource_id", output.FirewallConfig.ResourceId)

	return nil
}
