// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontOriginAccessIdentityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	dataSourceName := "data.aws_cloudfront_origin_access_identity.test"
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentityDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessIdentityExists(ctx, resourceName, &origin),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_arn", resourceName, "iam_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrComment, resourceName, names.AttrComment),
					resource.TestCheckResourceAttrPair(dataSourceName, "caller_reference", resourceName, "caller_reference"),
					resource.TestCheckResourceAttrPair(dataSourceName, "s3_canonical_user_id", resourceName, "s3_canonical_user_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cloudfront_access_identity_path", resourceName, "cloudfront_access_identity_path"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
				),
			},
		},
	})
}

const testAccOriginAccessIdentityDataSourceConfig_basic = `
resource "aws_cloudfront_origin_access_identity" "test" {
  comment = "some comment"
}

data "aws_cloudfront_origin_access_identity" "test" {
  id = aws_cloudfront_origin_access_identity.test.id
}
`
