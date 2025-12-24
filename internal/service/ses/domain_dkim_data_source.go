// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ses_domain_dkim", name="Domain DKIM")
func dataSourceDomainDKIM() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDomainDKIMRead,

		Schema: map[string]*schema.Schema{
			"dkim_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"dkim_tokens": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dkim_verification_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	DSNameDomainDKIM = "Domain DKIM Data Source"
)

func dataSourceDomainDKIMRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	domainName := d.Get(names.AttrDomain).(string)

	out, err := findIdentityDKIMAttributesByIdentity(ctx, conn, domainName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain Identity (%s) attributes: %s", domainName, err)
	}

	d.SetId(domainName)
	d.Set(names.AttrDomain, domainName)
	d.Set("dkim_tokens", out.DkimTokens)
	d.Set("dkim_enabled", out.DkimEnabled)
	d.Set("dkim_verification_status", out.DkimVerificationStatus)
	return diags
}
