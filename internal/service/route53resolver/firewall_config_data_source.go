package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceFirewallConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallConfigRead,

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

func dataSourceFirewallConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.GetFirewallConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	output, err := conn.GetFirewallConfigWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Firewall Config: %s", err)
	}

	firewallConfig := output.FirewallConfig
	d.SetId(aws.StringValue(firewallConfig.Id))
	d.Set("firewall_fail_open", firewallConfig.FirewallFailOpen)
	d.Set("owner_id", firewallConfig.OwnerId)
	d.Set("resource_id", firewallConfig.ResourceId)

	return nil
}
