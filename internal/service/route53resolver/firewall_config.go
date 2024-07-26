// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_firewall_config")
func ResourceFirewallConfig() *schema.Resource {
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.FirewallFailOpenStatus_Values(), false),
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

func resourceFirewallConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = aws.String(v.(string))
	}

	output, err := conn.UpdateFirewallConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Config: %s", err)
	}

	d.SetId(aws.StringValue(output.FirewallConfig.Id))

	return append(diags, resourceFirewallConfigRead(ctx, d, meta)...)
}

func resourceFirewallConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	firewallConfig, err := FindFirewallConfigByID(ctx, conn, d.Id())

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

func resourceFirewallConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = aws.String(v.(string))
	}

	_, err := conn.UpdateFirewallConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Firewall Config: %s", err)
	}

	return append(diags, resourceFirewallConfigRead(ctx, d, meta)...)
}

func resourceFirewallConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Config: %s", d.Id())
	_, err := conn.UpdateFirewallConfigWithContext(ctx, &route53resolver.UpdateFirewallConfigInput{
		ResourceId:       aws.String(d.Get(names.AttrResourceID).(string)),
		FirewallFailOpen: aws.String(route53resolver.FirewallFailOpenStatusDisabled),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Firewall Config (%s): %s", d.Id(), err)
	}

	return diags
}

func FindFirewallConfigByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallConfig, error) {
	input := &route53resolver.ListFirewallConfigsInput{}
	var output *route53resolver.FirewallConfig

	// GetFirewallConfig does not support query by ID.
	err := conn.ListFirewallConfigsPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FirewallConfigs {
			if aws.StringValue(v.Id) == id {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{LastRequest: input}
	}

	return output, nil
}
