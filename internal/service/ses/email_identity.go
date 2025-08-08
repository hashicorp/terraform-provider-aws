// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_email_identity", name="Email Identity")
func resourceEmailIdentity() *schema.Resource {
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
				StateFunc: func(v any) string {
					return strings.TrimSuffix(v.(string), ".")
				},
			},
		},
	}
}

func resourceEmailIdentityCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	email := strings.TrimSuffix(d.Get(names.AttrEmail).(string), ".")
	input := &ses.VerifyEmailIdentityInput{
		EmailAddress: aws.String(email),
	}

	_, err := conn.VerifyEmailIdentity(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "requesting SES Email Identity (%s) verification: %s", email, err)
	}

	d.SetId(email)

	return append(diags, resourceEmailIdentityRead(ctx, d, meta)...)
}

func resourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	_, err := findIdentityVerificationAttributesByIdentity(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Email Identity (%s) verification not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Email Identity (%s) verification: %s", d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
		Service:   "ses",
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrEmail, d.Id())

	return diags
}

func resourceEmailIdentityDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting SES Email Identity: %s", d.Id())
	_, err := conn.DeleteIdentity(ctx, &ses.DeleteIdentityInput{
		Identity: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Email Identity (%s): %s", d.Id(), err)
	}

	return diags
}

func findIdentityVerificationAttributesByIdentity(ctx context.Context, conn *ses.Client, identity string) (*awstypes.IdentityVerificationAttributes, error) {
	input := &ses.GetIdentityVerificationAttributesInput{
		Identities: []string{identity},
	}
	output, err := findIdentityVerificationAttributes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if v, ok := output[identity]; ok {
		return &v, nil
	}

	return nil, &retry.NotFoundError{}
}

func findIdentityVerificationAttributes(ctx context.Context, conn *ses.Client, input *ses.GetIdentityVerificationAttributesInput) (map[string]awstypes.IdentityVerificationAttributes, error) {
	output, err := conn.GetIdentityVerificationAttributes(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.VerificationAttributes == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VerificationAttributes, nil
}
