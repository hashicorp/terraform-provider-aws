// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ses_email_identity", name="Email Identity")
func dataSourceEmailIdentity() *schema.Resource {
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
				StateFunc: func(v any) string {
					return strings.TrimSuffix(v.(string), ".")
				},
			},
		},
	}
}

func dataSourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	email := strings.TrimSuffix(d.Get(names.AttrEmail).(string), ".")
	_, err := findIdentityNotificationAttributesByIdentity(ctx, conn, email)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Email Identity (%s) verification: %s", email, err)
	}

	d.SetId(email)
	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Resource:  fmt.Sprintf("identity/%s", email),
		Service:   "ses",
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrEmail, email)

	return diags
}
