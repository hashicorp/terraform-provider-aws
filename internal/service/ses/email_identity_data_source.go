// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
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

// @SDKDataSource("aws_ses_email_identity")
func DataSourceEmailIdentity() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEmailIdentityRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEmail: {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(v interface{}) string {
					return strings.TrimSuffix(v.(string), ".")
				},
			},
		},
	}
}

func dataSourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	email := d.Get(names.AttrEmail).(string)
	email = strings.TrimSuffix(email, ".")

	d.SetId(email)
	d.Set(names.AttrEmail, email)

	readOpts := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(email),
		},
	}

	response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, readOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "[WARN] Error fetching identity verification attributes for %s: %s", email, err)
	}

	_, ok := response.VerificationAttributes[email]
	if !ok {
		return sdkdiag.AppendErrorf(diags, "[WARN] Email not listed in response when fetching verification attributes for %s", d.Id())
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("identity/%s", email),
		Service:   "ses",
	}.String()
	d.Set(names.AttrARN, arn)
	return diags
}
