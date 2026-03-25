// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlAccessPointDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	accessPointName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"
	dataSourceName := "data.aws_s3_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointDataSourceConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAlias, dataSourceName, names.AttrAlias),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, dataSourceName, names.AttrBucket),
					resource.TestCheckResourceAttrPair(resourceName, "bucket_account_id", dataSourceName, "bucket_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, "network_origin", dataSourceName, "network_origin"),
				),
			},
		},
	})
}

func testAccAccessPointDataSourceConfig_basic(bucketName, accessPointName string) string {
	return acctest.ConfigCompose(testAccAccessPointConfig_tags1(bucketName, accessPointName, acctest.CtKey1, acctest.CtValue1), `
data "aws_s3_access_point" "test" {
  name = aws_s3_access_point.test.name
}
`)
}
