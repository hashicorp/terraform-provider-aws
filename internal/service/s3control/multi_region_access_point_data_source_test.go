// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlMultiRegionAccessPointDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_multi_region_access_point.test"
	dataSourceName := "data.aws_s3control_multi_region_access_point.test"
	bucket1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucket2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointDataSourceConfig_basic(bucket1Name, bucket2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAlias, dataSourceName, names.AttrAlias),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, dataSourceName, names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.name", dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.public_access_block.0.block_public_acls", dataSourceName, "public_access_block.0.block_public_acls"),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.public_access_block.0.block_public_policy", dataSourceName, "public_access_block.0.block_public_policy"),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.public_access_block.0.ignore_public_acls", dataSourceName, "public_access_block.0.ignore_public_acls"),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.public_access_block.0.restrict_public_buckets", dataSourceName, "public_access_block.0.restrict_public_buckets"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "details.0.region.*", map[string]string{
						names.AttrBucket: bucket1Name,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "details.0.region.*", map[string]string{
						names.AttrBucket: bucket2Name,
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceName, names.AttrStatus),
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
