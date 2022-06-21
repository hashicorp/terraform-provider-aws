package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFrontCachePolicyDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSource1Name := "data.aws_cloudfront_cache_policy.by_id"
	dataSource2Name := "data.aws_cloudfront_cache_policy.by_name"
	resourceName := "aws_cloudfront_cache_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachePolicyDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, "comment", resourceName, "comment"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "default_ttl", resourceName, "default_ttl"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "min_ttl", resourceName, "min_ttl"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "max_ttl", resourceName, "max_ttl"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#"),

					resource.TestCheckResourceAttrPair(dataSource2Name, "comment", resourceName, "comment"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "default_ttl", resourceName, "default_ttl"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "min_ttl", resourceName, "min_ttl"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "max_ttl", resourceName, "max_ttl"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#"),
				),
			},
		},
	})
}

func testAccCachePolicyDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_cloudfront_cache_policy" "by_name" {
  name = aws_cloudfront_cache_policy.test.name
}

data "aws_cloudfront_cache_policy" "by_id" {
  id = aws_cloudfront_cache_policy.test.id
}

resource "aws_cloudfront_cache_policy" "test" {
  name        = %[1]q
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1

  parameters_in_cache_key_and_forwarded_to_origin {
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
}
`, rName)
}
