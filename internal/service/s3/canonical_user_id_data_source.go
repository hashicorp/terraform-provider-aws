// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_canonical_user_id")
func DataSourceCanonicalUserID() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCanonicalUserIDRead,

		Schema: map[string]*schema.Schema{
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCanonicalUserIDRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	log.Printf("[DEBUG] Reading S3 Buckets")

	req := &s3.ListBucketsInput{}
	resp, err := conn.ListBucketsWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing S3 Buckets: %s", err)
	}
	if resp == nil || resp.Owner == nil {
		return sdkdiag.AppendErrorf(diags, "no canonical user ID found")
	}

	d.SetId(aws.StringValue(resp.Owner.ID))
	d.Set("display_name", resp.Owner.DisplayName)

	return diags
}
