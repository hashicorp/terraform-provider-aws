package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnectHoursOfOperation_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccHoursOfOperation_basic,
		"disappears": testAccHoursOfOperation_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccCheckHoursOfOperationExists(resourceName string, function *connect.DescribeHoursOfOperationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Hours of Operation not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Hours of Operation ID not set")
		}
		instanceID, hoursOfOperationID, err := tfconnect.HoursOfOperationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeHoursOfOperationInput{
			HoursOfOperationId: aws.String(hoursOfOperationID),
			InstanceId:         aws.String(instanceID),
		}

		getFunction, err := conn.DescribeHoursOfOperation(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccHoursOfOperationBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccHoursOfOperationBasicConfig(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccHoursOfOperationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q
  time_zone   = "EST"

  config {
	day = "MONDAY"

	end_time {
		hours   = 23
		minutes = 08
	}

    start_time {
		hours   = 8
		minutes = 0
	}
  }

  tags = {
    "Name" = "Test Hours of Operation"
  }
}
`, rName2, label))
}
