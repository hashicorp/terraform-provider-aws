// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ses

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ses_domain_identity", name="Domain Identity")
func dataSourceDomainIdentity() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDomainIdentityRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDomain: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringDoesNotMatch(regexache.MustCompile(`\.$`), "cannot end with a period"),
				},
				"verification_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceDomainIdentityRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.SESClient(ctx)

	domainName := d.Get(names.AttrDomain).(string)
	verificationAttrs, err := findIdentityVerificationAttributesByIdentity(ctx, conn, domainName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain Identity (%s) verification: %s", domainName, err)
	}

	d.SetId(domainName)
	d.Set(names.AttrARN, identityARN(ctx, c, domainName))
	d.Set(names.AttrDomain, domainName)
	d.Set("verification_token", verificationAttrs.VerificationToken)

	return diags
}
