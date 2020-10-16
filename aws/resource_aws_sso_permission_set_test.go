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

func testAccPreCheckAWSSSOPermissionSet(t *testing.T) {
	ssoadminconn := testAccProvider.Meta().(*AWSClient).ssoadminconn

	input := &ssoadmin.ListInstancesInput{}

	_, err := ssoadminconn.ListInstances(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func TestAccAWSSSOPermissionSet_basic(t *testing.T) {
	var permissionSet, updatedPermissionSet ssoadmin.PermissionSet
	resourceName := "aws_sso_permission_set.example"
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOPermissionSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOPermissionSetBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOPermissionSetExists(resourceName, &permissionSet),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "managed_policy_arns.*", "arn:aws:iam::aws:policy/ReadOnlyAccess"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("Test_Permission_Set_%s", name)),
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
				Config: testAccAWSSSOPermissionSetBasicConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOPermissionSetExists(resourceName, &updatedPermissionSet),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "managed_policy_arns.*", "arn:aws:iam::aws:policy/ReadOnlyAccess"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "managed_policy_arns.*", "arn:aws:iam::aws:policy/job-function/ViewOnlyAccess"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("Test_Permission_Set_Update_%s", name)),
					resource.TestCheckResourceAttr(resourceName, "description", "Just a test update"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

		if isAWSErr(err, "ResourceNotFoundException", "") {
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

func testAccAWSSSOPermissionSetBasicConfig(rName string) string {
	return fmt.Sprintf(`
	data "aws_caller_identity" "current" {}

	data "aws_sso_instance" "selected" { }
	
	resource "aws_sso_permission_set" "example" {
		name = "Test_Permission_Set_%s"
		description = "Just a test"
		instance_arn       = data.aws_sso_instance.selected.arn
		managed_policy_arns = [
		"arn:aws:iam::aws:policy/ReadOnlyAccess",
	  ]
	}
`, rName)
}

func testAccAWSSSOPermissionSetBasicConfigUpdated(rName string) string {
	return fmt.Sprintf(`
	data "aws_caller_identity" "current" {}

	data "aws_sso_instance" "selected" { }
	
	resource "aws_sso_permission_set" "example" {
		name = "Test_Permission_Set_Update_%s"
		description = "Just a test update"
		instance_arn       = data.aws_sso_instance.selected.arn
		managed_policy_arns = [
		"arn:aws:iam::aws:policy/ReadOnlyAccess",
		"arn:aws:iam::aws:policy/job-function/ViewOnlyAccess",
	  ]
	}
`, rName)
}
