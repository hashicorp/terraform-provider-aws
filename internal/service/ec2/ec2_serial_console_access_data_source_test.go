package ec2_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2SerialConsoleAccessDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSerialConsoleAccessDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSerialConsoleAccessDataSource("data.aws_ec2_serial_console_access.current"),
				),
			},
		},
	})
}

func testAccCheckSerialConsoleAccessDataSource(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		actual, err := conn.GetSerialConsoleAccessStatus(&ec2.GetSerialConsoleAccessStatusInput{})
		if err != nil {
			return fmt.Errorf("Error reading serial console access toggle: %q", err)
		}

		attr, _ := strconv.ParseBool(rs.Primary.Attributes["enabled"])

		if attr != aws.BoolValue(actual.SerialConsoleAccessEnabled) {
			return fmt.Errorf("Serial console access is not in expected state (%t)", aws.BoolValue(actual.SerialConsoleAccessEnabled))
		}

		return nil
	}
}

const testAccSerialConsoleAccessDataSourceConfig_basic = `
data "aws_ec2_serial_console_access" "current" {}
`
