// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontMonitoringSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudfront.GetMonitoringSubscriptionOutput
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitoringSubscriptionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringSubscriptionConfig_basic("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var v cloudfront.GetMonitoringSubscriptionOutput
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitoringSubscriptionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringSubscriptionConfig_basic("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringSubscriptionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceMonitoringSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontMonitoringSubscription_update(t *testing.T) {
	acctest.Skip(t, "MonitoringSubscriptionAlreadyExists: A monitoring subscription already exists for this distribution")

	ctx := acctest.Context(t)
	var v cloudfront.GetMonitoringSubscriptionOutput
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitoringSubscriptionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringSubscriptionConfig_basic("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", acctest.Ct1),
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
					testAccCheckMonitoringSubscriptionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.0.realtime_metrics_subscription_status", "Disabled"),
				),
			},
		},
	})
}

func testAccCheckMonitoringSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_monitoring_subscription" {
				continue
			}

			_, err := tfcloudfront.FindMonitoringSubscriptionByDistributionID(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckMonitoringSubscriptionExists(ctx context.Context, n string, v *cloudfront.GetMonitoringSubscriptionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindMonitoringSubscriptionByDistributionID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccMonitoringSubscriptionConfig_base() string {
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
	return acctest.ConfigCompose(testAccMonitoringSubscriptionConfig_base(), fmt.Sprintf(`
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
