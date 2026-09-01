// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontCreateInvalidationAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckDistributionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCreateInvalidationActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, t, resourceName, &distribution),
					testAccCheckInvalidationExists(ctx, t, &distribution, []string{"/*"}),
				),
			},
		},
	})
}

// Helper: Check invalidation exists and is completed
func testAccCheckInvalidationExists(ctx context.Context, t *testing.T, distribution *awstypes.Distribution, expectedPaths []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if distribution == nil || distribution.Id == nil {
			return fmt.Errorf("Distribution is nil or has no ID")
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		// List invalidations for this distribution
		listInput := &cloudfront.ListInvalidationsInput{
			DistributionId: distribution.Id,
		}
		out, err := conn.ListInvalidations(ctx, listInput)
		if err != nil {
			return fmt.Errorf("failed to list invalidations: %w", err)
		}

		if len(out.InvalidationList.Items) == 0 {
			return fmt.Errorf("no invalidations found for distribution %s", *distribution.Id)
		}

		// Get the most recent invalidation
		latest := out.InvalidationList.Items[0]

		// Get invalidation details
		getInput := &cloudfront.GetInvalidationInput{
			DistributionId: distribution.Id,
			Id:             latest.Id,
		}
		getOut, err := conn.GetInvalidation(ctx, getInput)
		if err != nil {
			return fmt.Errorf("failed to get invalidation %s: %w", *latest.Id, err)
		}

		invalidation := getOut.Invalidation

		// Check that the invalidation contains the expected paths
		if invalidation.InvalidationBatch == nil || invalidation.InvalidationBatch.Paths == nil {
			return fmt.Errorf("invalidation batch or paths is nil")
		}

		actualPaths := invalidation.InvalidationBatch.Paths.Items
		if len(actualPaths) != len(expectedPaths) {
			return fmt.Errorf("expected %d paths, got %d", len(expectedPaths), len(actualPaths))
		}

		// Create a map for easy lookup
		pathMap := make(map[string]bool)
		for _, path := range actualPaths {
			pathMap[path] = true
		}

		// Check each expected path exists
		for _, expectedPath := range expectedPaths {
			if !pathMap[expectedPath] {
				return fmt.Errorf("expected path %s not found in invalidation", expectedPath)
			}
		}

		// Wait for invalidation to complete (with timeout)
		maxAttempts := 60 // 10 minutes at 10-second intervals
		for attempt := range maxAttempts {
			statusInput := &cloudfront.GetInvalidationInput{
				DistributionId: distribution.Id,
				Id:             latest.Id,
			}
			statusOut, err := conn.GetInvalidation(ctx, statusInput)
			if err != nil {
				return fmt.Errorf("failed to check invalidation status: %w", err)
			}

			if *statusOut.Invalidation.Status == "Completed" {
				return nil
			}

			if attempt < maxAttempts-1 {
				time.Sleep(10 * time.Second)
			}
		}

		return fmt.Errorf("invalidation %s did not complete within timeout", *latest.Id)
	}
}

// Terraform configuration with action trigger
func testAccCreateInvalidationActionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Use faster settings for testing
  enabled             = true
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    min_ttl     = 0
    default_ttl = 0
    max_ttl     = 0
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

  tags = {
    Name = %[1]q
  }
}

action "aws_cloudfront_create_invalidation" "test" {
  config {
    distribution_id = aws_cloudfront_distribution.test.id
    paths           = ["/*"]
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_cloudfront_create_invalidation.test]
    }
  }
}
`, rName)
}
