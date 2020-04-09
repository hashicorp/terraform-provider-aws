package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAwsIoTEventConfiguration(t *testing.T) {
	var eventConfiguration iot.DescribeEventConfigurationsOutput
	rString := acctest.RandString(8)
	attributeName := fmt.Sprintf("tf_acc_authorizer_%s", rString)
	resourceName := "aws_iot_event_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testEventConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testEventConfigurationBasic(attributeName),
				Check: resource.ComposeTestCheckFunc(
					testEventConfigurationExists(resourceName, &eventConfiguration),
					resource.TestCheckResourceAttr(resourceName, "configurations_map", attributeName),
					resource.TestCheckResourceAttrSet(resourceName, "authorizer_arn"),
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

func TestEventConfigurationFull(t *testing.T) {
	var eventConfiguration iot.DescribeEventConfigurationsOutput
	rString := acctest.RandString(8)
	attributeName := fmt.Sprintf("tf_acc_authorizer_%s", rString)
	resourceName := "aws_iot_event_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testEventConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testEventConfigurationFull(attributeName),
				Check: resource.ComposeTestCheckFunc(
					testEventConfigurationExists(resourceName, &eventConfiguration),
					resource.TestCheckResourceAttr(resourceName, "configurations_map", attributeName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Answer", "42"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Update attribute
				Config: testEventConfigurationFull(attributeName),
				Check: resource.ComposeTestCheckFunc(
					testEventConfigurationExists(resourceName, &eventConfiguration),
					resource.TestCheckResourceAttr(resourceName, "configurations_map", attributeName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Answer", "differentOne"),
				),
			},
			{ // Remove thing type association
				Config: testEventConfigurationBasic(attributeName),
				Check: resource.ComposeTestCheckFunc(
					testEventConfigurationExists(resourceName, &eventConfiguration),
					resource.TestCheckResourceAttr(resourceName, "configurations_map", attributeName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
				),
			},
		},
	})
}

func testEventConfigurationExists(n string, thing *iot.DescribeEventConfigurationsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Evnet Configuration ID is set")
		}

		conn := testAccProvider.Meta().(*iot.IoT)
		params := &iot.DescribeEventConfigurationsInput{}
		resp, err := conn.DescribeEventConfigurations(params)
		if err != nil {
			return err
		}

		*thing = *resp

		return nil
	}
}

func testEventConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*iot.IoT)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_event_configuration" {
			continue
		}

		params := &iot.DescribeEventConfigurationsInput{}

		_, err := conn.DescribeEventConfigurations(params)
		if err != nil {
			return err
		}
		return fmt.Errorf("Expected IoT Event Configuration to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testEventConfigurationBasic(eventConfigurationName string) string {
	return fmt.Sprintf(`
resource "aws_iot_event" "test" {
  name = "%s"
}
`, eventConfigurationName)
}

func testEventConfigurationFull(attributeName string) string {
	return fmt.Sprintf(`
resource "aws_iot_event_configuration" "test" {
	configurations_map {
	  attribute_name = "%s"
	  enabled        = true
	}
}
`, attributeName)
}
