package cloudwatchevents_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchevents "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchevents"
)

const uuidRegex = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"

func TestAccCloudWatchEventsAPIDestination_basic(t *testing.T) {
	var v1, v2, v3 events.DescribeApiDestinationOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpoint := "https://www.hashicorp.com/"
	httpMethod := "GET"

	nameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpointModified := "https://www.hashicorp.com/products/terraform"
	httpMethodModified := "POST"

	resourceName := "aws_cloudwatch_event_api_destination.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAPIDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIDestinationConfig(
					name,
					invocationEndpoint,
					httpMethod,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v1),
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
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("api-destination/%s/%s", nameModified, uuidRegex))),
					testAccCheckCloudWatchEventApiDestinationRecreated(&v1, &v2),
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
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v3),
					testAccCheckCloudWatchEventApiDestinationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
				),
			},
		},
	})
}

func TestAccCloudWatchEventsAPIDestination_optional(t *testing.T) {
	var v1, v2, v3 events.DescribeApiDestinationOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpoint := "https://www.hashicorp.com/"
	httpMethod := "GET"
	description := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationRateLimitPerSecond := 10

	nameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpointModified := "https://www.hashicorp.com/products/terraform"
	httpMethodModified := "POST"
	descriptionModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationRateLimitPerSecondModified := 12

	resourceName := "aws_cloudwatch_event_api_destination.optional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAPIDestinationDestroy,
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
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v1),
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
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v2),
					testAccCheckCloudWatchEventApiDestinationRecreated(&v1, &v2),
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
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v3),
					testAccCheckCloudWatchEventApiDestinationNotRecreated(&v2, &v3),
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

func TestAccCloudWatchEventsAPIDestination_disappears(t *testing.T) {
	var v events.DescribeApiDestinationOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	invocationEndpoint := "https://www.hashicorp.com/"
	httpMethod := "GET"

	resourceName := "aws_cloudwatch_event_api_destination.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAPIDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIDestinationConfig(
					name,
					invocationEndpoint,
					httpMethod,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatchevents.ResourceAPIDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAPIDestinationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEventsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_api_destination" {
			continue
		}

		params := events.DescribeApiDestinationInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeApiDestination(&params)

		if err == nil {
			return fmt.Errorf("CloudWatch Events Api Destination (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckCloudWatchEventApiDestinationExists(n string, v *events.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEventsConn
		params := events.DescribeApiDestinationInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeApiDestination(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("CloudWatch Events Api Destination (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccCheckCloudWatchEventApiDestinationRecreated(i, j *events.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ApiDestinationArn) == aws.StringValue(j.ApiDestinationArn) {
			return fmt.Errorf("CloudWatch Events Api Destination not recreated")
		}
		return nil
	}
}

func testAccCheckCloudWatchEventApiDestinationNotRecreated(i, j *events.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ApiDestinationArn) != aws.StringValue(j.ApiDestinationArn) {
			return fmt.Errorf("CloudWatch Events Api Destination was recreated")
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
