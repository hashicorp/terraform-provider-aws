// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_domain_identity_verification")
func ResourceDomainIdentityVerification() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainIdentityVerificationCreate,
		ReadWithoutTimeout:   resourceDomainIdentityVerificationRead,
		DeleteWithoutTimeout: resourceDomainIdentityVerificationDelete,

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

func getIdentityVerificationAttributes(ctx context.Context, conn *ses.SES, domainName string) (*ses.IdentityVerificationAttributes, error) {
	input := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("getting identity verification attributes: %s", err)
	}

	return response.VerificationAttributes[domainName], nil
}

func resourceDomainIdentityVerificationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)
	domainName := d.Get(names.AttrDomain).(string)
	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		att, err := getIdentityVerificationAttributes(ctx, conn, domainName)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("getting identity verification attributes: %s", err))
		}

		if att == nil {
			return retry.NonRetryableError(fmt.Errorf("SES Domain Identity %s not found in AWS", domainName))
		}

		if aws.StringValue(att.VerificationStatus) != ses.VerificationStatusSuccess {
			return retry.RetryableError(fmt.Errorf("Expected domain verification Success, but was in state %s", aws.StringValue(att.VerificationStatus)))
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		var att *ses.IdentityVerificationAttributes
		att, err = getIdentityVerificationAttributes(ctx, conn, domainName)

		if att != nil && aws.StringValue(att.VerificationStatus) != ses.VerificationStatusSuccess {
			return sdkdiag.AppendErrorf(diags, "Expected domain verification Success, but was in state %s", aws.StringValue(att.VerificationStatus))
		}
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES domain identity verification: %s", err)
	}

	log.Printf("[INFO] Domain verification successful for %s", domainName)
	d.SetId(domainName)
	return append(diags, resourceDomainIdentityVerificationRead(ctx, d, meta)...)
}

func resourceDomainIdentityVerificationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	domainName := d.Id()
	d.Set(names.AttrDomain, domainName)

	att, err := getIdentityVerificationAttributes(ctx, conn, domainName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain Identity Verification (%s): %s", domainName, err)
	}

	if att == nil {
		log.Printf("[WARN] SES Domain Identity Verification (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if aws.StringValue(att.VerificationStatus) != ses.VerificationStatusSuccess {
		log.Printf("[WARN] Expected domain verification Success, but was %s, tainting verification", aws.StringValue(att.VerificationStatus))
		d.SetId("")
		return diags
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)

	return diags
}

func resourceDomainIdentityVerificationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// No need to do anything, domain identity will be deleted when aws_ses_domain_identity is deleted
	diags diag.Diagnostics

	return diags
}
