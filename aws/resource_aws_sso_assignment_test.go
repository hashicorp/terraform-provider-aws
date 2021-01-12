package aws

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccPreCheckAWSSIdentityStoreGroup(t *testing.T, identityStoreGroup string) {
	if identityStoreGroup == "" {
		t.Skip("skipping acceptance testing: No Identity Store Group was provided")
	}
}

func testAccPreCheckAWSSIdentityStoreUser(t *testing.T, identityStoreUser string) {
	if identityStoreUser == "" {
		t.Skip("skipping acceptance testing: No Identity Store User was provided")
	}
}

func TestAccAWSSSOAssignmentGroup_basic(t *testing.T) {
	var accountAssignment ssoadmin.AccountAssignment
	resourceName := "aws_sso_assignment.example"
	rName := acctest.RandomWithPrefix("tf-sso-test")

	// Read identity store group from environment since they must exist in the caller's identity store
	identityStoreGroup := os.Getenv("AWS_IDENTITY_STORE_GROUP")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSSIdentityStoreGroup(t, identityStoreGroup)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAssignmentBasicGroupConfig(identityStoreGroup, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAssignmentExists(resourceName, &accountAssignment),
					resource.TestCheckResourceAttr(resourceName, "target_type", "AWS_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "GROUP"),
					resource.TestMatchResourceAttr(resourceName, "principal_id", regexp.MustCompile("^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}")),
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

func TestAccAWSSSOAssignmentUser_basic(t *testing.T) {
	var accountAssignment ssoadmin.AccountAssignment
	resourceName := "aws_sso_assignment.example"
	rName := acctest.RandomWithPrefix("tf-sso-test")

	// Read identity store user from environment since they must exist in the caller's identity store
	identityStoreUser := os.Getenv("AWS_IDENTITY_STORE_USER")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSSIdentityStoreUser(t, identityStoreUser)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAssignmentBasicUserConfig(identityStoreUser, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAssignmentExists(resourceName, &accountAssignment),
					resource.TestCheckResourceAttr(resourceName, "target_type", "AWS_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "USER"),
					resource.TestMatchResourceAttr(resourceName, "principal_id", regexp.MustCompile("^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}")),
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

func TestAccAWSSSOAssignmentGroup_disappears(t *testing.T) {
	var accountAssignment ssoadmin.AccountAssignment
	resourceName := "aws_sso_assignment.example"
	rName := acctest.RandomWithPrefix("tf-sso-test")

	// Read identity store group from environment since they must exist in the caller's identity store
	identityStoreGroup := os.Getenv("AWS_IDENTITY_STORE_GROUP")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSSIdentityStoreGroup(t, identityStoreGroup)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAssignmentBasicGroupConfig(identityStoreGroup, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAssignmentExists(resourceName, &accountAssignment),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsoAssignment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})

}

func testAccCheckAWSSSOAssignmentExists(resourceName string, accountAssignment *ssoadmin.AccountAssignment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		// id = ${InstanceID}/${PermissionSetID}/${TargetType}/${TargetID}/${PrincipalType}/${PrincipalID}
		idParts := strings.Split(rs.Primary.ID, "/")
		if len(idParts) != 6 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" || idParts[4] == "" || idParts[5] == "" {
			return fmt.Errorf("Unexpected format of id (%s), expected ${InstanceID}/${PermissionSetID}/${TargetType}/${TargetID}/${PrincipalType}/${PrincipalID}", rs.Primary.ID)
		}

		instanceID := idParts[0]
		permissionSetID := idParts[1]
		targetType := idParts[2]
		targetID := idParts[3]
		principalType := idParts[4]
		principalID := idParts[5]

		// arn:${Partition}:sso:::instance/${InstanceId}
		instanceArn := arn.ARN{
			Partition: testAccProvider.Meta().(*AWSClient).partition,
			Service:   "sso",
			Resource:  fmt.Sprintf("instance/%s", instanceID),
		}.String()

		// arn:${Partition}:sso:::permissionSet/${InstanceId}/${PermissionSetId}
		permissionSetArn := arn.ARN{
			Partition: testAccProvider.Meta().(*AWSClient).partition,
			Service:   "sso",
			Resource:  fmt.Sprintf("permissionSet/%s/%s", instanceID, permissionSetID),
		}.String()

		ssoadminconn := testAccProvider.Meta().(*AWSClient).ssoadminconn

		accountAssignmentResp, getAccountAssignmentErr := resourceAwsSsoAssignmentGet(
			ssoadminconn,
			instanceArn,
			permissionSetArn,
			targetType,
			targetID,
			principalType,
			principalID,
		)
		if getAccountAssignmentErr != nil {
			return getAccountAssignmentErr
		}

		*accountAssignment = *accountAssignmentResp
		return nil
	}
}

func testAccCheckAWSSSOAssignmentDestroy(s *terraform.State) error {
	ssoadminconn := testAccProvider.Meta().(*AWSClient).ssoadminconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sso_assignment" {
			continue
		}

		// id = ${InstanceID}/${PermissionSetID}/${TargetType}/${TargetID}/${PrincipalType}/${PrincipalID}
		idParts := strings.Split(rs.Primary.ID, "/")
		if len(idParts) != 6 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" || idParts[4] == "" || idParts[5] == "" {
			return fmt.Errorf("Unexpected format of id (%s), expected ${InstanceID}/${PermissionSetID}/${TargetType}/${TargetID}/${PrincipalType}/${PrincipalID}", rs.Primary.ID)
		}

		instanceID := idParts[0]
		permissionSetID := idParts[1]
		targetType := idParts[2]
		targetID := idParts[3]
		principalType := idParts[4]
		principalID := idParts[5]

		// arn:${Partition}:sso:::instance/${InstanceId}
		instanceArn := arn.ARN{
			Partition: testAccProvider.Meta().(*AWSClient).partition,
			Service:   "sso",
			Resource:  fmt.Sprintf("instance/%s", instanceID),
		}.String()

		// arn:${Partition}:sso:::permissionSet/${InstanceId}/${PermissionSetId}
		permissionSetArn := arn.ARN{
			Partition: testAccProvider.Meta().(*AWSClient).partition,
			Service:   "sso",
			Resource:  fmt.Sprintf("permissionSet/%s/%s", instanceID, permissionSetID),
		}.String()

		accountAssignment, getAccountAssignmentErr := resourceAwsSsoAssignmentGet(
			ssoadminconn,
			instanceArn,
			permissionSetArn,
			targetType,
			targetID,
			principalType,
			principalID,
		)

		if isAWSErr(getAccountAssignmentErr, "ResourceNotFoundException", "") {
			continue
		}

		if getAccountAssignmentErr != nil {
			return getAccountAssignmentErr
		}

		if accountAssignment != nil {
			return fmt.Errorf("AWS SSO Account Assignment (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccSSOAssignmentBasicGroupConfig(identityStoreGroup, rName string) string {
	return fmt.Sprintf(`
data "aws_sso_instance" "selected" {}

data "aws_caller_identity" "current" {}

data "aws_identity_store_group" "example_group" {
  identity_store_id = data.aws_sso_instance.selected.identity_store_id
  display_name      = "%s"
}

resource "aws_sso_permission_set" "example" {
  name                = "%s"
  description         = "testing"
  instance_arn        = data.aws_sso_instance.selected.arn
  managed_policy_arns = ["arn:aws:iam::aws:policy/ReadOnlyAccess"]
}

resource "aws_sso_assignment" "example" {
  instance_arn       = data.aws_sso_instance.selected.arn
  permission_set_arn = aws_sso_permission_set.example.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "GROUP"
  principal_id       = data.aws_identity_store_group.example_group.group_id
}
`, identityStoreGroup, rName)
}

func testAccSSOAssignmentBasicUserConfig(identityStoreUser, rName string) string {
	return fmt.Sprintf(`
data "aws_sso_instance" "selected" {}

data "aws_caller_identity" "current" {}

data "aws_identity_store_user" "example_user" {
  identity_store_id = data.aws_sso_instance.selected.identity_store_id
  user_name         = "%s"
}

resource "aws_sso_permission_set" "example" {
  name                = "%s"
  description         = "testing"
  instance_arn        = data.aws_sso_instance.selected.arn
  managed_policy_arns = ["arn:aws:iam::aws:policy/ReadOnlyAccess"]
}

resource "aws_sso_assignment" "example" {
  instance_arn       = data.aws_sso_instance.selected.arn
  permission_set_arn = aws_sso_permission_set.example.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "USER"
  principal_id       = data.aws_identity_store_user.example_user.user_id
}
`, identityStoreUser, rName)
}
