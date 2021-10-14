package cloudfront_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)





func TestAccCloudFrontMonitoringSubscription_basic(t *testing.T) {
	var v cloudfront.MonitoringSubscription
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontMonitoringSubscriptionDestroy,
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringSubscriptionConfig("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontMonitoringSubscriptionExists(resourceName, &v),
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
	var v cloudfront.MonitoringSubscription
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontMonitoringSubscriptionDestroy,
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringSubscriptionConfig("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontMonitoringSubscriptionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceMonitoringSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontMonitoringSubscription_update(t *testing.T) {
	var v cloudfront.MonitoringSubscription
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontMonitoringSubscriptionDestroy,
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringSubscriptionConfig("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontMonitoringSubscriptionExists(resourceName, &v),
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
				Config: testAccMonitoringSubscriptionConfig("Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontMonitoringSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.0.realtime_metrics_subscription_status", "Disabled"),
				),
			},
		},
	})
}

func testAccCheckCloudFrontMonitoringSubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_monitoring_subscription" {
			continue
		}

		s, err := tfcloudfront.FindMonitoringSubscriptionByDistributionID(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, cloudfront.ErrCodeNoSuchDistribution, "") {
			continue
		}
		if err != nil {
			return err
		}
		if s != nil {
			continue
		}
		return fmt.Errorf("CloudFront Monitoring Subscription still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCloudFrontMonitoringSubscriptionExists(n string, v *cloudfront.MonitoringSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudFront Monitoring Subscription ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn
		out, err := tfcloudfront.FindMonitoringSubscriptionByDistributionID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

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

func testAccMonitoringSubscriptionConfig(status string) string {
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
