// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_canonical_user_id", name="Canonical User ID")
func dataSourceCanonicalUserID() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCanonicalUserIDRead,

		Schema: map[string]*schema.Schema{
			names.AttrDisplayName: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCanonicalUserIDRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	output, err := conn.ListBuckets(ctx, &s3.ListBucketsInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing S3 Buckets: %s", err)
	}

	if output == nil || output.Owner == nil {
		return sdkdiag.AppendErrorf(diags, "S3 Canonical User ID not found")
	}

	d.SetId(aws.ToString(output.Owner.ID))
	d.Set(names.AttrDisplayName, output.Owner.DisplayName)

	return diags
}
