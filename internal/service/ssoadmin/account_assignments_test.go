package ssoadmin_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
)

func TestAccSSOAdminAccountAssignments_Basic_group(t *testing.T) {
	resourceName := "aws_ssoadmin_account_assignments.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	groupName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckInstances(t)
			testAccPreCheckIdentityStoreGroupName(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAssignmentBasicGroupConfig(groupName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountAssignmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "AWS_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "GROUP"),
					resource.TestMatchResourceAttr(resourceName, "principal_ids", "PRINCIPAL_IDS"),
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

func TestAccSSOAdminAccountAssignments_Basic_user(t *testing.T) {
	resourceName := "aws_ssoadmin_account_assignments.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckInstances(t)
			testAccPreCheckIdentityStoreUserName(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAssignmentBasicUserConfig(userName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountAssignmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "AWS_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "USER"),
					resource.TestMatchResourceAttr(resourceName, "principal_ids", "PRINCIPAL_IDS"),
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

func testAccAccountAssignmentsBaseConfig(rName string) string {
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
		testAccAccountAssignmentsBaseConfig(rName),
		fmt.Sprintf(`
data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  filter {
    attribute_path  = "DisplayName"
    attribute_value = %q
  }
}

resource "aws_ssoadmin_account_assignments" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "GROUP"
  principal_ids      = [data.aws_identitystore_group.test.group_id]
}
`, groupName))
}

func testAccAccountAssignmentsBasicUserConfig(userName, rName string) string {
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

resource "aws_ssoadmin_account_assignments" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "USER"
  principal_ids      = [data.aws_identitystore_user.test.user_id]
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
