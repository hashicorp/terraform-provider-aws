// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_domain_dkim")
func ResourceDomainDKIM() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainDKIMCreate,
		ReadWithoutTimeout:   resourceDomainDKIMRead,
		DeleteWithoutTimeout: resourceDomainDKIMDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dkim_tokens": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDomainDKIMCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	domainName := d.Get(names.AttrDomain).(string)

	createOpts := &ses.VerifyDomainDkimInput{
		Domain: aws.String(domainName),
	}

	_, err := conn.VerifyDomainDkimWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "requesting SES domain identity verification: %s", err)
	}

	d.SetId(domainName)

	return append(diags, resourceDomainDKIMRead(ctx, d, meta)...)
}

func resourceDomainDKIMRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	domainName := d.Id()
	d.Set(names.AttrDomain, domainName)

	readOpts := &ses.GetIdentityDkimAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	response, err := conn.GetIdentityDkimAttributesWithContext(ctx, readOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain DKIM (%s): %s", d.Id(), err)
	}

	verificationAttrs, ok := response.DkimAttributes[domainName]
	if !ok {
		log.Printf("[WARN] SES Domain DKIM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("dkim_tokens", aws.StringValueSlice(verificationAttrs.DkimTokens))
	return diags
}

func resourceDomainDKIMDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
