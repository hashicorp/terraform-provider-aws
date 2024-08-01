// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminAccountAssignment_Basic_group(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_account_assignment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	groupName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
			testAccPreCheckIdentityStoreGroupName(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAssignmentConfig_basicGroup(groupName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountAssignmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "AWS_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "GROUP"),
					resource.TestMatchResourceAttr(resourceName, "principal_id", regexache.MustCompile("^([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}")),
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
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_account_assignment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
			testAccPreCheckIdentityStoreUserName(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAssignmentConfig_basicUser(userName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountAssignmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "AWS_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "USER"),
					resource.TestMatchResourceAttr(resourceName, "principal_id", regexache.MustCompile("^([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}")),
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

func TestAccSSOAdminAccountAssignment_MissingPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
			testAccPreCheckIdentityStoreUserName(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// We assign a policy called rName on the assumption it doesn't exist due to being randomly generated, hoping to generate an error
				Config:      testAccAccountAssignmentConfig_withCustomerPolicy(userName, "/", rName, rName),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`Received a 404 status error: Not supported policy.*%s`, rName)),
			},
		},
	})
}

func TestAccSSOAdminAccountAssignment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_account_assignment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	groupName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
			testAccPreCheckIdentityStoreGroupName(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAssignmentConfig_basicGroup(groupName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountAssignmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourceAccountAssignment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountAssignmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_account_assignment" {
				continue
			}

			idParts, err := tfssoadmin.ParseAccountAssignmentID(rs.Primary.ID)
			if err != nil {
				return err
			}

			principalID := idParts[0]
			principalType := idParts[1]
			targetID := idParts[2]
			permissionSetARN := idParts[4]
			instanceARN := idParts[5]

			_, err = tfssoadmin.FindAccountAssignment(ctx, conn, principalID, principalType, targetID, permissionSetARN, instanceARN)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSO Account Assignment for Principal (%s) still exists", principalID)
		}

		return nil
	}
}

func testAccCheckAccountAssignmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		idParts, err := tfssoadmin.ParseAccountAssignmentID(rs.Primary.ID)
		if err != nil {
			return err
		}

		principalID := idParts[0]
		principalType := idParts[1]
		targetID := idParts[2]
		permissionSetARN := idParts[4]
		instanceARN := idParts[5]

		_, err = tfssoadmin.FindAccountAssignment(ctx, conn, principalID, principalType, targetID, permissionSetARN, instanceARN)

		return err
	}
}

func testAccAccountAssignmentConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_caller_identity" "current" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccAccountAssignmentConfig_basicGroup(groupName, rName string) string {
	return acctest.ConfigCompose(testAccAccountAssignmentConfig_base(rName), fmt.Sprintf(`
data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "DisplayName"
      attribute_value = %[1]q
    }
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

func testAccAccountAssignmentConfig_basicUser(userName, rName string) string {
	return acctest.ConfigCompose(testAccAccountAssignmentConfig_base(rName), fmt.Sprintf(`
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "UserName"
      attribute_value = %[1]q
    }
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

func testAccAccountAssignmentConfig_withCustomerPolicy(userName, policyPath, policyName, rName string) string {
	return acctest.ConfigCompose(
		testAccAccountAssignmentConfig_basicUser(userName, rName),
		fmt.Sprintf(`
resource "aws_ssoadmin_customer_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn

  customer_managed_policy_reference {
    name = %[1]q
    path = %[2]q
  }
}
`, policyName, policyPath))
}
