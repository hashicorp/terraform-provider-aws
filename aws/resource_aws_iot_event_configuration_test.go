package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIoTEventConfiguration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEventConfiguration_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventConfiguration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_event_configuration.example", "name"),
				),
			},
		},
	})
}

func testAccCheckAWSEventConfiguration_basic(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_event_configuration" {
			continue
		}

		// Try to find the Cert
		DescribeEventConfiguration := &iot.DescribeEventConfigurationsInput{}

		_, err := conn.DescribeEventConfigurations(DescribeEventConfiguration)

		// Verify the error is what we want
		if err != nil {
			iotErr, ok := err.(awserr.Error)
			if !ok || iotErr.Code() != "ResourceNotFoundException" {
				return err
			}
		}

	}

	return nil
}

var testAccAWSIoTEventConfiguration = `
resource "aws_iot_event_configuration" "example" {
  configurations_map {
		attribute_name = "THING"
		enabled        = true
	  }
  name = "Example Configuration"
}
`
