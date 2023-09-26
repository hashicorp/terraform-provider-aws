// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfshield "github.com/hashicorp/terraform-provider-aws/internal/service/shield"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccShieldApplicationLayerAutomaticResponse_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var applicationlayerautomaticresponse shield.DescribeProtectionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_application_layer_automatic_response.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckWAFV2CloudFrontScope(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationLayerAutomaticResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationLayerAutomaticResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationLayerAutomaticResponseExists(ctx, resourceName, &applicationlayerautomaticresponse),
				),
			},
		},
	})
}

func TestAccShieldApplicationLayerAutomaticResponse_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var applicationlayerautomaticresponse shield.DescribeProtectionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_application_layer_automatic_response.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckWAFV2CloudFrontScope(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationLayerAutomaticResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationLayerAutomaticResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationLayerAutomaticResponseExists(ctx, resourceName, &applicationlayerautomaticresponse),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfshield.ResourceApplicationLayerAutomaticResponse, resourceName),
				),
			},
		},
	})
}

func testAccCheckApplicationLayerAutomaticResponseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_shield_application_layer_automatic_response" {
				continue
			}

			input := &shield.DescribeProtectionInput{
				ResourceArn: aws.String(rs.Primary.ID),
			}
			resp, err := conn.DescribeProtectionWithContext(ctx, input)
			if errs.IsA[*shield.ResourceNotFoundException](err) {
				return nil
			}

			if err != nil {
				return err
			}

			if resp != nil && *resp.Protection.ApplicationLayerAutomaticResponseConfiguration.Status == "DISABLED" {
				return nil
			}
			return create.Error(names.Shield, create.ErrActionCheckingDestroyed, tfshield.ResNameApplicationLayerAutomaticResponse, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApplicationLayerAutomaticResponseExists(ctx context.Context, name string, applicationlayerautomaticresponse *shield.DescribeProtectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameApplicationLayerAutomaticResponse, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameApplicationLayerAutomaticResponse, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn(ctx)
		resp, err := conn.DescribeProtectionWithContext(ctx, &shield.DescribeProtectionInput{
			ResourceArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameApplicationLayerAutomaticResponse, rs.Primary.ID, err)
		}

		*applicationlayerautomaticresponse = *resp

		return nil
	}
}

func testAccApplicationLayerAutomaticResponseConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "CLOUDFRONT"
  default_action {
    allow {}
  }
  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
  lifecycle {
    ignore_changes = [
      rule,
    ]
  }
}
resource "aws_cloudfront_distribution" "test" {
  origin {
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols = [
        "TLSv1",
        "TLSv1.1",
        "TLSv1.2",
      ]
    }
    # This is a fake origin and it's set to this name to indicate that.
    domain_name = "%[1]s.com"
    origin_id   = %[1]q
  }
  enabled             = false
  wait_for_deployment = false
  web_acl_id          = aws_wafv2_web_acl.test.arn
  default_cache_behavior {
    allowed_methods  = ["HEAD", "DELETE", "POST", "GET", "OPTIONS", "PUT", "PATCH"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = %[1]q
    forwarded_values {
      query_string = false
      headers      = ["*"]
      cookies {
        forward = "none"
      }
    }
    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0
  }
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }
  tags = {
    Name = %[1]q
  }
  viewer_certificate {
    cloudfront_default_certificate = true
  }
}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = aws_cloudfront_distribution.test.arn
}
resource "aws_shield_application_layer_automatic_response" "test" {
  resource_arn = aws_cloudfront_distribution.test.arn
  action       = "COUNT"

  depends_on = [
    aws_shield_protection.test,
    aws_cloudfront_distribution.test,
    aws_wafv2_web_acl.test
  ]
}
`, rName)
}
