package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAWSSSOPermissionSet_Basic(t *testing.T) {
	var permissionSet, updatedPermissionSet ssoadmin.PermissionSet
	resourceName := "aws_sso_permission_set.example"
	rName := acctest.RandomWithPrefix("tf-sso-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOInstance(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOPermissionSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOPermissionSetExists(resourceName, &permissionSet),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "managed_policy_arns.*", "arn:aws:iam::aws:policy/ReadOnlyAccess"), // lintignore:AWSAT005
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Just a test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSSOPermissionSetBasicConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOPermissionSetExists(resourceName, &updatedPermissionSet),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "managed_policy_arns.*", "arn:aws:iam::aws:policy/ReadOnlyAccess"),              // lintignore:AWSAT005
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "managed_policy_arns.*", "arn:aws:iam::aws:policy/job-function/ViewOnlyAccess"), // lintignore:AWSAT005
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Just a test update"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSSSOPermissionSet_Disappears(t *testing.T) {
	var permissionSet ssoadmin.PermissionSet
	resourceName := "aws_sso_permission_set.example"
	rName := acctest.RandomWithPrefix("tf-sso-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOInstance(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOPermissionSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOPermissionSetExists(resourceName, &permissionSet),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsoPermissionSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSOPermissionSet_Tags(t *testing.T) {
	var permissionSet ssoadmin.PermissionSet
	resourceName := "aws_sso_permission_set.example"
	rName := acctest.RandomWithPrefix("tf-sso-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOInstance(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOPermissionSetConfigTagsSingle(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOPermissionSetExists(resourceName, &permissionSet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSSOPermissionSetConfigTagsMultiple(rName, "key1", "updatedvalue1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOPermissionSetExists(resourceName, &permissionSet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updatedvalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSSOPermissionSetConfigTagsSingle(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOPermissionSetExists(resourceName, &permissionSet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSSSOPermissionSetExists(resourceName string, permissionSet *ssoadmin.PermissionSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		instanceArn, err := resourceAwsSsoPermissionSetParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		ssoadminconn := testAccProvider.Meta().(*AWSClient).ssoadminconn

		permissionSetResp, permissionSetErr := ssoadminconn.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(rs.Primary.ID),
		})

		if permissionSetErr != nil {
			return permissionSetErr
		}

		if *permissionSetResp.PermissionSet.PermissionSetArn == rs.Primary.ID {
			*permissionSet = *permissionSetResp.PermissionSet
			return nil
		}

		return fmt.Errorf("AWS SSO Permission Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSSSOPermissionSetDestroy(s *terraform.State) error {
	ssoadminconn := testAccProvider.Meta().(*AWSClient).ssoadminconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sso_permission_set" {
			continue
		}

		idFormatErr := fmt.Errorf("Unexpected format of ARN (%s), expected arn:${Partition}:sso:::permissionSet/${InstanceId}/${PermissionSetId}", rs.Primary.ID)
		permissionSetArn, err := arn.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}

		resourceParts := strings.Split(permissionSetArn.Resource, "/")
		if len(resourceParts) != 3 || resourceParts[0] != "permissionSet" || resourceParts[1] == "" || resourceParts[2] == "" {
			return idFormatErr
		}

		// resourceParts = ["permissionSet","ins-123456A", "ps-56789B"]
		instanceArn := arn.ARN{
			Partition: permissionSetArn.Partition,
			Service:   permissionSetArn.Service,
			Resource:  fmt.Sprintf("instance/%s", resourceParts[1]),
		}.String()

		input := &ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(rs.Primary.ID),
		}

		output, err := ssoadminconn.DescribePermissionSet(input)

		if isAWSErr(err, ssoadmin.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("AWS SSO Permission Set (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccSSOPermissionSetBasicConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_sso_instance" "selected" {}

resource "aws_sso_permission_set" "example" {
  name                = "%s"
  description         = "Just a test"
  instance_arn        = data.aws_sso_instance.selected.arn
  managed_policy_arns = ["arn:aws:iam::aws:policy/ReadOnlyAccess"]
}
`, rName) // lintignore:AWSAT005
}

func testAccSSOPermissionSetBasicConfigUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_sso_instance" "selected" {}

resource "aws_sso_permission_set" "example" {
  name         = "%s"
  description  = "Just a test update"
  instance_arn = data.aws_sso_instance.selected.arn
  managed_policy_arns = [
    "arn:aws:iam::aws:policy/ReadOnlyAccess",
    "arn:aws:iam::aws:policy/job-function/ViewOnlyAccess"
  ]
}
`, rName) // lintignore:AWSAT005
}

func testAccSSOPermissionSetConfigTagsSingle(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_sso_instance" "selected" {}

resource "aws_sso_permission_set" "example" {
  name                = "%s"
  description         = "Just a test"
  instance_arn        = data.aws_sso_instance.selected.arn
  managed_policy_arns = ["arn:aws:iam::aws:policy/ReadOnlyAccess"]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1) // lintignore:AWSAT005
}

func testAccSSOPermissionSetConfigTagsMultiple(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_sso_instance" "selected" {}

resource "aws_sso_permission_set" "example" {
  name                = "%s"
  description         = "Just a test"
  instance_arn        = data.aws_sso_instance.selected.arn
  managed_policy_arns = ["arn:aws:iam::aws:policy/ReadOnlyAccess"]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2) // lintignore:AWSAT005
}
