// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_firewall_config", name="Firewall Config")
func resourceFirewallConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallConfigCreate,
		ReadWithoutTimeout:   resourceFirewallConfigRead,
		UpdateWithoutTimeout: resourceFirewallConfigUpdate,
		DeleteWithoutTimeout: resourceFirewallConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"firewall_fail_open": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FirewallFailOpenStatus](),
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFirewallConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = awstypes.FirewallFailOpenStatus(v.(string))
	}

	output, err := conn.UpdateFirewallConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Config: %s", err)
	}

	d.SetId(aws.ToString(output.FirewallConfig.Id))

	return append(diags, resourceFirewallConfigRead(ctx, d, meta)...)
}

func resourceFirewallConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	firewallConfig, err := findFirewallConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Config (%s): %s", d.Id(), err)
	}

	d.Set("firewall_fail_open", firewallConfig.FirewallFailOpen)
	d.Set(names.AttrOwnerID, firewallConfig.OwnerId)
	d.Set(names.AttrResourceID, firewallConfig.ResourceId)

	return diags
}

func resourceFirewallConfigUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = awstypes.FirewallFailOpenStatus(v.(string))
	}

	_, err := conn.UpdateFirewallConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Firewall Config: %s", err)
	}

	return append(diags, resourceFirewallConfigRead(ctx, d, meta)...)
}

func resourceFirewallConfigDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Config: %s", d.Id())
	_, err := conn.UpdateFirewallConfig(ctx, &route53resolver.UpdateFirewallConfigInput{
		ResourceId:       aws.String(d.Get(names.AttrResourceID).(string)),
		FirewallFailOpen: awstypes.FirewallFailOpenStatusDisabled,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Firewall Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findFirewallConfigByID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallConfig, error) {
	input := &route53resolver.ListFirewallConfigsInput{}

	// GetFirewallConfig does not support query by ID.
	pages := route53resolver.NewListFirewallConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.FirewallConfigs {
			if aws.ToString(v.Id) == id {
				return &v, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}
