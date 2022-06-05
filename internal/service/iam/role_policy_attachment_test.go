package iam_test

import (
	"fmt"
	"strings"
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

func TestAccIAMRolePolicyAttachment_basic(t *testing.T) {
	var out iam.ListAttachedRolePoliciesOutput
	rInt := sdkacctest.RandInt()
	testPolicy := fmt.Sprintf("tf-acctest-%d", rInt)
	testPolicy2 := fmt.Sprintf("tf-acctest2-%d", rInt)
	testPolicy3 := fmt.Sprintf("tf-acctest3-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyAttachmentExists("aws_iam_role_policy_attachment.test-attach", 1, &out),
					testAccCheckRolePolicyAttachmentAttributes([]string{testPolicy}, &out),
				),
			},
			{
				ResourceName:      "aws_iam_role_policy_attachment.test-attach",
				ImportState:       true,
				ImportStateIdFunc: testAccRolePolicyAttachmentImportStateIdFunc("aws_iam_role_policy_attachment.test-attach"),
				// We do not have a way to align IDs since the Create function uses resource.PrefixedUniqueId()
				// Failed state verification, resource with ID ROLE-POLICYARN not found
				// ImportStateVerify: true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected 1 state: %#v", s)
					}

					rs := s[0]

					if !strings.HasPrefix(rs.Attributes["policy_arn"], "arn:") {
						return fmt.Errorf("expected policy_arn attribute to be set and begin with arn:, received: %s", rs.Attributes["policy_arn"])
					}

					return nil
				},
			},
			{
				Config: testAccRolePolicyAttachUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyAttachmentExists("aws_iam_role_policy_attachment.test-attach", 2, &out),
					testAccCheckRolePolicyAttachmentAttributes([]string{testPolicy2, testPolicy3}, &out),
				),
			},
		},
	})
}

func TestAccIAMRolePolicyAttachment_disappears(t *testing.T) {
	var attachedRolePolicies iam.ListAttachedRolePoliciesOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyAttachmentExists(resourceName, 1, &attachedRolePolicies),
					testAccCheckRolePolicyAttachmentDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMRolePolicyAttachment_Disappears_role(t *testing.T) {
	var attachedRolePolicies iam.ListAttachedRolePoliciesOutput
	var role iam.Role

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_iam_role_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRolePolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(iamRoleResourceName, &role),
					testAccCheckRolePolicyAttachmentExists(resourceName, 1, &attachedRolePolicies),
					// DeleteConflict: Cannot delete entity, must detach all policies first.
					testAccCheckRolePolicyAttachmentDisappears(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceRole(), iamRoleResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRolePolicyAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_role_policy_attachment" {
			continue
		}

		policyARN := rs.Primary.Attributes["policy_arn"]
		role := rs.Primary.Attributes["role"]

		hasPolicyAttachment, err := tfiam.RoleHasPolicyARNAttachment(conn, role, policyARN)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}

		if err != nil {
			return err
		}

		if hasPolicyAttachment {
			return fmt.Errorf("IAM Role (%s) Policy Attachment (%s) still exists", role, policyARN)
		}
	}

	return nil
}

func testAccCheckRolePolicyAttachmentExists(n string, c int, out *iam.ListAttachedRolePoliciesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		role := rs.Primary.Attributes["role"]

		attachedPolicies, err := conn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(role),
		})
		if err != nil {
			return fmt.Errorf("Error: Failed to get attached policies for role %s (%s)", role, n)
		}
		if c != len(attachedPolicies.AttachedPolicies) {
			return fmt.Errorf("Error: Role (%s) has wrong number of policies attached on initial creation", n)
		}

		*out = *attachedPolicies
		return nil
	}
}

func testAccCheckRolePolicyAttachmentAttributes(policies []string, out *iam.ListAttachedRolePoliciesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		matched := 0

		for _, p := range policies {
			for _, ap := range out.AttachedPolicies {
				// *ap.PolicyArn like arn:aws:iam::111111111111:policy/test-policy
				parts := strings.Split(*ap.PolicyArn, "/")
				if len(parts) == 2 && p == parts[1] {
					matched++
				}
			}
		}
		if matched != len(policies) || matched != len(out.AttachedPolicies) {
			return fmt.Errorf("Error: Number of attached policies was incorrect: expected %d matched policies, matched %d of %d", len(policies), matched, len(out.AttachedPolicies))
		}
		return nil
	}
}

func testAccCheckRolePolicyAttachmentDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		policyARN := rs.Primary.Attributes["policy_arn"]
		role := rs.Primary.Attributes["role"]

		return tfiam.DetachPolicyFromRole(conn, role, policyARN)
	}
}

func testAccRolePolicyAttachmentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["role"], rs.Primary.Attributes["policy_arn"]), nil
	}
}

func testAccRolePolicyAttachConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "test-role-%d"

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

resource "aws_iam_policy" "policy" {
  name        = "tf-acctest-%d"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = aws_iam_role.role.name
  policy_arn = aws_iam_policy.policy.arn
}
`, rInt, rInt)
}

func testAccRolePolicyAttachUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "test-role-%d"

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

resource "aws_iam_policy" "policy" {
  name        = "tf-acctest-%d"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy2" {
  name        = "tf-acctest2-%d"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy3" {
  name        = "tf-acctest3-%d"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = aws_iam_role.role.name
  policy_arn = aws_iam_policy.policy2.arn
}

resource "aws_iam_role_policy_attachment" "test-attach2" {
  role       = aws_iam_role.role.name
  policy_arn = aws_iam_policy.policy3.arn
}
`, rInt, rInt, rInt, rInt)
}

func testAccRolePolicyAttachmentConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

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

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AdministratorAccess"
  role       = aws_iam_role.test.name
}
`, rName)
}
