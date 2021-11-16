package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
)

func TestAccIoTThingGroupMembership_basic(t *testing.T) {
	rString := sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckThingGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupMembershipConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_thing_group_membership.test_attachment", "thing_name", fmt.Sprintf("test_thing_%s", rString)),
					resource.TestCheckResourceAttr("aws_iot_thing_group_membership.test_attachment", "thing_group_name", fmt.Sprintf("test_group_%s", rString)),
					resource.TestCheckResourceAttr("aws_iot_thing_group_membership.test_attachment", "override_dynamics_group", "false"),
					testAccCheckThingGroupMembershipExists(rString),
				),
			},
			{
				ResourceName:      "aws_iot_thing_group_membership.test_attachment",
				ImportStateIdFunc: testAccCheckThingGroupMembershipImportStateIdFunc("aws_iot_thing_group_membership.test_attachment"),
				ImportState:       true,
				// We do not have a way to align IDs since the Create function uses resource.PrefixedUniqueId()
				// Failed state verification, resource with ID ROLE-POLICYARN not found
				// ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckThingGroupMembershipExists(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_thing_group_membership" {
				continue
			}

			thingName := rs.Primary.Attributes["thing_name"]
			thingGroupName := rs.Primary.Attributes["thing_group_name"]
			hasThingGroup, err := tfiot.IotThingHasThingGroup(conn, thingName, thingGroupName, "")

			if err != nil {
				return err
			}

			if !hasThingGroup {
				return fmt.Errorf("IoT Thing (%s) is not in IoT Thing Group (%s)", thingName, thingGroupName)
			}

			return nil
		}
		return nil
	}
}

func testAccCheckThingGroupMembershipDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_group_membership" {
			continue
		}

		thingName := rs.Primary.Attributes["thing_name"]
		thingGroupName := rs.Primary.Attributes["thing_group_name"]

		hasThingGroup, err := tfiot.IotThingHasThingGroup(conn, thingName, thingGroupName, "")

		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			return nil
		}

		if err != nil {
			return err
		}

		if hasThingGroup {
			return fmt.Errorf("IoT Thing (%s) still in IoT Thing Group (%s)", thingName, thingGroupName)
		}
	}
	return nil
}

func testAccCheckThingGroupMembershipImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["thing_name"], rs.Primary.Attributes["thing_group_name"]), nil
	}
}

func testAccThingGroupMembershipConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing" "test_thing" {
	name = "test_thing_%s"
}

resource "aws_iot_thing_group" "test_thing_group" {
	name = "test_group_%[1]s"
	properties {
		attributes = {
			"attr1": "val1",
		}
		merge = false
	}
}

resource "aws_iot_thing_group_membership" "test_attachment" {
	thing_name = "${aws_iot_thing.test_thing.name}"
	thing_group_name = "${aws_iot_thing_group.test_thing_group.name}"
	override_dynamics_group = false
}
`, rString)
}
