// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontLogDeliveryCanonicalUserIDDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_log_delivery_canonical_user_id.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryCanonicalUserIdDataSourceConfig_basic(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0"),
				),
			},
		},
	})
}

func TestAccCloudFrontLogDeliveryCanonicalUserIDDataSource_default(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_log_delivery_canonical_user_id.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryCanonicalUserIdDataSourceConfig_basic(names.USWest2RegionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0"),
				),
			},
		},
	})
}

func TestAccCloudFrontLogDeliveryCanonicalUserIDDataSource_cn(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_log_delivery_canonical_user_id.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryCanonicalUserIdDataSourceConfig_basic(names.CNNorthwest1RegionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, "a52cb28745c0c06e84ec548334e44bfa7fc2a85c54af20cd59e4969344b7af56"),
				),
			},
		},
	})
}

func testAccLogDeliveryCanonicalUserIdDataSourceConfig_basic(region string) string {
	if region == "" {
		region = "null"
	}

	return fmt.Sprintf(`
data "aws_cloudfront_log_delivery_canonical_user_id" "test" {
  region = %[1]q
}
`, region)
}
