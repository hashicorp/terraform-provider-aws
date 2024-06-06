// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_domain_identity")
func ResourceDomainIdentity() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainIdentityCreate,
		ReadWithoutTimeout:   resourceDomainIdentityRead,
		DeleteWithoutTimeout: resourceDomainIdentityDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			"verification_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDomainIdentityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	domainName := d.Get(names.AttrDomain).(string)

	createOpts := &ses.VerifyDomainIdentityInput{
		Domain: aws.String(domainName),
	}

	_, err := conn.VerifyDomainIdentityWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "requesting SES domain identity verification: %s", err)
	}

	d.SetId(domainName)

	return append(diags, resourceDomainIdentityRead(ctx, d, meta)...)
}

func resourceDomainIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	domainName := d.Id()
	d.Set(names.AttrDomain, domainName)

	readOpts := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, readOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain Identity (%s): %s", domainName, err)
	}

	verificationAttrs, ok := response.VerificationAttributes[domainName]
	if !ok {
		log.Printf("[WARN] SES Domain Identity (%s) not found, removing from state", d.Id())
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
	d.Set("verification_token", verificationAttrs.VerificationToken)
	return diags
}

func resourceDomainIdentityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	domainName := d.Get(names.AttrDomain).(string)

	deleteOpts := &ses.DeleteIdentityInput{
		Identity: aws.String(domainName),
	}

	_, err := conn.DeleteIdentityWithContext(ctx, deleteOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES domain identity: %s", err)
	}

	return diags
}
