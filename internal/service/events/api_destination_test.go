// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const uuidRegex = "[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$"

func TestAccEventsAPIDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeApiDestinationOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	invocationEndpoint := "https://example.com/"
	httpMethod := "GET"

	nameModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
				Config: testAccAPIDestinationConfig_basic(
					name,
					invocationEndpoint,
					httpMethod,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf("api-destination/%s/%s", name, uuidRegex))),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethod),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpoint),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIDestinationConfig_basic(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, nameModified),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf("api-destination/%s/%s", nameModified, uuidRegex))),
					testAccCheckAPIDestinationRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
				),
			},
			{
				Config: testAccAPIDestinationConfig_basic(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v3),
					testAccCheckAPIDestinationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
				),
			},
		},
	})
}

func TestAccEventsAPIDestination_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeApiDestinationOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	invocationEndpoint := "https://example.com/"
	httpMethod := "GET"
	description := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	invocationRateLimitPerSecond := 10

	nameModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	invocationEndpointModified := "https://example.com/modified"
	httpMethodModified := "POST"
	descriptionModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	invocationRateLimitPerSecondModified := 12

	resourceName := "aws_cloudwatch_event_api_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIDestinationConfig_optional(
					name,
					invocationEndpoint,
					httpMethod,
					description,
					int64(invocationRateLimitPerSecond),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethod),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "invocation_rate_limit_per_second", strconv.Itoa(invocationRateLimitPerSecond)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIDestinationConfig_optional(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
					descriptionModified,
					int64(invocationRateLimitPerSecondModified),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v2),
					testAccCheckAPIDestinationRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, nameModified),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_rate_limit_per_second", strconv.Itoa(invocationRateLimitPerSecondModified)),
				),
			},
			{
				Config: testAccAPIDestinationConfig_optional(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
					descriptionModified,
					int64(invocationRateLimitPerSecond),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v3),
					testAccCheckAPIDestinationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, nameModified),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_rate_limit_per_second", strconv.Itoa(invocationRateLimitPerSecond)),
				),
			},
		},
	})
}

func TestAccEventsAPIDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeApiDestinationOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	invocationEndpoint := "https://example.com/"
	httpMethod := "GET"

	resourceName := "aws_cloudwatch_event_api_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIDestinationConfig_basic(
					name,
					invocationEndpoint,
					httpMethod,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfevents.ResourceAPIDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

func testAccCheckAPIDestinationRecreated(i, j *eventbridge.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.ApiDestinationArn) == aws.ToString(j.ApiDestinationArn) {
			return fmt.Errorf("EventBridge API Destination not recreated")
		}
		return nil
	}
}

func testAccCheckAPIDestinationNotRecreated(i, j *eventbridge.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.ApiDestinationArn) != aws.ToString(j.ApiDestinationArn) {
			return fmt.Errorf("EventBridge API Destination was recreated")
		}
		return nil
	}
}

func testAccAPIDestinationConfig_basic(name, invocationEndpoint, httpMethod string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_api_destination" "test" {
  name                = %[1]q
  invocation_endpoint = %[2]q
  http_method         = %[3]q
  connection_arn      = aws_cloudwatch_event_connection.test.arn
}

resource "aws_cloudwatch_event_connection" "test" {
  name               = %[1]q
  authorization_type = "API_KEY"
  auth_parameters {
    api_key {
      key   = "testKey"
      value = "testValue"
    }
  }
}
`, name, invocationEndpoint, httpMethod)
}

func testAccAPIDestinationConfig_optional(name, invocationEndpoint, httpMethod, description string, invocationRateLimitPerSecond int64) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_api_destination" "test" {
  name                = %[1]q
  invocation_endpoint = %[2]q
  http_method         = %[3]q
  connection_arn      = aws_cloudwatch_event_connection.test.arn

  description                      = %[4]q
  invocation_rate_limit_per_second = %[5]d
}

resource "aws_cloudwatch_event_connection" "test" {
  name               = %[1]q
  authorization_type = "API_KEY"
  auth_parameters {
    api_key {
      key   = "testKey"
      value = "testValue"
    }
  }
}
`, name, invocationEndpoint, httpMethod, description, invocationRateLimitPerSecond)
}
