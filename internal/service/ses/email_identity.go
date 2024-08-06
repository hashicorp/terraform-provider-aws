// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_email_identity")
func ResourceEmailIdentity() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEmailIdentityCreate,
		ReadWithoutTimeout:   resourceEmailIdentityRead,
		DeleteWithoutTimeout: resourceEmailIdentityDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEmail: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return strings.TrimSuffix(v.(string), ".")
				},
			},
		},
	}
}

func resourceEmailIdentityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	email := d.Get(names.AttrEmail).(string)
	email = strings.TrimSuffix(email, ".")

	createOpts := &ses.VerifyEmailIdentityInput{
		EmailAddress: aws.String(email),
	}

	_, err := conn.VerifyEmailIdentityWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "requesting SES email identity verification: %s", err)
	}

	d.SetId(email)

	return append(diags, resourceEmailIdentityRead(ctx, d, meta)...)
}

func resourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	email := d.Id()
	d.Set(names.AttrEmail, email)

	readOpts := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(email),
		},
	}

	response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, readOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Identity Verification Attributes (%s): %s", d.Id(), err)
	}

	_, ok := response.VerificationAttributes[email]
	if !ok {
		log.Printf("[WARN] SES Identity Verification Attributes (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
		Service:   "ses",
	}.String()
	d.Set(names.AttrARN, arn)
	return diags
}

func resourceEmailIdentityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	email := d.Get(names.AttrEmail).(string)

	deleteOpts := &ses.DeleteIdentityInput{
		Identity: aws.String(email),
	}

	_, err := conn.DeleteIdentityWithContext(ctx, deleteOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES email identity: %s", err)
	}

	return diags
}
