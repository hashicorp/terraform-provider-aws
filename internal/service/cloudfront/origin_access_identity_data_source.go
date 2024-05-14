// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudfront_origin_access_identity", name="Origin Access Identity")
func dataSourceOriginAccessIdentity() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOriginAccessIdentityRead,

		Schema: map[string]*schema.Schema{
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_access_identity_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrComment: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
			"s3_canonical_user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceOriginAccessIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	id := d.Get(names.AttrID).(string)
	output, err := findOriginAccessIdentityByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Origin Access Identity (%s): %s", id, err)
	}

	apiObject := output.CloudFrontOriginAccessIdentity.CloudFrontOriginAccessIdentityConfig
	d.SetId(aws.ToString(output.CloudFrontOriginAccessIdentity.Id))
	d.Set("caller_reference", apiObject.CallerReference)
	d.Set("cloudfront_access_identity_path", "origin-access-identity/cloudfront/"+d.Id())
	d.Set(names.AttrComment, apiObject.Comment)
	d.Set("etag", output.ETag)
	d.Set("iam_arn", originAccessIdentityARN(meta.(*conns.AWSClient), d.Id()))
	d.Set("s3_canonical_user_id", output.CloudFrontOriginAccessIdentity.S3CanonicalUserId)

	return diags
}
