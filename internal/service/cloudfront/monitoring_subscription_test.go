package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudFrontMonitoringSubscription_basic(t *testing.T) {
	var v cloudfront.GetMonitoringSubscriptionOutput
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMonitoringSubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringSubscriptionConfig_basic("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.0.realtime_metrics_subscription_status", "Enabled"),
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

func TestAccCloudFrontMonitoringSubscription_disappears(t *testing.T) {
	var v cloudfront.GetMonitoringSubscriptionOutput
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMonitoringSubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringSubscriptionConfig_basic("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringSubscriptionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceMonitoringSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontMonitoringSubscription_update(t *testing.T) {
	var v cloudfront.GetMonitoringSubscriptionOutput
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMonitoringSubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringSubscriptionConfig_basic("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.0.realtime_metrics_subscription_status", "Enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitoringSubscriptionConfig_basic("Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.0.realtime_metrics_subscription_status", "Disabled"),
				),
			},
		},
	})
}

func testAccCheckMonitoringSubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_monitoring_subscription" {
			continue
		}

		_, err := tfcloudfront.FindMonitoringSubscriptionByDistributionID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudFront Monitoring Subscription %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckMonitoringSubscriptionExists(n string, v *cloudfront.GetMonitoringSubscriptionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudFront Monitoring Subscription ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		output, err := tfcloudfront.FindMonitoringSubscriptionByDistributionID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccMonitoringSubscriptionBaseConfig() string {
	return `
resource "aws_cloudfront_distribution" "test" {
  enabled          = true
  retain_on_delete = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`
}

func testAccMonitoringSubscriptionConfig_basic(status string) string {
	return acctest.ConfigCompose(
		testAccMonitoringSubscriptionBaseConfig(),
		fmt.Sprintf(`
resource "aws_cloudfront_monitoring_subscription" "test" {
  distribution_id = aws_cloudfront_distribution.test.id

  monitoring_subscription {
    realtime_metrics_subscription_config {
      realtime_metrics_subscription_status = %[1]q
    }
  }
}
`, status))
}
