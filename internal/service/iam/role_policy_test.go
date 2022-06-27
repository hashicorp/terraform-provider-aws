package iam_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyExists(
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
				Config: testAccRolePolicyConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyExists(
						roleName,
						resourceName,
						&rolePolicy2,
					),
					testAccCheckRolePolicyExists(
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyExists(
						roleResourceName,
						rolePolicyResourceName,
						&out,
					),
					testAccCheckRolePolicyDisappears(&out),
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyConfig_order(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyExists(
						roleName,
						resourceName,
						&rolePolicy1,
					),
				),
			},
			{
				Config:   testAccRolePolicyConfig_newOrder(rName),
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyConfig_namePrefix(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyExists(
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
				Config: testAccRolePolicyConfig_namePrefix(rName, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyExists(
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyConfig_generatedName(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyExists(
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
				Config: testAccRolePolicyConfig_generatedName(rName, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyExists(
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRolePolicyConfig_invalidJSON(rName),
				ExpectError: regexp.MustCompile("invalid JSON"),
			},
		},
	})
}

func TestAccIAMRolePolicy_Policy_invalidResource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRolePolicyConfig_invalidResource(rName),
				ExpectError: regexp.MustCompile("MalformedPolicyDocument"),
			},
		},
	})
}

func testAccCheckRolePolicyDestroy(s *terraform.State) error {
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

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("Error reading IAM policy %s from role %s: %s", name, role, err)
		}

		if getResp != nil {
			return fmt.Errorf("Found IAM Role, expected none: %s", getResp)
		}
	}

	return nil
}

func testAccCheckRolePolicyDisappears(out *iam.GetRolePolicyOutput) resource.TestCheckFunc {
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

func testAccCheckRolePolicyExists(
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

func testAccRolePolicyConfig_basic(rName string) string {
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

func testAccRolePolicyConfig_namePrefix(rName, action string) string {
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

func testAccRolePolicyConfig_generatedName(rName, policyAction string) string {
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

func testAccRolePolicyConfig_update(rName string) string {
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

func testAccRolePolicyConfig_invalidJSON(rName string) string {
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

func testAccRolePolicyConfig_invalidResource(rName string) string {
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

func testAccRolePolicyConfig_order(rName string) string {
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

func testAccRolePolicyConfig_newOrder(rName string) string {
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
