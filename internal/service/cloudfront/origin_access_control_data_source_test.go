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
)

func TestAccCloudFrontOriginAccessControlDataSource_description(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudfront_origin_access_control.test"
	resourceName := "aws_cloudfront_origin_access_control.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlDataSourceConfig_description(rName, "Acceptance Test 1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "etag"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "origin_access_control_origin_type", resourceName, "origin_access_control_origin_type"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "signing_behavior", resourceName, "signing_behavior"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "signing_protocol", resourceName, "signing_protocol"),
				),
			},
		},
	})
}

func testAccOriginAccessControlDataSourceConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test1" {
  name                              = %[1]q
  description                       = %[2]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

data "aws_cloudfront_origin_access_control" "test" {
	id = aws_cloudfront_origin_access_control.test1.id
}
`, rName, description)
}
