// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontOriginAccessIdentitiesDataSource_comments(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_origin_access_identities.test"
	resourceName := "aws_cloudfront_origin_access_identity.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentitiesDataSourceConfig_comments(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "iam_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "s3_canonical_user_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "iam_arns.*", resourceName, "iam_arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ids.*", resourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "s3_canonical_user_ids.*", resourceName, "s3_canonical_user_id"),
				),
			},
		},
	})
}

func TestAccCloudFrontOriginAccessIdentitiesDataSource_all(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_origin_access_identities.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentitiesDataSourceConfig_noComments(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "iam_arns.#", 1),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", 1),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "s3_canonical_user_ids.#", 1),
				),
			},
		},
	})
}

func testAccOriginAccessIdentitiesDataSourceConfig_comments(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_identity" "test1" {
  comment = "%[1]s-1-comment"
}

resource "aws_cloudfront_origin_access_identity" "test2" {
  comment = "%[1]s-2-comment"
}

data "aws_cloudfront_origin_access_identities" "test" {
  comments = ["%[1]s-1-comment"]

  depends_on = [aws_cloudfront_origin_access_identity.test1, aws_cloudfront_origin_access_identity.test2]
}
`, rName)
}

func testAccOriginAccessIdentitiesDataSourceConfig_noComments(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_identity" "test1" {
  comment = "%[1]s-1-comment"
}

resource "aws_cloudfront_origin_access_identity" "test2" {
  comment = "%[1]s-2-comment"
}

data "aws_cloudfront_origin_access_identities" "test" {
  depends_on = [aws_cloudfront_origin_access_identity.test1, aws_cloudfront_origin_access_identity.test2]
}
`, rName)
}
