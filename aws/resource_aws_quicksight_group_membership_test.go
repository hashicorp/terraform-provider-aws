package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/quicksight/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAWSQuickSightGroupMembership_basic(t *testing.T) {
	groupName := sdkacctest.RandomWithPrefix("tf-acc-test")
	memberName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_group_membership.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy: testAccCheckQuickSightGroupMembershipDestroy,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightGroupMembershipConfig(groupName, memberName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightGroupMembershipExists(resourceName),
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

func TestAccAWSQuickSightGroupMembership_disappears(t *testing.T) {
	groupName := sdkacctest.RandomWithPrefix("tf-acc-test")
	memberName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_group_membership.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, quicksight.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckQuickSightGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightGroupMembershipConfig(groupName, memberName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightGroupMembershipExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsQuickSightGroupMembership(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQuickSightGroupMembershipDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_group_membership" {
			continue
		}
		awsAccountID, namespace, groupName, userName, err := resourceAwsQuickSightGroupMembershipParseID(rs.Primary.ID)
		if err != nil {
			return err
		}
		listInput := &quicksight.ListGroupMembershipsInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			GroupName:    aws.String(groupName),
		}

		found, err := finder.GroupMembership(conn, listInput, userName)

		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		if found {
			return fmt.Errorf("QuickSight Group (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckQuickSightGroupMembershipExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, namespace, groupName, userName, err := resourceAwsQuickSightGroupMembershipParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn

		listInput := &quicksight.ListGroupMembershipsInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			GroupName:    aws.String(groupName),
		}

		found, err := finder.GroupMembership(conn, listInput, userName)
		if err != nil {
			return fmt.Errorf("Error listing QuickSight Group Memberships: %s", err)
		}

		if !found {
			return fmt.Errorf("QuickSight Group Membership (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSQuickSightGroupMembershipConfig(groupName string, memberName string) string {
	return acctest.ConfigCompose(
		testAccAWSQuickSightGroupConfig(groupName),
		testAccAWSQuickSightUserConfig(memberName),
		fmt.Sprintf(`
resource "aws_quicksight_group_membership" "default" {
  group_name  = aws_quicksight_group.default.group_name
  member_name = aws_quicksight_user.%s.user_name
}
`, memberName))
}
