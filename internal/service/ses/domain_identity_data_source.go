// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func dataSourceDomainIdentityRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	domainName := d.Get(names.AttrDomain).(string)
	verificationAttrs, err := findIdentityVerificationAttributesByIdentity(ctx, conn, domainName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain Identity (%s) verification: %s", domainName, err)
	}

	d.SetId(domainName)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("identity/%s", domainName),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDomain, domainName)
	d.Set("verification_token", verificationAttrs.VerificationToken)

	return diags
}
