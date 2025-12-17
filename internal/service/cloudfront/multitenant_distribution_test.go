// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontMultiTenantDistribution_basic(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.0.parameter_definition.0.definition.0.string_schema.0.required", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccCloudFrontMultiTenantDistribution_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, resourceName, &distribution),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceMultiTenantDistribution, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMultiTenantDistributionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_multitenant_distribution" {
				continue
			}

			_, err := tfcloudfront.FindDistributionByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Multi-tenant Distribution %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMultiTenantDistributionExists(ctx context.Context, n string, v *awstypes.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindDistributionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.Distribution

		return nil
	}
}

func testAccMultiTenantDistributionConfig_basic() string {
	return `
resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled = false
  comment = "Test multi-tenant distribution"

  origin {
    domain_name = "example.com"
    id          = "example"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id       = "example"
    viewer_protocol_policy = "redirect-to-https"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # AWS Managed CachingDisabled policy

    allowed_methods {
      items          = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
      cached_methods = ["GET", "HEAD", "OPTIONS"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  tenant_config {
    parameter_definition {
      name = "origin_domain"
      definition {
        string_schema {
          required = true
          comment  = "Origin domain parameter for tenants"
        }
      }
    }
  }
}
`
}
