// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDelivery_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Delivery
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"field_delimiter",
					"s3_delivery_configuration.0.enable_hive_compatible_path",
				},
			},
		},
	})
}

func testAccDelivery_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Delivery
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflogs.ResourceDelivery, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDelivery_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Delivery
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"field_delimiter",
					"s3_delivery_configuration.0.enable_hive_compatible_path",
				},
			},
			{
				Config: testAccDeliveryConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccDeliveryConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccDelivery_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Delivery
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryConfig_allAttributes(rName, " ", "{region}/{yyyy}/{MM}/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("field_delimiter"), knownvalue.StringExact(" ")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("record_fields"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("event_timestamp"),
						knownvalue.StringExact("event"),
					})),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("s3_delivery_configuration").AtSliceIndex(0).AtMapKey("suffix_path"),
						knownvalue.StringExact("{region}/{yyyy}/{MM}/")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"field_delimiter",
					"s3_delivery_configuration.0.enable_hive_compatible_path",
				},
			},
			{
				Config: testAccDeliveryConfig_allAttributes(rName, "", "{region}/{yyyy}/{MM}/{dd}/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("field_delimiter"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("record_fields"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("event_timestamp"),
						knownvalue.StringExact("event"),
					})),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("s3_delivery_configuration").AtSliceIndex(0).AtMapKey("suffix_path"),
						knownvalue.StringExact("{region}/{yyyy}/{MM}/{dd}/")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"field_delimiter",
					"s3_delivery_configuration.0.enable_hive_compatible_path",
				},
			},
		},
	})
}

func testAccDelivery_cloudFrontDistribution(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Delivery
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// CloudFront distribution delivery source must be in us-east-1.
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryConfig_cloudFrontDistribution(rName, "/123456678910/{DistributionId}/{yyyy}/{MM}/{dd}/{HH}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_delivery_configuration").AtSliceIndex(0).AtMapKey("suffix_path"), knownvalue.StringExact("/123456678910/{DistributionId}/{yyyy}/{MM}/{dd}/{HH}")),
				},
			},
			{
				Config: testAccDeliveryConfig_cloudFrontDistribution(rName, "/987654321098/{DistributionId}/{yyyy}/{MM}/{dd}/{HH}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_delivery_configuration").AtSliceIndex(0).AtMapKey("suffix_path"), knownvalue.StringExact("/987654321098/{DistributionId}/{yyyy}/{MM}/{dd}/{HH}")),
				},
			},
		},
	})
}

func testAccCheckDeliveryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_delivery" {
				continue
			}

			_, err := tflogs.FindDeliveryByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Delivery still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDeliveryExists(ctx context.Context, t *testing.T, n string, v *awstypes.Delivery) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindDeliveryByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDeliveryConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDeliverySourceConfig_basic(rName), testAccDeliveryDestinationConfig_basic(rName), `
resource "aws_cloudwatch_log_delivery" "test" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.test.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.test.arn
}
`)
}

func testAccDeliveryConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccDeliverySourceConfig_basic(rName), testAccDeliveryDestinationConfig_basic(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_delivery" "test" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.test.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.test.arn

  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value))
}

func testAccDeliveryConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccDeliverySourceConfig_basic(rName), testAccDeliveryDestinationConfig_basic(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_delivery" "test" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.test.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccDeliveryConfig_allAttributes(rName, fieldDelimiter, suffixPath string) string {
	return acctest.ConfigCompose(testAccDeliverySourceConfig_basic(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_cloudwatch_log_delivery_destination" "test" {
  name          = %[1]q
  output_format = "w3c"

  delivery_destination_configuration {
    destination_resource_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_cloudwatch_log_delivery" "test" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.test.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.test.arn

  field_delimiter = %[2]q

  record_fields = ["event_timestamp", "event"]

  s3_delivery_configuration {
    enable_hive_compatible_path = false
    suffix_path                 = %[3]q
  }
}
`, rName, fieldDelimiter, suffixPath))
}

func testAccDeliveryConfig_cloudFrontDistribution(rName, suffixPath string) string {
	return acctest.ConfigCompose(testAccDeliverySourceConfig_baseCloudFrontDistribution(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_delivery_source" "test" {
  name         = %[1]q
  log_type     = "ACCESS_LOGS"
  resource_arn = aws_cloudfront_distribution.test.arn
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_cloudwatch_log_delivery_destination" "test" {
  name          = %[1]q
  output_format = "parquet"

  delivery_destination_configuration {
    destination_resource_arn = "${aws_s3_bucket.test.arn}/prefix"
  }
}

resource "aws_cloudwatch_log_delivery" "test" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.test.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.test.arn

  s3_delivery_configuration {
    suffix_path = %[2]q
  }
}
`, rName, suffixPath))
}
