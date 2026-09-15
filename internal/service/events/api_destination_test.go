// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const uuidRegex = "[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$"

func checkAPIDestinationARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("events", regexache.MustCompile(`api-destination/`+name+`/`+uuidRegex))
}

func TestAccEventsAPIDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeApiDestinationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_api_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIDestination/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkAPIDestinationARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("http_method"), knownvalue.StringExact("GET")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invocation_endpoint"), knownvalue.StringExact("https://example.com/")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIDestination/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEventsAPIDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeApiDestinationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_api_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIDestination/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfevents.ResourceAPIDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccEventsAPIDestination_updateRequiredAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeApiDestinationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	invocationEndpoint := "https://example.com/"
	httpMethod := "GET"
	rNameModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	invocationEndpointModified := "https://example.com/modified"
	httpMethodModified := "POST"

	resourceName := "aws_cloudwatch_event_api_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIDestination/required_attributes/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:       config.StringVariable(rName),
					"http_method":         config.StringVariable(httpMethod),
					"invocation_endpoint": config.StringVariable(invocationEndpoint),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkAPIDestinationARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("http_method"), knownvalue.StringExact(httpMethod)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invocation_endpoint"), knownvalue.StringExact(invocationEndpoint)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIDestination/required_attributes/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:       config.StringVariable(rNameModified),
					"http_method":         config.StringVariable(httpMethod),
					"invocation_endpoint": config.StringVariable(invocationEndpoint),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkAPIDestinationARN(rNameModified)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("http_method"), knownvalue.StringExact(httpMethod)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invocation_endpoint"), knownvalue.StringExact(invocationEndpoint)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameModified)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIDestination/required_attributes/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:       config.StringVariable(rNameModified),
					"http_method":         config.StringVariable(httpMethodModified),
					"invocation_endpoint": config.StringVariable(invocationEndpointModified),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkAPIDestinationARN(rNameModified)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("http_method"), knownvalue.StringExact(httpMethodModified)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invocation_endpoint"), knownvalue.StringExact(invocationEndpointModified)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameModified)),
				},
			},
		},
	})
}

func TestAccEventsAPIDestination_updateOptionalAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeApiDestinationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var invocationRateLimitPerSecond int64 = 10
	descriptionModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var invocationRateLimitPerSecondModified int64 = 12

	resourceName := "aws_cloudwatch_event_api_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIDestination/optional_attributes/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                    config.StringVariable(rName),
					"description":                      config.StringVariable(description),
					"invocation_rate_limit_per_second": config.IntegerVariable(invocationRateLimitPerSecond),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invocation_rate_limit_per_second"), knownvalue.Int64Exact(invocationRateLimitPerSecond)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIDestination/optional_attributes/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                    config.StringVariable(rName),
					"description":                      config.StringVariable(description),
					"invocation_rate_limit_per_second": config.IntegerVariable(invocationRateLimitPerSecond),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/APIDestination/optional_attributes/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                    config.StringVariable(rName),
					"description":                      config.StringVariable(descriptionModified),
					"invocation_rate_limit_per_second": config.IntegerVariable(invocationRateLimitPerSecondModified),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(descriptionModified)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invocation_rate_limit_per_second"), knownvalue.Int64Exact(invocationRateLimitPerSecondModified)),
				},
			},
		},
	})
}

func testAccCheckAPIDestinationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_event_api_destination" {
				continue
			}

			_, err := tfevents.FindAPIDestinationByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge API Destination %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAPIDestinationExists(ctx context.Context, t *testing.T, n string, v *eventbridge.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		output, err := tfevents.FindAPIDestinationByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
