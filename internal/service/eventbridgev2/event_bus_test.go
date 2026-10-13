// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/eventbridgev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridgev2/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfeventbridgev2 "github.com/hashicorp/terraform-provider-aws/internal/service/eventbridgev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func checkEventBusARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("events", regexache.MustCompile(`event-busv2/`+name+`/.+`))
}

func checkEventBusARNAlternateRegion(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNAlternateRegionRegexp("events", regexache.MustCompile(`event-busv2/`+name+`/.+`))
}

func TestAccEventBridgeV2EventBus_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeEventBusOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventBusConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventBusExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkEventBusARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.BusStateActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccEventBridgeV2EventBus_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeEventBusOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventBusConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventBusExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfeventbridgev2.ResourceEventBus, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccEventBridgeV2EventBus_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeEventBusOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_event_bus.test"
	key1ResourceName := "aws_kms_key.test[0]"
	key2ResourceName := "aws_kms_key.test[1]"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventBusConfig_full(rName, "description one", 0, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventBusExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description one")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration).AtMapKey("kms_key_identifier"), key1ResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.BusStateActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration").AtMapKey("retention_period_in_days"), knownvalue.Int32Exact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// Update the description, switch to a different customer managed key, and change the retention period.
			{
				Config: testAccEventBusConfig_full(rName, "description two", 1, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventBusExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description two")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration).AtMapKey("kms_key_identifier"), key2ResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.BusStateActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration").AtMapKey("retention_period_in_days"), knownvalue.Int32Exact(2)),
				},
			},
			// Remove the customer managed key, reverting to an AWS owned key.
			{
				Config: testAccEventBusConfig_noEncryption(rName, "description two", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventBusExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.BusStateActive)),
				},
			},
		},
	})
}

func TestAccEventBridgeV2EventBus_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeEventBusOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_event_bus.test"
	keyResourceName := "aws_kms_key.test[0]"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventBusConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventBusExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.Null()),
				},
			},
			{
				Config: testAccEventBusConfig_encryptionConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventBusExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration).AtMapKey("kms_key_identifier"), keyResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				// Keep the keys so that only the event bus changes; UpdateEventBus fails if the current key is pending deletion.
				Config: acctest.ConfigCompose(testAccEventBusConfig_kmsKeys(rName), testAccEventBusConfig_basic(rName)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventBusExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccCheckEventBusDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eventbridgev2_event_bus" {
				continue
			}

			ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

			_, err := tfeventbridgev2.FindEventBusByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge V2 Event Bus %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckEventBusExists(ctx context.Context, t *testing.T, n string, v *eventbridgev2.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

		output, err := tfeventbridgev2.FindEventBusByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

	var input eventbridgev2.ListEventBusesInput
	_, err := conn.ListEventBuses(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEventBusConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_eventbridgev2_event_bus" "test" {
  name = %[1]q
}
`, rName)
}

func testAccEventBusConfig_kmsKeys(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_kms_key" "test" {
  count = 2

  description             = "%[1]s-${count.index}"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "EnableRootAccess"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "AllowEventBridgeServicePrincipal"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        }
        Action = [
          "kms:Decrypt",
          "kms:Encrypt",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:ReEncryptFrom",
          "kms:ReEncryptTo",
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          StringLike = {
            "kms:EncryptionContext:aws:events:event-busv2:arn" = "arn:${data.aws_partition.current.partition}:events:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:event-busv2/%[1]s/*"
          }
        }
      },
    ]
  })
}
`, rName)
}

func testAccEventBusConfig_full(rName, description string, keyIndex, retentionPeriod int) string {
	return acctest.ConfigCompose(testAccEventBusConfig_kmsKeys(rName), fmt.Sprintf(`
resource "aws_eventbridgev2_event_bus" "test" {
  name        = %[1]q
  description = %[2]q

  encryption_configuration = {
    kms_key_identifier = aws_kms_key.test[%[3]d].arn
  }

  storage_configuration = {
    retention_period_in_days = %[4]d
  }
}
`, rName, description, keyIndex, retentionPeriod))
}

func testAccEventBusConfig_noEncryption(rName, description string, retentionPeriod int) string {
	return acctest.ConfigCompose(testAccEventBusConfig_kmsKeys(rName), fmt.Sprintf(`
resource "aws_eventbridgev2_event_bus" "test" {
  name        = %[1]q
  description = %[2]q

  storage_configuration = {
    retention_period_in_days = %[3]d
  }
}
`, rName, description, retentionPeriod))
}

func testAccEventBusConfig_encryptionConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccEventBusConfig_kmsKeys(rName), fmt.Sprintf(`
resource "aws_eventbridgev2_event_bus" "test" {
  name = %[1]q

  encryption_configuration = {
    kms_key_identifier = aws_kms_key.test[0].arn
  }
}
`, rName))
}
