// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudfront_origin_access_identities", name="Origin Access Identities")
func dataSourceOriginAccessIdentities() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOriginAccessIdentitiesRead,

		Schema: map[string]*schema.Schema{
			"comments": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"iam_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"s3_canonical_user_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceOriginAccessIdentitiesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	var comments []any
	if v, ok := d.GetOk("comments"); ok && v.(*schema.Set).Len() > 0 {
		comments = v.(*schema.Set).List()
	}

	input := &cloudfront.ListCloudFrontOriginAccessIdentitiesInput{}
	var output []awstypes.CloudFrontOriginAccessIdentitySummary

	pages := cloudfront.NewListCloudFrontOriginAccessIdentitiesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing CloudFront Origin Access Identities: %s", err)
		}

		for _, v := range page.CloudFrontOriginAccessIdentityList.Items {
			if len(comments) > 0 {
				if idx := tfslices.IndexOf(comments, aws.ToString(v.Comment)); idx == -1 {
					continue
				}
			}

			output = append(output, v)
		}
	}

	var iamARNs, ids, s3CanonicalUserIDs []string

	for _, v := range output {
		id := aws.ToString(v.Id)
		iamARNs = append(iamARNs, originAccessIdentityARN(meta.(*conns.AWSClient), id))
		ids = append(ids, id)
		s3CanonicalUserIDs = append(s3CanonicalUserIDs, aws.ToString(v.S3CanonicalUserId))
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)
	d.Set("iam_arns", iamARNs)
	d.Set(names.AttrIDs, ids)
	d.Set("s3_canonical_user_ids", s3CanonicalUserIDs)

	return diags
}
