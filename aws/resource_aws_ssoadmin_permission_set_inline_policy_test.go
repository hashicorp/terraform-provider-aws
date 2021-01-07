package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ssoadmin/finder"
)

func TestAccAWSSSOAdminPermissionSetInlinePolicy_basic(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set_inline_policy.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetInlinePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAdminPermissionSetInlinePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminPermissionSetInlinePolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_set_arn", permissionSetResourceName, "arn"),
					resource.TestMatchResourceAttr(resourceName, "inline_policy", regexp.MustCompile("s3:ListAllMyBuckets")),
					resource.TestMatchResourceAttr(resourceName, "inline_policy", regexp.MustCompile("s3:GetBucketLocation")),
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

func TestAccAWSSSOAdminPermissionSetInlinePolicy_update(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set_inline_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetInlinePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAdminPermissionSetInlinePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminPermissionSetInlinePolicyExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSSOAdminPermissionSetInlinePolicyUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminPermissionSetInlinePolicyExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "inline_policy", regexp.MustCompile("s3:ListAllMyBuckets")),
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

func TestAccAWSSSOAdminPermissionSetInlinePolicy_disappears(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set_inline_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetInlinePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAdminPermissionSetInlinePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminPermissionSetInlinePolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsoAdminPermissionSetInlinePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSOInlinePolicy_disappears_permissionSet(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set_inline_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetInlinePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAdminPermissionSetInlinePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminPermissionSetInlinePolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsoAdminPermissionSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSSOAdminPermissionSetInlinePolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssoadminconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssoadmin_permission_set_inline_policy" {
			continue
		}

		permissionSetArn, instanceArn, err := parseSsoAdminResourceID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Permission Set Inline Policy ID (%s): %w", rs.Primary.ID, err)
		}

		policy, err := finder.InlinePolicy(conn, instanceArn, permissionSetArn)

		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if aws.StringValue(policy) == "" {
			continue
		}

		return fmt.Errorf("Inline Policy for SSO PermissionSet (%s) still exists", permissionSetArn)
	}

	return nil
}

func testAccCheckAWSSSOAdminPermissionSetInlinePolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		permissionSetArn, instanceArn, err := parseSsoAdminResourceID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Permission Set Inline Policy ID (%s): %w", rs.Primary.ID, err)
		}

		conn := testAccProvider.Meta().(*AWSClient).ssoadminconn

		policy, err := finder.InlinePolicy(conn, instanceArn, permissionSetArn)

		if err != nil {
			return err
		}

		if policy == nil {
			return fmt.Errorf("Inline Policy for SSO Permission Set (%s) not found", permissionSetArn)
		}

		return nil
	}
}

func testAccSSOAdminPermissionSetInlinePolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "arn:aws:s3:::*",
    ]
  }
}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_permission_set_inline_policy" "test" {
  inline_policy      = data.aws_iam_policy_document.test.json
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`, rName)
}

func testAccSSOAdminPermissionSetInlinePolicyUpdateConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
    ]

    resources = [
      "arn:aws:s3:::*",
    ]
  }
}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_permission_set_inline_policy" "test" {
  inline_policy      = data.aws_iam_policy_document.test.json
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`, rName)
}
