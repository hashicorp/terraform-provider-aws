// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/eventbridgev2"
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

func checkEventSourceARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("events", regexache.MustCompile(`event-sourcev2/.+/`+name+`/.+`))
}

func checkEventSourceARNAlternateRegion(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNAlternateRegionRegexp("events", regexache.MustCompile(`event-sourcev2/.+/`+name+`/.+`))
}

func TestAccEventBridgeV2EventSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeEventSourceOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_event_source.test"
	busResourceName := "aws_eventbridgev2_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventSourceExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkEventSourceARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"aws_service_events_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"aws_service": knownvalue.StringExact("aws.s3"),
								}),
							}),
							"partner_events_configuration": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("event_bus_arn"), busResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
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

func TestAccEventBridgeV2EventSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeEventSourceOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_event_source.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventSourceExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfeventbridgev2.ResourceEventSource, resourceName),
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

func TestAccEventBridgeV2EventSource_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.DescribeEventSourceOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_event_source.test"
	queueResourceName := "aws_sqs_queue.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourceConfig_full(rName, "description one", "Object Created"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventSourceExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description one")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrConfiguration).AtSliceIndex(0).AtMapKey("aws_service_events_configuration").AtSliceIndex(0).AtMapKey("on_failure_configuration").AtSliceIndex(0).AtMapKey(names.AttrARN), queueResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
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
				Config: testAccEventSourceConfig_full(rName, "description two", "Object Deleted"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventSourceExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description two")),
				},
			},
		},
	})
}

func testAccCheckEventSourceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eventbridgev2_event_source" {
				continue
			}

			ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

			_, err := tfeventbridgev2.FindEventSourceByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge V2 Event Source %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckEventSourceExists(ctx context.Context, t *testing.T, n string, v *eventbridgev2.DescribeEventSourceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

		output, err := tfeventbridgev2.FindEventSourceByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEventSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_eventbridgev2_event_bus" "test" {
  name = %[1]q
}

resource "aws_eventbridgev2_event_source" "test" {
  name          = %[1]q
  event_bus_arn = aws_eventbridgev2_event_bus.test.arn

  configuration {
    aws_service_events_configuration {
      aws_service = "aws.s3"
    }
  }
}
`, rName)
}

func testAccEventSourceConfig_full(rName, description, detailType string) string {
	return fmt.Sprintf(`
resource "aws_eventbridgev2_event_bus" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_eventbridgev2_event_source" "test" {
  name          = %[1]q
  description   = %[2]q
  event_bus_arn = aws_eventbridgev2_event_bus.test.arn

  configuration {
    aws_service_events_configuration {
      aws_service = "aws.s3"
      pattern     = jsonencode({ "detail-type" = [%[3]q] })

      on_failure_configuration {
        arn = aws_sqs_queue.test.arn
      }
    }
  }
}
`, rName, description, detailType)
}
