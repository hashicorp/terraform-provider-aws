// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_resolver_firewall_config", name="Firewall Config")
func dataSourceFirewallConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallConfigRead,

		Schema: map[string]*schema.Schema{
			"firewall_fail_open": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceFirewallConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	id := d.Get(names.AttrResourceID).(string)
	firewallConfig, err := findFirewallConfigByResourceID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Config (%s): %s", id, err)
	}

	d.SetId(aws.ToString(firewallConfig.Id))
	d.Set("firewall_fail_open", firewallConfig.FirewallFailOpen)
	d.Set(names.AttrOwnerID, firewallConfig.OwnerId)
	d.Set(names.AttrResourceID, firewallConfig.ResourceId)

	return diags
}

func findFirewallConfigByResourceID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallConfig, error) {
	input := &route53resolver.GetFirewallConfigInput{
		ResourceId: aws.String(id),
	}

	output, err := conn.GetFirewallConfig(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
