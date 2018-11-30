package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIamAssumeRolePolicy_basic(t *testing.T) {
	rName := acctest.RandString(10)
	oldService := "ec2.amazonaws.com"
	newService := "s3.amazonaws.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIamAssumeRolePolicy_roleConfig(rName, oldService),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAssumeRolePolicyServiceMatches(rName, oldService),
				),
			},
			{
				Config: testAccAWSIamAssumeRolePolicyConfig(rName, newService),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAssumeRolePolicyServiceMatches(rName, newService),
				),
			},
		},
	})
}

func testAccAWSIamAssumeRolePolicy_roleConfig(rName string, service string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "%s"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName, service)
}

func testAccAWSIamAssumeRolePolicyConfig(rName string, oldService string, newService string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "%s"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_assume_role_policy" "policy" {
  role_name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "%s"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName, oldService, newService)
}

func testAccCheckAWSAssumeRolePolicyServiceMatches(rName string, service string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		resp, err := iamconn.GetRole(&iam.GetRoleInput{
			RoleName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		actualAssumeRolePolicy := *resp.Role.AssumeRolePolicyDocument
		expectedAssumeRolePolicy := service

		if actualAssumeRolePolicy != expectedAssumeRolePolicy {
			return fmt.Errorf("AssumeRolePolicy: '%q', expected '%q'.", actualAssumeRolePolicy, expectedAssumeRolePolicy)
		}

		return nil
	}
}
