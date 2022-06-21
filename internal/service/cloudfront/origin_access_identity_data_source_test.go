package cloudfront_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFrontOriginAccessIdentityDataSource_basic(t *testing.T) {
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	dataSourceName := "data.aws_cloudfront_origin_access_identity.test"
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentityDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessIdentityExistence(resourceName, &origin),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_arn", resourceName, "iam_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "comment", resourceName, "comment"),
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
