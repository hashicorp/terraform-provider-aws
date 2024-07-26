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

func TestAccCloudFrontFunctionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_function.test"
	resourceName := "aws_cloudfront_function.test"
	keyValueStorerName := "aws_cloudfront_key_value_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "code", resourceName, "code"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrComment, resourceName, names.AttrComment),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "runtime", resourceName, "runtime"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, "key_value_store_associations.0", keyValueStorerName, names.AttrARN),
				),
			},
		},
	})
}

func testAccFunctionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name                         = %[1]q
  runtime                      = "cloudfront-js-2.0"
  comment                      = "test"
  key_value_store_associations = [aws_cloudfront_key_value_store.test.arn]
  code                         = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT
}

resource "aws_cloudfront_key_value_store" "test" {
  name = %[1]q
}

data "aws_cloudfront_function" "test" {
  name  = aws_cloudfront_function.test.name
  stage = "LIVE"
}
`, rName)
}
