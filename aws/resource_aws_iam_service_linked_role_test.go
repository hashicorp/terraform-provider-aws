package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIAMServiceLinkedRole_basic(t *testing.T) {
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "elasticbeanstalk.amazonaws.com"
	name := "AWSServiceRoleForElasticBeanstalk"
	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMServiceLinkedRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMServiceLinkedRoleConfig(awsServiceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMServiceLinkedRoleExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:iam::[^:]+:role%s%s$", path, name))),
					resource.TestCheckResourceAttr(resourceName, "aws_service_name", awsServiceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "path", path),
					resource.TestCheckResourceAttrSet(resourceName, "unique_id"),
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

func testAccAWSIAMServiceLinkedRoleConfig(awsServiceName string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
	aws_service_name = "%s"
}
`, awsServiceName)
}
