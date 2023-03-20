package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailDistribution_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lightsail.EndpointsID)
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_basic(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "alternative_domain_names.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", "small_1_0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.allowed_http_methods", "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.cached_http_methods", "GET,HEAD"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.default_ttl", "86400"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.cookies_allow_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.option", "none"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.headers_allow_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.option", "default"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.query_strings_allowed_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.option", "false"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.maximum_ttl", "31536000"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.minimum_ttl", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.behavior", "cache"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "dualstack"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.region_name"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.name", bucketName),
					resource.TestCheckResourceAttrSet(resourceName, "origin.0.region_name"),
					resource.TestCheckResourceAttrSet(resourceName, "origin.0.resource_type"),
					resource.TestCheckResourceAttrSet(resourceName, "origin_public_dns"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_type"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "support_code"),
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

func TestAccLightsailDistribution_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lightsail.EndpointsID)
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_basic(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceDistribution(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDistributionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_distribution" {
				continue
			}

			_, err := tflightsail.FindDistributionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResNameDistribution, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDistributionExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Lightsail, create.ErrActionCheckingExistence, tflightsail.ResNameDistribution, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Lightsail, create.ErrActionCheckingExistence, tflightsail.ResNameDistribution, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()
		resp, err := tflightsail.FindDistributionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("Distribution %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDistributionConfig_basic(rName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_bucket" "test" {
  name      = %[2]q
  bundle_id = "small_1_0"
}
resource "aws_lightsail_distribution" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    forwarded_cookies {
      cookies_allow_list = []
    }
    forwarded_headers {
      headers_allow_list = []
    }
    forwarded_query_strings {
      query_strings_allowed_list = []
    }
  }
}
`, rName, bucketName)
}
