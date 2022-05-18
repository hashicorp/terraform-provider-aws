package ssoadmin_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
)

func TestAccSSOAdminAccountAssignment_Basic_group(t *testing.T) {
	resourceName := "aws_ssoadmin_account_assignment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	groupName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckInstances(t)
			testAccPreCheckIdentityStoreGroupName(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAssignmentBasicGroupConfig(groupName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountAssignmentExists(resourceName),
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

func TestAccSSOAdminAccountAssignment_Basic_user(t *testing.T) {
	resourceName := "aws_ssoadmin_account_assignment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckInstances(t)
			testAccPreCheckIdentityStoreUserName(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAssignmentBasicUserConfig(userName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountAssignmentExists(resourceName),
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

func TestAccSSOAdminAccountAssignment_disappears(t *testing.T) {
	resourceName := "aws_ssoadmin_account_assignment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	groupName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckInstances(t)
			testAccPreCheckIdentityStoreGroupName(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAssignmentBasicGroupConfig(groupName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountAssignmentExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssoadmin.ResourceAccountAssignment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountAssignmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssoadmin_account_assignment" {
			continue
		}

		idParts, err := tfssoadmin.ParseAccountAssignmentID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Account Assignment ID (%s): %w", rs.Primary.ID, err)
		}

		principalID := idParts[0]
		principalType := idParts[1]
		targetID := idParts[2]
		permissionSetArn := idParts[4]
		instanceArn := idParts[5]

		accountAssignment, err := tfssoadmin.FindAccountAssignment(conn, principalID, principalType, targetID, permissionSetArn, instanceArn)

		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading SSO Account Assignment for Principal (%s): %w", principalID, err)
		}

		if accountAssignment != nil {
			return fmt.Errorf("SSO Account Assignment for Principal (%s) still exists", principalID)
		}
	}

	return nil
}

func testAccCheckAccountAssignmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn

		idParts, err := tfssoadmin.ParseAccountAssignmentID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Account Assignment ID (%s): %w", rs.Primary.ID, err)
		}

		principalID := idParts[0]
		principalType := idParts[1]
		targetID := idParts[2]
		permissionSetArn := idParts[4]
		instanceArn := idParts[5]

		accountAssignment, err := tfssoadmin.FindAccountAssignment(conn, principalID, principalType, targetID, permissionSetArn, instanceArn)

		if err != nil {
			return err
		}

		if accountAssignment == nil {
			return fmt.Errorf("Account Assignment for Principal (%s) not found", principalID)
		}

		return nil
	}
}

func testAccAccountAssignmentBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_caller_identity" "current" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccAccountAssignmentBasicGroupConfig(groupName, rName string) string {
	return acctest.ConfigCompose(
		testAccAccountAssignmentBaseConfig(rName),
		fmt.Sprintf(`
data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  filter {
    attribute_path  = "DisplayName"
    attribute_value = %q
  }
}

resource "aws_ssoadmin_account_assignment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "GROUP"
  principal_id       = data.aws_identitystore_group.test.group_id
}
`, groupName))
}

func testAccAccountAssignmentBasicUserConfig(userName, rName string) string {
	return acctest.ConfigCompose(
		testAccAccountAssignmentBaseConfig(rName),
		fmt.Sprintf(`
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  filter {
    attribute_path  = "UserName"
    attribute_value = %q
  }
}

resource "aws_ssoadmin_account_assignment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "USER"
  principal_id       = data.aws_identitystore_user.test.user_id
}
`, userName))
}

func testAccPreCheckIdentityStoreGroupName(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME") == "" {
		t.Skip("AWS_IDENTITY_STORE_GROUP_NAME env var must be set for AWS Identity Store Group acceptance test. " +
			"This is required until ListGroups API returns results without filtering by name.")
	}
}

func testAccPreCheckIdentityStoreUserName(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_USER_NAME") == "" {
		t.Skip("AWS_IDENTITY_STORE_USER_NAME env var must be set for AWS Identity Store User acceptance test. " +
			"This is required until ListUsers API returns results without filtering by name.")
	}
}
