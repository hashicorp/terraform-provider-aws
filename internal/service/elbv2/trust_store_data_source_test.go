// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2TrustStoreDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceNameByName := "data.aws_lb_trust_store.named"
	datasourceNameByArn := "data.aws_lb_trust_store.with_arn"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{

			{
				Config: testAccTrustStoreDataSourceConfig_withName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(datasourceNameByName, names.AttrARN),
				),
			},
			{
				Config: testAccTrustStoreDataSourceConfig_withARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceNameByArn, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(datasourceNameByArn, names.AttrARN),
				),
			},
		},
	})
}

func testAccTrustStoreDataSourceConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_baseS3BucketCA(rName), fmt.Sprintf(`
resource "aws_lb_trust_store" "test" {
  name                             = %[1]q
  ca_certificates_bundle_s3_bucket = aws_s3_bucket.test.bucket
  ca_certificates_bundle_s3_key    = aws_s3_object.test.key
}
`, rName))
}

func testAccTrustStoreDataSourceConfig_withName(rName string) string {
	return acctest.ConfigCompose(testAccTrustStoreDataSourceConfig_base(rName), fmt.Sprintf(`
data "aws_lb_trust_store" "named" {
  name       = %[1]q
  depends_on = [aws_lb_trust_store.test]
}
`, rName))
}

func testAccTrustStoreDataSourceConfig_withARN(rName string) string {
	return acctest.ConfigCompose(testAccTrustStoreDataSourceConfig_base(rName), `
data "aws_lb_trust_store" "with_arn" {
  arn        = aws_lb_trust_store.test.arn
  depends_on = [aws_lb_trust_store.test]
}
`)
}
