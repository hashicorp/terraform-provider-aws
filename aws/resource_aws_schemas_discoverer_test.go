package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	schemas "github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_schemas_discoverer", &resource.Sweeper{
		Name: "aws_schemas_discoverer",
		F:    testSweepSchemasDiscoverer,
		Dependencies: []string{
			"aws_cloudwatch_event_bus",
		},
	})
}

func testSweepSchemasDiscoverer(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).schemasconn

	var sweeperErrs *multierror.Error

	input := &schemas.ListDiscoverersInput{
		Limit: aws.Int64(100),
	}
	var discoverers []*schemas.DiscovererSummary
	for {
		output, err := conn.ListDiscoverers(input)
		if err != nil {
			return err
		}
		discoverers = append(discoverers, output.Discoverers...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, discoverer := range discoverers {

		input := &schemas.DeleteDiscovererInput{
			DiscovererId: discoverer.DiscovererId,
		}
		_, err := conn.DeleteDiscoverer(input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting Schemas Discoverer (%s): %w", *discoverer.DiscovererId, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d Schemas Discoverers", len(discoverers))

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSchemasDiscoverer_basic(t *testing.T) {
	var v1, v2, v3 schemas.DescribeDiscovererOutput
	eventbusName := acctest.RandomWithPrefix("tf-acc-test")
	eventbusNameModified := acctest.RandomWithPrefix("tf-acc-test")

	description := acctest.RandomWithPrefix("tf-acc-test")
	descriptionModified := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasDiscovererDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasDiscovererConfig(eventbusName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("discoverer/events-event-bus-%s", eventbusName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasDiscovererConfig(eventbusNameModified, descriptionModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("discoverer/events-event-bus-%s", eventbusNameModified)),
					testAccCheckSchemasDiscovererRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSSchemasDiscovererConfig_Tags1(eventbusNameModified, "key", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v3),
					testAccCheckSchemasDiscovererNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
		},
	})
}

func TestAccAWSSchemasDiscoverer_tags(t *testing.T) {
	var v1, v2, v3, v4 schemas.DescribeDiscovererOutput
	eventbusName := acctest.RandomWithPrefix("tf-acc-test")
	description := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasDiscovererDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasDiscovererConfig_Tags1(eventbusName, "key1", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasDiscovererConfig_Tags2(eventbusName, "key1", "updated", "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v2),
					testAccCheckSchemasDiscovererNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccAWSSchemasDiscovererConfig_Tags1(eventbusName, "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v3),
					testAccCheckSchemasDiscovererNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccAWSSchemasDiscovererConfig(eventbusName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v4),
					testAccCheckSchemasDiscovererNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSSchemasDiscoverer_disappears(t *testing.T) {
	var v schemas.DescribeDiscovererOutput
	eventbusName := acctest.RandomWithPrefix("tf-acc-test")
	description := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasDiscovererDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasDiscovererConfig(eventbusName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSchemasDiscoverer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSchemasDiscovererDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).schemasconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_schemas_discoverer" {
			continue
		}

		params := schemas.DescribeDiscovererInput{
			DiscovererId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDiscoverer(&params)

		if err == nil {
			return fmt.Errorf("Schemas Discoverer (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckSchemasDiscovererExists(n string, v *schemas.DescribeDiscovererOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).schemasconn
		params := schemas.DescribeDiscovererInput{
			DiscovererId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeDiscoverer(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("Schemas Discoverer (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccCheckSchemasDiscovererRecreated(i, j *schemas.DescribeDiscovererOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.DiscovererArn) == aws.StringValue(j.DiscovererArn) {
			return fmt.Errorf("Schemas Discoverer not recreated")
		}
		return nil
	}
}

func testAccCheckSchemasDiscovererNotRecreated(i, j *schemas.DescribeDiscovererOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.DiscovererArn) != aws.StringValue(j.DiscovererArn) {
			return fmt.Errorf("Schemas Discoverer was recreated")
		}
		return nil
	}
}

func testAccAWSSchemasDiscovererConfig(eventbusName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
	name = %[1]q
}

resource "aws_schemas_discoverer" "test" {
  source_arn = aws_cloudwatch_event_bus.test.arn
  description = %[2]q
}
`, eventbusName, description)
}

func testAccAWSSchemasDiscovererConfig_Tags1(eventbusName, key, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
	name = %[1]q
}
	
resource "aws_schemas_discoverer" "test" {
	source_arn = aws_cloudwatch_event_bus.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, eventbusName, key, value)
}

func testAccAWSSchemasDiscovererConfig_Tags2(eventbusName, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
	name = %[1]q
}
	
resource "aws_schemas_discoverer" "test" {
	source_arn = aws_cloudwatch_event_bus.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, eventbusName, key1, value1, key2, value2)
}
