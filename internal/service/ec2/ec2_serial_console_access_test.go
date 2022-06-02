package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2SerialConsoleAccess_basic(t *testing.T) {
	resourceName := "aws_ec2_serial_console_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSerialConsoleAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSerialConsoleAccessConfig_basic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSerialConsoleAccess(resourceName, false),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSerialConsoleAccessConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSerialConsoleAccess(resourceName, true),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckSerialConsoleAccessDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	response, err := conn.GetSerialConsoleAccessStatus(&ec2.GetSerialConsoleAccessStatusInput{})
	if err != nil {
		return err
	}

	if aws.BoolValue(response.SerialConsoleAccessEnabled) != false {
		return fmt.Errorf("Serial console access not disabled on resource removal")
	}

	return nil
}

func testAccCheckSerialConsoleAccess(n string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		response, err := conn.GetSerialConsoleAccessStatus(&ec2.GetSerialConsoleAccessStatusInput{})
		if err != nil {
			return err
		}

		if aws.BoolValue(response.SerialConsoleAccessEnabled) != enabled {
			return fmt.Errorf("Serial console access is not in expected state (%t)", enabled)
		}

		return nil
	}
}

func testAccSerialConsoleAccessConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ec2_serial_console_access" "test" {
  enabled = %[1]t
}
`, enabled)
}
