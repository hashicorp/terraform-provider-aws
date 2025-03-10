// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsDeliveryDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DeliveryDestination
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("delivery_destination_type"), knownvalue.StringExact("CWL")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("output_format"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccDeliveryDestinationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccLogsDeliveryDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DeliveryDestination
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflogs.ResourceDeliveryDestination, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsDeliveryDestination_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DeliveryDestination
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryDestinationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
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
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccDeliveryDestinationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccDeliveryDestinationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
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
				Config: testAccDeliveryDestinationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
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

func TestAccLogsDeliveryDestination_outputFormat(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DeliveryDestination
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryDestinationConfig_outputFormat(rName, "w3c"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("output_format"), knownvalue.StringExact("w3c")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccDeliveryDestinationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccDeliveryDestinationConfig_outputFormat(rName, "plain"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("output_format"), knownvalue.StringExact("plain")),
				},
			},
		},
	})
}

func TestAccLogsDeliveryDestination_updateDeliveryDestinationConfigurationSameType(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DeliveryDestination
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryDestinationConfig_deliveryDestinationConfigurationCWL1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("delivery_destination_type"), knownvalue.StringExact("CWL")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccDeliveryDestinationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccDeliveryDestinationConfig_deliveryDestinationConfigurationCWL2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("delivery_destination_type"), knownvalue.StringExact("CWL")),
				},
			},
		},
	})
}

func TestAccLogsDeliveryDestination_updateDeliveryDestinationConfigurationDifferentType(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DeliveryDestination
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_delivery_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryDestinationConfig_deliveryDestinationConfigurationCWL1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("delivery_destination_type"), knownvalue.StringExact("CWL")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccDeliveryDestinationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccDeliveryDestinationConfig_deliveryDestinationConfigurationS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryDestinationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("delivery_destination_type"), knownvalue.StringExact("S3")),
				},
			},
		},
	})
}

func testAccCheckDeliveryDestinationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_delivery_destination" {
				continue
			}

			_, err := tflogs.FindDeliveryDestinationByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Delivery Destination still exists: %s", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckDeliveryDestinationExists(ctx context.Context, n string, v *awstypes.DeliveryDestination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		output, err := tflogs.FindDeliveryDestinationByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDeliveryDestinationImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes[names.AttrName], nil
	}
}

func testAccDeliveryDestinationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_delivery_destination" "test" {
  name = %[1]q

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.test.arn
  }
}
`, rName)
}

func testAccDeliveryDestinationConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_delivery_destination" "test" {
  name = %[1]q

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.test.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccDeliveryDestinationConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_delivery_destination" "test" {
  name = %[1]q

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.test.arn
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccDeliveryDestinationConfig_outputFormat(rName, outputFormat string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_cloudwatch_log_delivery_destination" "test" {
  name          = %[1]q
  output_format = %[2]q

  delivery_destination_configuration {
    destination_resource_arn = aws_s3_bucket.test.arn
  }
}
`, rName, outputFormat)
}

func testAccDeliveryDestinationConfig_deliveryDestinationConfigurationBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccDeliveryDestinationConfig_deliveryDestinationConfigurationCWL1(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryDestinationConfig_deliveryDestinationConfigurationBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_delivery_destination" "test" {
  name = %[1]q

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.test1.arn
  }
}
`, rName))
}

func testAccDeliveryDestinationConfig_deliveryDestinationConfigurationCWL2(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryDestinationConfig_deliveryDestinationConfigurationBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_delivery_destination" "test" {
  name = %[1]q

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.test2.arn
  }
}
`, rName))
}

func testAccDeliveryDestinationConfig_deliveryDestinationConfigurationS3(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryDestinationConfig_deliveryDestinationConfigurationBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_delivery_destination" "test" {
  name = %[1]q

  delivery_destination_configuration {
    destination_resource_arn = aws_s3_bucket.test.arn
  }
}
`, rName))
}
