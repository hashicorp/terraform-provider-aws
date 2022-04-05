package iam_test

import (
	"errors"
	"fmt"
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

func TestAccIAMGroupPolicy_basic(t *testing.T) {
	var groupPolicy1, groupPolicy2 iam.GetGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.group",
						"aws_iam_group_policy.foo",
						&groupPolicy1,
					),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMGroupPolicyConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.group",
						"aws_iam_group_policy.bar",
						&groupPolicy2,
					),
					testAccCheckGroupPolicyNameChanged(&groupPolicy1, &groupPolicy2),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIAMGroupPolicy_disappears(t *testing.T) {
	var out iam.GetGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.group",
						"aws_iam_group_policy.foo",
						&out,
					),
					testAccCheckIAMGroupPolicyDisappears(&out),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMGroupPolicy_namePrefix(t *testing.T) {
	var groupPolicy1, groupPolicy2 iam.GetGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupPolicyConfig_namePrefix(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy1,
					),
				),
			},
			{
				Config: testAccIAMGroupPolicyConfig_namePrefix(rName, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy2,
					),
					testAccCheckGroupPolicyNameMatches(&groupPolicy1, &groupPolicy2),
				),
			},
			{
				ResourceName:            "aws_iam_group_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccIAMGroupPolicy_generatedName(t *testing.T) {
	var groupPolicy1, groupPolicy2 iam.GetGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupPolicyConfig_generatedName(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy1,
					),
				),
			},
			{
				Config: testAccIAMGroupPolicyConfig_generatedName(rName, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy2,
					),
					testAccCheckGroupPolicyNameMatches(&groupPolicy1, &groupPolicy2),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIAMGroupPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_group_policy" {
			continue
		}

		group, name, err := tfiam.GroupPolicyParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		request := &iam.GetGroupPolicyInput{
			PolicyName: aws.String(name),
			GroupName:  aws.String(group),
		}

		getResp, err := conn.GetGroupPolicy(request)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				// none found, that's good
				continue
			}
			return fmt.Errorf("Error reading IAM policy %s from group %s: %s", name, group, err)
		}

		if getResp != nil {
			return fmt.Errorf("Found IAM group policy, expected none: %s", getResp)
		}
	}

	return nil
}

func testAccCheckIAMGroupPolicyDisappears(out *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		params := &iam.DeleteGroupPolicyInput{
			PolicyName: out.PolicyName,
			GroupName:  out.GroupName,
		}

		_, err := conn.DeleteGroupPolicy(params)
		return err
	}
}

func testAccCheckIAMGroupPolicyExists(
	iamGroupResource string,
	iamGroupPolicyResource string,
	groupPolicy *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[iamGroupResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamGroupResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[iamGroupPolicyResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamGroupPolicyResource)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		group, name, err := tfiam.GroupPolicyParseID(policy.Primary.ID)
		if err != nil {
			return err
		}

		output, err := conn.GetGroupPolicy(&iam.GetGroupPolicyInput{
			GroupName:  aws.String(group),
			PolicyName: aws.String(name),
		})

		if err != nil {
			return err
		}

		*groupPolicy = *output

		return nil
	}
}

func testAccCheckGroupPolicyNameChanged(i, j *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) == aws.StringValue(j.PolicyName) {
			return errors.New("IAM Group Policy name did not change")
		}

		return nil
	}
}

func testAccCheckGroupPolicyNameMatches(i, j *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) != aws.StringValue(j.PolicyName) {
			return errors.New("IAM Group Policy name did not match")
		}

		return nil
	}
}

func testAccIAMGroupPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "foo" {
  name  = %[1]q
  group = aws_iam_group.group.name

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

func testAccIAMGroupPolicyConfig_namePrefix(rName, policyAction string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "test" {
  name_prefix = %[1]q
  group       = aws_iam_group.test.name

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

func testAccIAMGroupPolicyConfig_generatedName(rName, policyAction string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "test" {
  group = aws_iam_group.test.name

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

func testAccIAMGroupPolicyConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "foo" {
  name  = %[1]q
  group = aws_iam_group.group.name

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

resource "aws_iam_group_policy" "bar" {
  name  = "%[1]s-2"
  group = aws_iam_group.group.name

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
