package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFrontOriginAccessIdentitiesDataSource_comments(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_origin_access_identities.test"
	resourceName := "aws_cloudfront_origin_access_identity.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentitiesDataSourceConfig_comments(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "iam_arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "s3_canonical_user_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "iam_arns.*", resourceName, "iam_arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ids.*", resourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "s3_canonical_user_ids.*", resourceName, "s3_canonical_user_id"),
				),
			},
		},
	})
}

func TestAccCloudFrontOriginAccessIdentitiesDataSource_all(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_origin_access_identities.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentitiesDataSourceConfig_noComments(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "iam_arns.#", "1"),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "1"),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "s3_canonical_user_ids.#", "1"),
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
