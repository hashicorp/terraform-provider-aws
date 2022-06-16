package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudFrontCachePolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_cache_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "comment", ""),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", "86400"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", "31536000"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudFrontCachePolicy_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_cache_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceCachePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontCachePolicy_Items(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_cache_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachePolicyConfig_items(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "comment", "test comment"),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", "50"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", "1"),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", "100"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", "allExcept"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.*", "test1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.*", "test2"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.0.items.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.0.items.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.0.items.*", "test2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.0.items.*", "test3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCachePolicyConfig_itemsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "comment", "test comment updated"),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", "51"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", "10"),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", "99"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.*", "test2"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.*", "test1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", "all"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", "0"),
				),
			},
		},
	})
}

func TestAccCloudFrontCachePolicy_ZeroTTLs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_cache_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachePolicyConfig_zeroTTLs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "comment", ""),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCachePolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_cache_policy" {
			continue
		}

		_, err := tfcloudfront.FindCachePolicyByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudFront Cache Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCachePolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudFront Cache Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		_, err := tfcloudfront.FindCachePolicyByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCachePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
  name = %[1]q

  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "none"
    }

    headers_config {
      header_behavior = "none"
    }

    query_strings_config {
      query_string_behavior = "none"
    }
  }
}
`, rName)
}

func testAccCachePolicyConfig_items(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
  name        = %[1]q
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1

  parameters_in_cache_key_and_forwarded_to_origin {
    enable_accept_encoding_brotli = true
    enable_accept_encoding_gzip   = true

    cookies_config {
      cookie_behavior = "allExcept"

      cookies {
        items = ["test1"]
      }
    }

    headers_config {
      header_behavior = "whitelist"

      headers {
        items = ["test1", "test2"]
      }
    }

    query_strings_config {
      query_string_behavior = "whitelist"

      query_strings {
        items = ["test1", "test2", "test3"]
      }
    }
  }
}
`, rName)
}

func testAccCachePolicyConfig_itemsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
  name        = %[1]q
  comment     = "test comment updated"
  default_ttl = 51
  max_ttl     = 99
  min_ttl     = 10

  parameters_in_cache_key_and_forwarded_to_origin {
    enable_accept_encoding_brotli = false
    enable_accept_encoding_gzip   = false

    cookies_config {
      cookie_behavior = "whitelist"

      cookies {
        items = ["test2", "test1"]
      }
    }

    headers_config {
      header_behavior = "whitelist"

      headers {
        items = ["test1"]
      }
    }

    query_strings_config {
      query_string_behavior = "all"
    }
  }
}
`, rName)
}

func testAccCachePolicyConfig_zeroTTLs(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
  name = %[1]q

  default_ttl = 0
  max_ttl     = 0
  min_ttl     = 0

  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "none"
    }

    headers_config {
      header_behavior = "none"
    }

    query_strings_config {
      query_string_behavior = "none"
    }
  }
}
`, rName)
}
