// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_cloudfront_origin_access_identities")
func DataSourceOriginAccessIdentities() *schema.Resource {
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
			"ids": {
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

	var comments []interface{}

	if v, ok := d.GetOk("comments"); ok && v.(*schema.Set).Len() > 0 {
		comments = v.(*schema.Set).List()
	}
	var output []*awstypes.CloudFrontOriginAccessIdentitySummary

	input := &cloudfront.ListCloudFrontOriginAccessIdentitiesInput{}

	pages := cloudfront.NewListCloudFrontOriginAccessIdentitiesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing CloudFront origin access identities: %s", err)
		}
		comments = append(comments, page)
	}

	var iamARNs, ids, s3CanonicalUserIDs []string

	for _, v := range output {
		// See https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html#private-content-updating-s3-bucket-policies-principal.
		iamARN := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "iam",
			AccountID: "cloudfront",
			Resource:  fmt.Sprintf("user/CloudFront Origin Access Identity %s", *v.Id),
		}.String()
		iamARNs = append(iamARNs, iamARN)
		ids = append(ids, aws.ToString(v.Id))
		s3CanonicalUserIDs = append(s3CanonicalUserIDs, aws.ToString(v.S3CanonicalUserId))
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)
	d.Set("iam_arns", iamARNs)
	d.Set("ids", ids)
	d.Set("s3_canonical_user_ids", s3CanonicalUserIDs)

	return diags
}
