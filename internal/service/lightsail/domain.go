// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_lightsail_domain")
func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		DeleteWithoutTimeout: resourceDomainDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	_, err := conn.CreateDomain(ctx, &lightsail.CreateDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lightsail Domain: %s", err)
	}

	d.SetId(d.Get("domain_name").(string))

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	resp, err := conn.GetDomain(ctx, &lightsail.GetDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if err != nil {
		if IsANotFoundError(err) {
			log.Printf("[WARN] Lightsail Domain (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Lightsail Domain (%s):%s", d.Id(), err)
	}

	d.Set("arn", resp.Domain.Arn)
	return diags
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	_, err := conn.DeleteDomain(ctx, &lightsail.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lightsail Domain (%s):%s", d.Id(), err)
	}

	return nil
}
