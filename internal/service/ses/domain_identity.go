// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ses

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_domain_identity", name="Domain Identity")
func resourceDomainIdentity() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainIdentityCreate,
		ReadWithoutTimeout:   resourceDomainIdentityRead,
		DeleteWithoutTimeout: resourceDomainIdentityDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDomain: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
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

func resourceDomainIdentityCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	domainName := d.Get(names.AttrDomain).(string)
	input := ses.VerifyDomainIdentityInput{
		Domain: aws.String(domainName),
	}

	_, err := conn.VerifyDomainIdentity(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "requesting SES Domain Identity (%s) verification: %s", domainName, err)
	}

	d.SetId(domainName)

	return append(diags, resourceDomainIdentityRead(ctx, d, meta)...)
}

func resourceDomainIdentityRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.SESClient(ctx)

	verificationAttrs, err := findIdentityVerificationAttributesByIdentity(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SES Domain Identity (%s) verification not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain Identity (%s) verification: %s", d.Id(), err)
	}

	d.Set(names.AttrARN, identityARN(ctx, c, d.Id()))
	d.Set(names.AttrDomain, d.Id())
	d.Set("verification_token", verificationAttrs.VerificationToken)

	return diags
}

func resourceDomainIdentityDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting SES Domain Identity: %s", d.Id())
	input := ses.DeleteIdentityInput{
		Identity: aws.String(d.Id()),
	}
	_, err := conn.DeleteIdentity(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Domain Identity (%s): %s", d.Id(), err)
	}

	return diags
}
