// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_route53_resolver_firewall_config")
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
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	id := d.Get("resource_id").(string)
	firewallConfig, err := findFirewallConfigByResourceID(ctx, conn, id)

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Firewall Config (%s): %s", id, err)
	}

	d.SetId(aws.StringValue(firewallConfig.Id))
	d.Set("firewall_fail_open", firewallConfig.FirewallFailOpen)
	d.Set("owner_id", firewallConfig.OwnerId)
	d.Set("resource_id", firewallConfig.ResourceId)

	return nil
}

func findFirewallConfigByResourceID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallConfig, error) {
	input := &route53resolver.GetFirewallConfigInput{
		ResourceId: aws.String(id),
	}

	output, err := conn.GetFirewallConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FirewallConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.FirewallConfig, nil
}
