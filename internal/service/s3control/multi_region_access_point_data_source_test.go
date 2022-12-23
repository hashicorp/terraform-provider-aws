package s3control_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/s3control"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3ControlMultiRegionAccessPointDataSource_basic(t *testing.T) {
	resourceName := "aws_s3control_multi_region_access_point.test"
	dataSourceName := "data.aws_s3control_multi_region_access_point.test"

	bucket1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucket2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 2),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointDataSourceConfig_basic(bucket1Name, bucket2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceName, "account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "alias", dataSourceName, "alias"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", dataSourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.public_access_block.0.block_public_acls", dataSourceName, "public_access_block.0.block_public_acls"),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.public_access_block.0.block_public_policy", dataSourceName, "public_access_block.0.block_public_policy"),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.public_access_block.0.ignore_public_acls", dataSourceName, "public_access_block.0.ignore_public_acls"),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.public_access_block.0.restrict_public_buckets", dataSourceName, "public_access_block.0.restrict_public_buckets"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "details.0.region.*", map[string]string{
						"bucket": bucket1Name,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "details.0.region.*", map[string]string{
						"bucket": bucket2Name,
					}),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
				),
			},
		},
	})
}

func testAccMultiRegionAccessPointDataSource_base(bucket1Name string, bucket2Name string, rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test1" {
  provider = aws

  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  provider = awsalternate

  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  provider = aws

  details {
    name = %[3]q

    region {
      bucket = aws_s3_bucket.test1.id
    }

    region {
      bucket = aws_s3_bucket.test2.id
    }

    public_access_block {
      block_public_acls       = false
      block_public_policy     = false
      ignore_public_acls      = false
      restrict_public_buckets = false
    }
  }
}
`, bucket1Name, bucket2Name, rName))
}

func testAccMultiRegionAccessPointDataSourceConfig_basic(bucket1Name string, bucket2Name string, rName string) string {
	return acctest.ConfigCompose(testAccMultiRegionAccessPointDataSource_base(bucket1Name, bucket2Name, rName), fmt.Sprintf(`
data "aws_s3control_multi_region_access_point" "test" {
  provider = aws

  name = %[1]q

  depends_on = [aws_s3control_multi_region_access_point.test]
}
`, rName))
}
