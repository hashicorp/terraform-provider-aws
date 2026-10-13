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

func checkSubscriberARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("events", regexache.MustCompile(`.+/`+name+`(/.+)?$`))
}

func checkSubscriberARNAlternateRegion(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNAlternateRegionRegexp("events", regexache.MustCompile(`.+/`+name+`(/.+)?$`))
}

func TestAccEventBridgeV2Subscriber_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeSubscriberOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_subscriber.test"
	busResourceName := "aws_eventbridgev2_event_bus.test"
	queueResourceName := "aws_sqs_queue.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkSubscriberARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("batch_configuration"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.NotNull()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("event_bus_arn"), busResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("invoke_configuration").AtSliceIndex(0).AtMapKey(names.AttrTargetARN), queueResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.SubscriberStateRunning)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				// The service returns resolved defaults for these, which are not tracked unless configured.
				ImportStateVerifyIgnore: []string{
					"batch_configuration",
					"log_configuration",
					"retry_policy",
					"transformer",
				},
			},
		},
	})
}

func TestAccEventBridgeV2Subscriber_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeSubscriberOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_subscriber.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfeventbridgev2.ResourceSubscriber, resourceName),
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

func TestAccEventBridgeV2Subscriber_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeSubscriberOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_subscriber.test"
	queueResourceName := "aws_sqs_queue.dlq"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_full(rName, "description one", string(awstypes.SubscriberStateRunning), 10, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("batch_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_batch_size":              knownvalue.Int32Exact(10),
							"max_batch_window_in_seconds": knownvalue.NotNull(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description one")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("log_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"include_payload": knownvalue.Null(), // Not returned unless configured.
							"level":           tfknownvalue.StringExact(awstypes.LogLevelError),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("on_failure_configuration").AtSliceIndex(0).AtMapKey(names.AttrARN), queueResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("retry_policy"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_event_age_in_seconds": knownvalue.NotNull(),
							"max_retry_attempts":       knownvalue.Int32Exact(3),
							"retry_strategy":           knownvalue.NotNull(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.SubscriberStateRunning)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore: []string{
					"transformer", // Service default is not tracked unless configured.
				},
			},
			{
				Config: testAccSubscriberConfig_full(rName, "description two", string(awstypes.SubscriberStateStopped), 5, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("batch_configuration").AtSliceIndex(0).AtMapKey("max_batch_size"), knownvalue.Int32Exact(5)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description two")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("retry_policy").AtSliceIndex(0).AtMapKey("max_retry_attempts"), knownvalue.Int32Exact(5)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.SubscriberStateStopped)),
				},
			},
		},
	})
}

func testAccCheckSubscriberDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eventbridgev2_subscriber" {
				continue
			}

			ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

			_, err := tfeventbridgev2.FindSubscriberByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge V2 Subscriber %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckSubscriberExists(ctx context.Context, t *testing.T, n string, v *eventbridgev2.DescribeSubscriberOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

		output, err := tfeventbridgev2.FindSubscriberByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSubscriberConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_eventbridgev2_event_bus" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "events.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "sqs:SendMessage"
      Effect   = "Allow"
      Resource = aws_sqs_queue.test.arn
    }]
  })
}
`, rName)
}

func testAccSubscriberConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSubscriberConfig_base(rName), fmt.Sprintf(`
resource "aws_eventbridgev2_subscriber" "test" {
  name          = %[1]q
  event_bus_arn = aws_eventbridgev2_event_bus.test.arn

  filter_configuration {
    filter {
      scope   = "METADATA"
      pattern = jsonencode({ source = ["example"] })
    }
  }

  invoke_configuration {
    role_arn   = aws_iam_role.test.arn
    target_arn = aws_sqs_queue.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccSubscriberConfig_full(rName, description, state string, maxBatchSize, maxRetryAttempts int) string {
	return acctest.ConfigCompose(testAccSubscriberConfig_base(rName), fmt.Sprintf(`
resource "aws_sqs_queue" "dlq" {
  name = "%[1]s-dlq"
}

resource "aws_eventbridgev2_subscriber" "test" {
  name          = %[1]q
  description   = %[2]q
  event_bus_arn = aws_eventbridgev2_event_bus.test.arn
  state         = %[3]q

  batch_configuration {
    max_batch_size = %[4]d
  }

  filter_configuration {
    filter {
      scope   = "METADATA"
      pattern = jsonencode({ source = ["example"] })
    }
  }

  invoke_configuration {
    role_arn   = aws_iam_role.test.arn
    target_arn = aws_sqs_queue.test.arn

    sqs_parameters {
      message_attributes {
        name         = "origin"
        data_type    = "String"
        string_value = "eventbridge"
      }
    }
  }

  log_configuration {
    level = "ERROR"
  }

  on_failure_configuration {
    arn = aws_sqs_queue.dlq.arn
  }

  retry_policy {
    max_retry_attempts = %[5]d
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, description, state, maxBatchSize, maxRetryAttempts))
}
