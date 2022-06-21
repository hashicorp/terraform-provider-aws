package ssoadmin_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
)

func TestAccSSOAdminPermissionSetInlinePolicy_basic(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set_inline_policy.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetInlinePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetInlinePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionSetInlinePolicyExists(resourceName),
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

func TestAccSSOAdminPermissionSetInlinePolicy_update(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set_inline_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetInlinePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetInlinePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionSetInlinePolicyExists(resourceName),
				),
			},
			{
				Config: testAccPermissionSetInlinePolicyConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionSetInlinePolicyExists(resourceName),
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

func TestAccSSOAdminPermissionSetInlinePolicy_disappears(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set_inline_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetInlinePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetInlinePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionSetInlinePolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssoadmin.ResourcePermissionSetInlinePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminPermissionSetInlinePolicy_Disappears_permissionSet(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set_inline_policy.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetInlinePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetInlinePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionSetInlinePolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssoadmin.ResourcePermissionSet(), permissionSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPermissionSetInlinePolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssoadmin_permission_set_inline_policy" {
			continue
		}

		permissionSetArn, instanceArn, err := tfssoadmin.ParseResourceID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Permission Set Inline Policy ID (%s): %w", rs.Primary.ID, err)
		}

		input := &ssoadmin.GetInlinePolicyForPermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(permissionSetArn),
		}

		output, err := conn.GetInlinePolicyForPermissionSet(input)
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output == nil {
			continue
		}

		// SSO API returns empty string when removed from Permission Set
		if aws.StringValue(output.InlinePolicy) == "" {
			continue
		}

		return fmt.Errorf("Inline Policy for SSO PermissionSet (%s) still exists", permissionSetArn)
	}

	return nil
}

func testAccCheckPermissionSetInlinePolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		permissionSetArn, instanceArn, err := tfssoadmin.ParseResourceID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Permission Set Inline Policy ID (%s): %w", rs.Primary.ID, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn

		input := &ssoadmin.GetInlinePolicyForPermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(permissionSetArn),
		}

		output, err := conn.GetInlinePolicyForPermissionSet(input)
		if err != nil {
			return err
		}

		if output == nil || output.InlinePolicy == nil {
			return fmt.Errorf("Inline Policy for SSO Permission Set (%s) not found", permissionSetArn)
		}

		return nil
	}
}

func testAccPermissionSetInlinePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
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

func testAccPermissionSetInlinePolicyConfig_update(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
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
