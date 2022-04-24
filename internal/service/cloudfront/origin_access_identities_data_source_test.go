package cloudfront_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFrontOriginAccessIdentitiesDataSource_With_Filter(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_origin_access_identities.test"
	resource2Name := "aws_cloudfront_origin_access_identity.test.1"
	rCount := strconv.Itoa(sdkacctest.RandIntRange(2, 5))
	fCount := "1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentitiesBasicDataSourceConfigFilter(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", fCount),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", fCount),
					resource.TestCheckResourceAttr(dataSourceName, "s3_canonical_user_ids.#", fCount),
					resource.TestCheckResourceAttrPair(dataSourceName, "arns.0", resource2Name, "iam_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", resource2Name, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "s3_canonical_user_ids.0", resource2Name, "s3_canonical_user_id"),
				),
			},
		},
	})
}

func TestAccCloudFrontOriginAccessIdentitiesDataSource_No_Filter(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_origin_access_identities.test"
	resource1Name := "aws_cloudfront_origin_access_identity.test.0"
	rCount := strconv.Itoa(sdkacctest.RandIntRange(1, 4))
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentitiesBasicDataSourceConfigNoFilter(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "s3_canonical_user_ids.#", rCount),
					resource.TestCheckResourceAttrPair(dataSourceName, "arns.0", resource1Name, "iam_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", resource1Name, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "s3_canonical_user_ids.0", resource1Name, "s3_canonical_user_id"),
				),
			},
		},
	})
}

func testAccOriginAccessIdentitiesBasicDataSourceConfigFilter(rCount, rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_identity" "test" {
  count = %[1]s
  comment  = "%[2]s-${count.index}-comment"
}

data "aws_cloudfront_origin_access_identities" "test" {
	filter{
		name = "comment"
		values = [aws_cloudfront_origin_access_identity.test.1.comment]

	}
}
`, rCount, rName)
}

func testAccOriginAccessIdentitiesBasicDataSourceConfigNoFilter(rCount, rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_identity" "test" {
  count = %[1]s
  comment  = "%[2]s-${count.index}-comment"
}

data "aws_cloudfront_origin_access_identities" "test" {

}
`, rCount, rName)
}
