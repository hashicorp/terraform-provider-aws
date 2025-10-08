// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_s3_bucket_policy", name="Bucket Policy")
func dataSourceBucketPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBucketPolicyRead,

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	policy, err := findBucketPolicy(ctx, conn, bucket)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) Policy: %s", bucket, err)
	}

	policy, err = structure.NormalizeJsonString(policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(bucket)
	d.Set(names.AttrPolicy, policy)

	return diags
}
