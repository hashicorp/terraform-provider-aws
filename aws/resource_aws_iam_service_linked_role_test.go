package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIAMServiceLinkedRole_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMServiceLinkedRoleDestroy,
		Steps: []resource.TestStep{
			{
				ResourceName: "aws_iam_service_linked_role.test",
				Config:       fmt.Sprintf(testAccAWSIAMServiceLinkedRoleConfig, "elasticbeanstalk.amazonaws.com"),
				Check:        testAccCheckAWSIAMServiceLinkedRoleExists("aws_iam_service_linked_role.test"),
			},
			{
				ResourceName:      "aws_iam_service_linked_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSIAMServiceLinkedRoleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_service_linked_role" {
			continue
		}

		arnSplit := strings.Split(rs.Primary.ID, "/")
		roleName := arnSplit[len(arnSplit)-1]

		params := &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		}

		_, err := conn.GetRole(params)

		if err == nil {
			return fmt.Errorf("Service-Linked Role still exists: %q", rs.Primary.ID)
		}

		if !isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return err
		}
	}

	return nil

}

func testAccCheckAWSIAMServiceLinkedRoleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		arnSplit := strings.Split(rs.Primary.ID, "/")
		roleName := arnSplit[len(arnSplit)-1]

		params := &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		}

		_, err := conn.GetRole(params)

		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				return fmt.Errorf("Service-Linked Role doesn't exists: %q", rs.Primary.ID)
			}
			return err
		}

		return nil
	}
}

const testAccAWSIAMServiceLinkedRoleConfig = `
resource "aws_iam_service_linked_role" "test" {
	aws_service_name = "%s"
}
`
