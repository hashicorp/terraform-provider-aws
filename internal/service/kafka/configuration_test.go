package kafka_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
)

func TestAccKafkaConfiguration_basic(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName, &configuration1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kafka", regexp.MustCompile(`configuration/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_versions.#", "0"),
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

func TestAccKafkaConfiguration_disappears(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName, &configuration1),
					acctest.CheckResourceDisappears(acctest.Provider, tfkafka.ResourceConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaConfiguration_description(t *testing.T) {
	var configuration1, configuration2 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "2"),
				),
			},
		},
	})
}

func TestAccKafkaConfiguration_kafkaVersions(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_versions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "kafka_versions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "kafka_versions.*", "2.6.0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "kafka_versions.*", "2.7.0"),
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

func TestAccKafkaConfiguration_serverProperties(t *testing.T) {
	var configuration1, configuration2 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"
	serverProperty1 := "auto.create.topics.enable = false"
	serverProperty2 := "auto.create.topics.enable = true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfigServerProperties(rName, serverProperty1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName, &configuration1),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexp.MustCompile(serverProperty1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationConfigServerProperties(rName, serverProperty2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "2"),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexp.MustCompile(serverProperty2)),
				),
			},
		},
	})
}

func testAccCheckConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_msk_configuration" {
			continue
		}

		input := &kafka.DescribeConfigurationInput{
			Arn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeConfiguration(input)

		if tfawserr.ErrMessageContains(err, kafka.ErrCodeBadRequestException, "Configuration ARN does not exist") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("MSK Configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckConfigurationExists(resourceName string, configuration *kafka.DescribeConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID not set: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConn

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

func testAccConfigurationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  name = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
delete.topic.enable = true
PROPERTIES
}
`, rName)
}

func testAccConfigurationConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  description = %[2]q
  name        = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
PROPERTIES
}
`, rName, description)
}

func testAccConfigurationConfig_versions(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["2.6.0", "2.7.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
PROPERTIES
}
`, rName)
}

func testAccConfigurationConfigServerProperties(rName string, serverProperty string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  name = %[1]q

  server_properties = <<PROPERTIES
%[2]s
PROPERTIES
}
`, rName, serverProperty)
}
