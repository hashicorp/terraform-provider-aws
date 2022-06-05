package events_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
)

const uuidRegex = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"

func TestAccEventsAPIDestination_basic(t *testing.T) {
	var v1, v2, v3 eventbridge.DescribeApiDestinationOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpoint := "https://example.com/"
	httpMethod := "GET"

	nameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpointModified := "https://example.com/modified"
	httpMethodModified := "POST"

	resourceName := "aws_cloudwatch_event_api_destination.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIDestinationConfig(
					name,
					invocationEndpoint,
					httpMethod,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("api-destination/%s/%s", name, uuidRegex))),
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
				Config: testAccAPIDestinationConfig(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("api-destination/%s/%s", nameModified, uuidRegex))),
					testAccCheckAPIDestinationRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
				),
			},
			{
				Config: testAccAPIDestinationConfig(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(resourceName, &v3),
					testAccCheckAPIDestinationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
				),
			},
		},
	})
}

func TestAccEventsAPIDestination_optional(t *testing.T) {
	var v1, v2, v3 eventbridge.DescribeApiDestinationOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpoint := "https://example.com/"
	httpMethod := "GET"
	description := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationRateLimitPerSecond := 10

	nameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpointModified := "https://example.com/modified"
	httpMethodModified := "POST"
	descriptionModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationRateLimitPerSecondModified := 12

	resourceName := "aws_cloudwatch_event_api_destination.optional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestinationDestroy,
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
					testAccCheckAPIDestinationExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethod),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpoint),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "invocation_rate_limit_per_second", fmt.Sprint(invocationRateLimitPerSecond)),
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
					testAccCheckAPIDestinationExists(resourceName, &v2),
					testAccCheckAPIDestinationRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_rate_limit_per_second", fmt.Sprint(invocationRateLimitPerSecondModified)),
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
					testAccCheckAPIDestinationExists(resourceName, &v3),
					testAccCheckAPIDestinationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_rate_limit_per_second", fmt.Sprint(invocationRateLimitPerSecond)),
				),
			},
		},
	})
}

func TestAccEventsAPIDestination_disappears(t *testing.T) {
	var v eventbridge.DescribeApiDestinationOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpoint := "https://example.com/"
	httpMethod := "GET"

	resourceName := "aws_cloudwatch_event_api_destination.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIDestinationConfig(
					name,
					invocationEndpoint,
					httpMethod,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIDestinationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfevents.ResourceAPIDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAPIDestinationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_api_destination" {
			continue
		}

		params := eventbridge.DescribeApiDestinationInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeApiDestination(&params)

		if err == nil {
			return fmt.Errorf("EventBridge API Destination (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckAPIDestinationExists(n string, v *eventbridge.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn
		params := eventbridge.DescribeApiDestinationInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeApiDestination(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("EventBridge API Destination (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccCheckAPIDestinationRecreated(i, j *eventbridge.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ApiDestinationArn) == aws.StringValue(j.ApiDestinationArn) {
			return fmt.Errorf("EventBridge API Destination not recreated")
		}
		return nil
	}
}

func testAccCheckAPIDestinationNotRecreated(i, j *eventbridge.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ApiDestinationArn) != aws.StringValue(j.ApiDestinationArn) {
			return fmt.Errorf("EventBridge API Destination was recreated")
		}
		return nil
	}
}

func testAccAPIDestinationConfig(name, invocationEndpoint, httpMethod string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_api_destination" "basic" {
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
resource "aws_cloudwatch_event_api_destination" "optional" {
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
