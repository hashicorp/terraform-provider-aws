// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_domain_identity_verification", name="Domain Identity Verification")
func resourceDomainIdentityVerification() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainIdentityVerificationCreate,
		ReadWithoutTimeout:   resourceDomainIdentityVerificationRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
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
	_, err := tfresource.RetryUntilEqual(ctx, d.Timeout(schema.TimeoutCreate), awstypes.VerificationStatusSuccess, func() (awstypes.VerificationStatus, error) {
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
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	att, err := findIdentityVerificationAttributesByIdentity(ctx, conn, d.Id())

	if err == nil {
		if status := att.VerificationStatus; status != awstypes.VerificationStatusSuccess {
			err = &retry.NotFoundError{
				Message: string(status),
			}
		}
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Domain Identity Verification (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain Identity Verification (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDomain, d.Id())

	return diags
}
