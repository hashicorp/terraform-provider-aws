// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudtrail_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudTrailInsightSelectors_trailBasic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_insight_selectors.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightSelectorsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightSelectorsConfig_trailBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInsightSelectorsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type": "ApiCallRateInsight",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Add a second insight type.
				Config: testAccInsightSelectorsConfig_trailMultipleTypes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type": "ApiCallRateInsight",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type": "ApiErrorRateInsight",
					}),
				),
			},
			{
				// Reduce back to one.
				Config: testAccInsightSelectorsConfig_trailBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type": "ApiCallRateInsight",
					}),
				),
			},
		},
	})
}

func TestAccCloudTrailInsightSelectors_trailEventCategories(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_insight_selectors.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightSelectorsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightSelectorsConfig_trailEventCategories(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInsightSelectorsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type":       "ApiCallRateInsight",
						"event_categories.#": "1",
						"event_categories.0": "Management",
					}),
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

func TestAccCloudTrailInsightSelectors_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_insight_selectors.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightSelectorsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightSelectorsConfig_trailBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInsightSelectorsExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcloudtrail.ResourceInsightSelectors(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudTrailInsightSelectors_eventDataStore(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_insight_selectors.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightSelectorsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightSelectorsConfig_eventDataStore(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInsightSelectorsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "event_data_store"),
					resource.TestCheckResourceAttrSet(resourceName, "insights_destination"),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type": "ApiCallRateInsight",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Add a second insight type.
				Config: testAccInsightSelectorsConfig_eventDataStoreMultipleTypes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type": "ApiCallRateInsight",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type": "ApiErrorRateInsight",
					}),
				),
			},
		},
	})
}

// testAccCheckInsightSelectorsExists verifies that insight selectors are enabled on the
// trail or event data store identified by the resource's primary ID.
func testAccCheckInsightSelectorsExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudTrailClient(ctx)

		if strings.Contains(rs.Primary.ID, ":eventdatastore/") {
			_, err := tfcloudtrail.FindInsightSelectorsByEventDataStore(ctx, conn, rs.Primary.ID)
			return err
		}

		_, err := tfcloudtrail.FindInsightSelectorsByTrailName(ctx, conn, rs.Primary.ID)
		return err
	}
}

func testAccCheckInsightSelectorsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudTrailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudtrail_insight_selectors" {
				continue
			}

			var err error
			if strings.Contains(rs.Primary.ID, ":eventdatastore/") {
				_, err = tfcloudtrail.FindInsightSelectorsByEventDataStore(ctx, conn, rs.Primary.ID)
			} else {
				_, err = tfcloudtrail.FindInsightSelectorsByTrailName(ctx, conn, rs.Primary.ID)
			}

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudTrail Insight Selectors (%s) still exist", rs.Primary.ID)
		}

		return nil
	}
}

// testAccInsightSelectorsConfig_trailBase creates the S3/IAM infrastructure (via
// testAccCloudTrailConfig_base) plus a minimal trail that logs management events.
func testAccInsightSelectorsConfig_trailBase(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.bucket
}
`, rName))
}

func testAccInsightSelectorsConfig_trailBasic(rName string) string {
	return acctest.ConfigCompose(testAccInsightSelectorsConfig_trailBase(rName), `
resource "aws_cloudtrail_insight_selectors" "test" {
  trail_name = aws_cloudtrail.test.arn

  insight_selector {
    insight_type = "ApiCallRateInsight"
  }
}
`)
}

func testAccInsightSelectorsConfig_trailMultipleTypes(rName string) string {
	return acctest.ConfigCompose(testAccInsightSelectorsConfig_trailBase(rName), `
resource "aws_cloudtrail_insight_selectors" "test" {
  trail_name = aws_cloudtrail.test.arn

  insight_selector {
    insight_type = "ApiCallRateInsight"
  }

  insight_selector {
    insight_type = "ApiErrorRateInsight"
  }
}
`)
}

func testAccInsightSelectorsConfig_trailEventCategories(rName string) string {
	return acctest.ConfigCompose(testAccInsightSelectorsConfig_trailBase(rName), `
resource "aws_cloudtrail_insight_selectors" "test" {
  trail_name = aws_cloudtrail.test.arn

  insight_selector {
    insight_type     = "ApiCallRateInsight"
    event_categories = ["Management"]
  }
}
`)
}

func testAccInsightSelectorsConfig_eventDataStore(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "source" {
  name                           = "%[1]s-src"
  termination_protection_enabled = false
}

resource "aws_cloudtrail_event_data_store" "destination" {
  name                           = "%[1]s-dst"
  termination_protection_enabled = false
}

resource "aws_cloudtrail_insight_selectors" "test" {
  event_data_store     = aws_cloudtrail_event_data_store.source.arn
  insights_destination = aws_cloudtrail_event_data_store.destination.arn

  insight_selector {
    insight_type = "ApiCallRateInsight"
  }
}
`, rName)
}

func testAccInsightSelectorsConfig_eventDataStoreMultipleTypes(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "source" {
  name                           = "%[1]s-src"
  termination_protection_enabled = false
}

resource "aws_cloudtrail_event_data_store" "destination" {
  name                           = "%[1]s-dst"
  termination_protection_enabled = false
}

resource "aws_cloudtrail_insight_selectors" "test" {
  event_data_store     = aws_cloudtrail_event_data_store.source.arn
  insights_destination = aws_cloudtrail_event_data_store.destination.arn

  insight_selector {
    insight_type = "ApiCallRateInsight"
  }

  insight_selector {
    insight_type = "ApiErrorRateInsight"
  }
}
`, rName)
}
