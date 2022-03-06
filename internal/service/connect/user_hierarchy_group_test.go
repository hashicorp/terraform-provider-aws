package connect_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

func testAccCheckUserHierarchyGroupExists(resourceName string, function *connect.DescribeUserHierarchyGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect User Hierarchy Group not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect User Hierarchy Group ID not set")
		}
		instanceID, userHierarchyGroupID, err := tfconnect.UserHierarchyGroupParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeUserHierarchyGroupInput{
			HierarchyGroupId: aws.String(userHierarchyGroupID),
			InstanceId:       aws.String(instanceID),
		}

		getFunction, err := conn.DescribeUserHierarchyGroup(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}
