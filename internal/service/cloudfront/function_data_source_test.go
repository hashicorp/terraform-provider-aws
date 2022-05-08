package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFrontFunctionDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_function.test"
	resourceName := "aws_cloudfront_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionBasicDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "code", resourceName, "code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "comment", resourceName, "comment"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "runtime", resourceName, "runtime"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
				),
			},
		},
	})
}

func testAccFunctionBasicDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  comment = "test"
  code    = <<-EOT
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

data "aws_cloudfront_function" "test" {
  name  = aws_cloudfront_function.test.name
  stage = "LIVE"
}
`, rName)
}
