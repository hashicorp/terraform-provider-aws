// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontOriginAccessControlDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_origin_access_control.this"
	resourceName := "aws_cloudfront_origin_access_control.this"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "origin_access_control_origin_type", resourceName, "origin_access_control_origin_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_behavior", resourceName, "signing_behavior"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_protocol", resourceName, "signing_protocol"),
				),
			},
		},
	})
}

func testAccOriginAccessControlDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "this" {
  name                              = %[1]q
  description                       = %[1]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

data "aws_cloudfront_origin_access_control" "this" {
  id = aws_cloudfront_origin_access_control.this.id
}
`, rName)
}
