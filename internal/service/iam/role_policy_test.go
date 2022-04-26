package iam_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func TestAccIAMRolePolicy_basic(t *testing.T) {
	var rolePolicy1, rolePolicy2, rolePolicy3 iam.GetRolePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy.test"
	resourceName2 := "aws_iam_role_policy.test2"
	roleName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMRolePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						roleName,
						resourceName,
						&rolePolicy1,
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMRolePolicyConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						roleName,
						resourceName,
						&rolePolicy2,
					),
					testAccCheckIAMRolePolicyExists(
						roleName,
						resourceName2,
						&rolePolicy3,
					),
					testAccCheckRolePolicyNameMatches(&rolePolicy1, &rolePolicy2),
					testAccCheckRolePolicyNameChanged(&rolePolicy1, &rolePolicy3),
				),
			},
		},
	})
}

func TestAccIAMRolePolicy_disappears(t *testing.T) {
	var out iam.GetRolePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleResourceName := "aws_iam_role.test"
	rolePolicyResourceName := "aws_iam_role_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						roleResourceName,
						rolePolicyResourceName,
						&out,
					),
					testAccCheckIAMRolePolicyDisappears(&out),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMRolePolicy_policyOrder(t *testing.T) {
	var rolePolicy1 iam.GetRolePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy.test"
	roleName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMRolePolicyOrderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						roleName,
						resourceName,
						&rolePolicy1,
					),
				),
			},
			{
				Config:   testAccIAMRolePolicyNewOrderConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccIAMRolePolicy_namePrefix(t *testing.T) {
	var rolePolicy1, rolePolicy2 iam.GetRolePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy.test"
	roleName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMRolePolicyConfig_namePrefix(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						roleName,
						resourceName,
						&rolePolicy1,
					),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccIAMRolePolicyConfig_namePrefix(rName, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						roleName,
						resourceName,
						&rolePolicy2,
					),
					testAccCheckRolePolicyNameMatches(&rolePolicy1, &rolePolicy2),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
		},
	})
}

func TestAccIAMRolePolicy_generatedName(t *testing.T) {
	var rolePolicy1, rolePolicy2 iam.GetRolePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy.test"
	roleName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMRolePolicyConfig_generatedName(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						roleName,
						resourceName,
						&rolePolicy1,
					),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMRolePolicyConfig_generatedName(rName, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						roleName,
						resourceName,
						&rolePolicy2,
					),
					testAccCheckRolePolicyNameMatches(&rolePolicy1, &rolePolicy2),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
		},
	})
}

func TestAccIAMRolePolicy_invalidJSON(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccIAMRolePolicyConfig_invalidJSON(rName),
				ExpectError: regexp.MustCompile("invalid JSON"),
			},
		},
	})
}

func TestAccIAMRolePolicy_Policy_invalidResource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccIAMRolePolicyConfig_Policy_InvalidResource(rName),
				ExpectError: regexp.MustCompile("MalformedPolicyDocument"),
			},
		},
	})
}

func testAccCheckIAMRolePolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_role_policy" {
			continue
		}

		role, name, err := tfiam.RolePolicyParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		request := &iam.GetRolePolicyInput{
			PolicyName: aws.String(name),
			RoleName:   aws.String(role),
		}

		getResp, err := conn.GetRolePolicy(request)
		if err != nil {
			if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
				// none found, that's good
				return nil
			}
			return fmt.Errorf("Error reading IAM policy %s from role %s: %s", name, role, err)
		}

		if getResp != nil {
			return fmt.Errorf("Found IAM Role, expected none: %s", getResp)
		}
	}

	return nil
}

func testAccCheckIAMRolePolicyDisappears(out *iam.GetRolePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		params := &iam.DeleteRolePolicyInput{
			PolicyName: out.PolicyName,
			RoleName:   out.RoleName,
		}

		_, err := conn.DeleteRolePolicy(params)
		return err
	}
}

func testAccCheckIAMRolePolicyExists(
	iamRoleResource string,
	iamRolePolicyResource string,
	rolePolicy *iam.GetRolePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[iamRoleResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamRoleResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[iamRolePolicyResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamRolePolicyResource)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		role, name, err := tfiam.RolePolicyParseID(policy.Primary.ID)
		if err != nil {
			return err
		}

		output, err := conn.GetRolePolicy(&iam.GetRolePolicyInput{
			RoleName:   aws.String(role),
			PolicyName: aws.String(name),
		})
		if err != nil {
			return err
		}

		*rolePolicy = *output

		return nil
	}
}

func testAccCheckRolePolicyNameChanged(i, j *iam.GetRolePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) == aws.StringValue(j.PolicyName) {
			return errors.New("IAM Role Policy name did not change")
		}

		return nil
	}
}

func testAccCheckRolePolicyNameMatches(i, j *iam.GetRolePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) != aws.StringValue(j.PolicyName) {
			return errors.New("IAM Role Policy name did not match")
		}

		return nil
	}
}

func testAccRolePolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}
`, rName)
}

func testAccIAMRolePolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}
`, rName)
}

func testAccIAMRolePolicyConfig_namePrefix(rName, action string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name_prefix = "tf_test_policy_"
  role        = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": %[2]q,
    "Resource": "*"
  }
}
EOF
}
`, rName, action)
}

func testAccIAMRolePolicyConfig_generatedName(rName, policyAction string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": %[2]q,
    "Resource": "*"
  }
}
EOF
}
`, rName, policyAction)
}

func testAccIAMRolePolicyConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_role_policy" "test2" {
  name = "%[1]s-2"
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}
`, rName)
}

func testAccIAMRolePolicyConfig_invalidJSON(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
  EOF
}
`, rName)
}

func testAccIAMRolePolicyConfig_Policy_InvalidResource(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = jsonencode({
    Statement = [{
      Effect   = "Allow"
      Action   = "*"
      Resource = [["*"]]
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccIAMRolePolicyOrderConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": [
      "ec2:DescribeScheduledInstances",
      "ec2:DescribeScheduledInstanceAvailability",
      "ec2:DescribeFastSnapshotRestores",
      "ec2:DescribeElasticGpus"
    ],
    "Resource": "*"
  }
}
EOF
}
`, rName)
}

func testAccIAMRolePolicyNewOrderConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
	"Action": [
      "ec2:DescribeFastSnapshotRestores",
      "ec2:DescribeScheduledInstanceAvailability",
      "ec2:DescribeScheduledInstances",
      "ec2:DescribeElasticGpus"
    ],
    "Resource": "*"
  }
}
EOF
}
`, rName)
}
