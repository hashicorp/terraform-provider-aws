package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFrontResponseHeadersPolicyDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSource1Name := "data.aws_cloudfront_response_headers_policy.by_id"
	dataSource2Name := "data.aws_cloudfront_response_headers_policy.by_name"
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResponseHeadersPolicyDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, "comment", resourceName, "comment"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.#", resourceName, "cors_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_allow_credentials", resourceName, "cors_config.0.access_control_allow_credentials"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_allow_headers.#", resourceName, "cors_config.0.access_control_allow_headers.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_allow_headers.0.items.#", resourceName, "cors_config.0.access_control_allow_headers.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_allow_methods.#", resourceName, "cors_config.0.access_control_allow_methods.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_allow_methods.0.items.#", resourceName, "cors_config.0.access_control_allow_methods.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_allow_origins.#", resourceName, "cors_config.0.access_control_allow_origins.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_allow_origins.0.items.#", resourceName, "cors_config.0.access_control_allow_origins.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_expose_headers.#", resourceName, "cors_config.0.access_control_expose_headers.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_expose_headers.0.items.#", resourceName, "cors_config.0.access_control_expose_headers.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.access_control_max_age_sec", resourceName, "cors_config.0.access_control_max_age_sec"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "cors_config.0.origin_override", resourceName, "cors_config.0.origin_override"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "custom_headers_config.#", resourceName, "custom_headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "custom_headers_config.0.items.#", resourceName, "custom_headers_config.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "security_headers_config.#", resourceName, "security_headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "security_headers_config.0.content_security_policy.#", resourceName, "security_headers_config.0.content_security_policy.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "security_headers_config.0.frame_options.#", resourceName, "security_headers_config.0.frame_options.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "security_headers_config.0.referrer_policy.#", resourceName, "security_headers_config.0.referrer_policy.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "security_headers_config.0.strict_transport_security.#", resourceName, "security_headers_config.0.strict_transport_security.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "security_headers_config.0.xss_protection.#", resourceName, "security_headers_config.0.xss_protection.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "server_timing_headers_config.#", resourceName, "server_timing_headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "server_timing_headers_config.0.enabled", resourceName, "server_timing_headers_config.0.enabled"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "server_timing_headers_config.0.sampling_rate", resourceName, "server_timing_headers_config.0.sampling_rate"),

					resource.TestCheckResourceAttrPair(dataSource2Name, "comment", resourceName, "comment"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.#", resourceName, "cors_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_allow_credentials", resourceName, "cors_config.0.access_control_allow_credentials"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_allow_headers.#", resourceName, "cors_config.0.access_control_allow_headers.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_allow_headers.0.items.#", resourceName, "cors_config.0.access_control_allow_headers.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_allow_methods.#", resourceName, "cors_config.0.access_control_allow_methods.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_allow_methods.0.items.#", resourceName, "cors_config.0.access_control_allow_methods.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_allow_origins.#", resourceName, "cors_config.0.access_control_allow_origins.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_allow_origins.0.items.#", resourceName, "cors_config.0.access_control_allow_origins.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_expose_headers.#", resourceName, "cors_config.0.access_control_expose_headers.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_expose_headers.0.items.#", resourceName, "cors_config.0.access_control_expose_headers.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.access_control_max_age_sec", resourceName, "cors_config.0.access_control_max_age_sec"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "cors_config.0.origin_override", resourceName, "cors_config.0.origin_override"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "custom_headers_config.#", resourceName, "custom_headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "custom_headers_config.0.items.#", resourceName, "custom_headers_config.0.items.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "security_headers_config.#", resourceName, "security_headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "security_headers_config.0.content_security_policy.#", resourceName, "security_headers_config.0.content_security_policy.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "security_headers_config.0.frame_options.#", resourceName, "security_headers_config.0.frame_options.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "security_headers_config.0.referrer_policy.#", resourceName, "security_headers_config.0.referrer_policy.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "security_headers_config.0.strict_transport_security.#", resourceName, "security_headers_config.0.strict_transport_security.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "security_headers_config.0.xss_protection.#", resourceName, "security_headers_config.0.xss_protection.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "server_timing_headers_config.#", resourceName, "server_timing_headers_config.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "server_timing_headers_config.0.enabled", resourceName, "server_timing_headers_config.0.enabled"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "server_timing_headers_config.0.sampling_rate", resourceName, "server_timing_headers_config.0.sampling_rate"),
				),
			},
		},
	})
}

func testAccResponseHeadersPolicyDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_cloudfront_response_headers_policy" "by_name" {
  name = aws_cloudfront_response_headers_policy.test.name
}

data "aws_cloudfront_response_headers_policy" "by_id" {
  id = aws_cloudfront_response_headers_policy.test.id
}

resource "aws_cloudfront_response_headers_policy" "test" {
  name    = %[1]q
  comment = "test comment"

  cors_config {
    access_control_allow_credentials = true

    access_control_allow_headers {
      items = ["test"]
    }

    access_control_allow_methods {
      items = ["GET"]
    }

    access_control_allow_origins {
      items = ["test.example.comtest"]
    }

    origin_override = true
  }

  custom_headers_config {
    items {
      header   = "X-Header2"
      override = false
      value    = "value2"
    }

    items {
      header   = "X-Header1"
      override = true
      value    = "value1"
    }
  }

  security_headers_config {
    content_security_policy {
      content_security_policy = "policy1"
      override                = true
    }

    frame_options {
      frame_option = "DENY"
      override     = false
    }

    strict_transport_security {
      access_control_max_age_sec = 90
      override                   = true
      preload                    = true
    }
  }

  server_timing_headers_config {
    enabled       = true
    sampling_rate = 10
  }
}
`, rName)
}
