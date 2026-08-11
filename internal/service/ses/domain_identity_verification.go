// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ses

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_domain_identity_verification", name="Domain Identity Verification")
func resourceDomainIdentityVerification() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainIdentityVerificationCreate,
		ReadWithoutTimeout:   resourceDomainIdentityVerificationRead,
		DeleteWithoutTimeout: schema.NoopContext,

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
			}
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
		},
	}
}

func resourceDomainIdentityVerificationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	domainName := d.Get(names.AttrDomain).(string)
	_, err := tfresource.RetryUntilEqual(ctx, d.Timeout(schema.TimeoutCreate), awstypes.VerificationStatusSuccess, func(ctx context.Context) (awstypes.VerificationStatus, error) {
		att, err := findIdentityVerificationAttributesByIdentity(ctx, conn, domainName)

		if err != nil {
			return "", err
		}

		return att.VerificationStatus, nil
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Domain Identity Verification (%s): %s", domainName, err)
	}

	d.SetId(domainName)

	return append(diags, resourceDomainIdentityVerificationRead(ctx, d, meta)...)
}

func resourceDomainIdentityVerificationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.SESClient(ctx)

	att, err := findIdentityVerificationAttributesByIdentity(ctx, conn, d.Id())

	if err == nil {
		if status := att.VerificationStatus; status != awstypes.VerificationStatusSuccess {
			err = &retry.NotFoundError{
				Message: string(status),
			}
		}
	}

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SES Domain Identity Verification (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain Identity Verification (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, identityARN(ctx, c, d.Id()))
	d.Set(names.AttrDomain, d.Id())

	return diags
}
