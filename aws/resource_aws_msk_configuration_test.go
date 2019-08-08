package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSMskConfiguration_basic(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "kafka", regexp.MustCompile(`configuration/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_versions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexp.MustCompile(`auto.create.topics.enable = true`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSMskConfiguration_Description(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSMskConfiguration_KafkaVersions(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationConfigKafkaVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "kafka_versions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSMskConfiguration_ServerProperties(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationConfigServerProperties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration1),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexp.MustCompile(`auto.create.topics.enable = false`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMskConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_msk_configuration" {
			continue
		}

		// The API does not support deletions at this time
	}

	return nil
}

func testAccCheckMskConfigurationExists(resourceName string, configuration *kafka.DescribeConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID not set: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).kafkaconn

		input := &kafka.DescribeConfigurationInput{
			Arn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeConfiguration(input)

		if err != nil {
			return fmt.Errorf("error describing MSK Cluster (%s): %s", rs.Primary.ID, err)
		}

		*configuration = *output

		return nil
	}
}

func testAccMskConfigurationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["2.1.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
delete.topic.enable = true
PROPERTIES
}
`, rName)
}

func testAccMskConfigurationConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  description    = %[2]q
  kafka_versions = ["2.1.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
PROPERTIES
}
`, rName, description)
}

func testAccMskConfigurationConfigKafkaVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["1.1.1", "2.1.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
PROPERTIES
}
`, rName)
}

func testAccMskConfigurationConfigServerProperties(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["2.1.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = false
PROPERTIES
}
`, rName)
}
