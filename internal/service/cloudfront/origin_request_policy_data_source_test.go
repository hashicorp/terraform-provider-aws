package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFrontOriginRequestPolicyDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSource1Name := "data.aws_cloudfront_origin_request_policy.by_id"
	dataSource2Name := "data.aws_cloudfront_origin_request_policy.by_name"
	resourceName := "aws_cloudfront_origin_request_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginRequestPolicyDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, "comment", resourceName, "comment"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cookies_config.#", resourceName, "cookies_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cookies_config.0.cookie_behavior", resourceName, "cookies_config.0.cookie_behavior"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cookies_config.0.cookies.#", resourceName, "cookies_config.0.cookies.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "headers_config.#", resourceName, "headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "headers_config.0.header_behavior", resourceName, "headers_config.0.header_behavior"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "headers_config.0.headers.#", resourceName, "headers_config.0.headers.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "query_strings_config.#", resourceName, "query_strings_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "query_strings_config.0.query_string_behavior", resourceName, "query_strings_config.0.query_string_behavior"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "query_strings_config.0.query_strings.#", resourceName, "query_strings_config.0.query_strings.#"),

					resource.TestCheckResourceAttrPair(dataSource2Name, "comment", resourceName, "comment"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cookies_config.#", resourceName, "cookies_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cookies_config.0.cookie_behavior", resourceName, "cookies_config.0.cookie_behavior"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cookies_config.0.cookies.#", resourceName, "cookies_config.0.cookies.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "headers_config.#", resourceName, "headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "headers_config.0.header_behavior", resourceName, "headers_config.0.header_behavior"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "headers_config.0.headers.#", resourceName, "headers_config.0.headers.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "query_strings_config.#", resourceName, "query_strings_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "query_strings_config.0.query_string_behavior", resourceName, "query_strings_config.0.query_string_behavior"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "query_strings_config.0.query_strings.#", resourceName, "query_strings_config.0.query_strings.#"),
				),
			},
		},
	})
}

func testAccOriginRequestPolicyDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_cloudfront_origin_request_policy" "by_name" {
  name = aws_cloudfront_origin_request_policy.test.name
}

data "aws_cloudfront_origin_request_policy" "by_id" {
  id = aws_cloudfront_origin_request_policy.test.id
}

resource "aws_cloudfront_origin_request_policy" "test" {
  name    = %[1]q
  comment = "test comment"

  cookies_config {
    cookie_behavior = "whitelist"

    cookies {
      items = ["test"]
    }
  }

  headers_config {
    header_behavior = "whitelist"

    headers {
      items = ["test"]
    }
  }

  query_strings_config {
    query_string_behavior = "whitelist"

    query_strings {
      items = ["test"]
    }
  }
}
`, rName)
}
