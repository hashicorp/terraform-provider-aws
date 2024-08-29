// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgrafana "github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRoleAssociation_usersAdmin(t *testing.T) {
	ctx := acctest.Context(t)
	key := "GRAFANA_SSO_USER_ID"
	userID := os.Getenv(key)
	if userID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	role := string(awstypes.RoleAdmin)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(ctx, t)
				acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
				acctest.PreCheckSSOAdminInstances(ctx, t)
			},
			ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
			CheckDestroy:             testAccCheckRoleAssociationDestroy(ctx),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccRoleAssociationConfig_workspaceUsers(rName, role, userID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(ctx, resourceName),
						resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", acctest.Ct1),
						resource.TestCheckTypeSetElemAttr(resourceName, "user_ids.*", userID),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
					),
				},
			},
		})
}

func testAccRoleAssociation_usersEditor(t *testing.T) {
	ctx := acctest.Context(t)
	key := "GRAFANA_SSO_USER_ID"
	userID := os.Getenv(key)
	if userID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	role := string(awstypes.RoleEditor)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(ctx, t)
				acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
				acctest.PreCheckSSOAdminInstances(ctx, t)
			},
			ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
			CheckDestroy:             testAccCheckRoleAssociationDestroy(ctx),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccRoleAssociationConfig_workspaceUsers(rName, role, userID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(ctx, resourceName),
						resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", acctest.Ct1),
						resource.TestCheckTypeSetElemAttr(resourceName, "user_ids.*", userID),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
					),
				},
			},
		})
}

func testAccRoleAssociation_groupsAdmin(t *testing.T) {
	ctx := acctest.Context(t)
	key := "GRAFANA_SSO_GROUP_ID"
	groupID := os.Getenv(key)
	if groupID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	role := string(awstypes.RoleAdmin)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(ctx, t)
				acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
				acctest.PreCheckSSOAdminInstances(ctx, t)
			},
			ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
			CheckDestroy:             testAccCheckRoleAssociationDestroy(ctx),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccRoleAssociationConfig_workspaceGroups(rName, role, groupID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(ctx, resourceName),
						resource.TestCheckResourceAttr(resourceName, "group_ids.#", acctest.Ct1),
						resource.TestCheckTypeSetElemAttr(resourceName, "group_ids.*", groupID),
						resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
					),
				},
			},
		})
}

func testAccRoleAssociation_groupsEditor(t *testing.T) {
	ctx := acctest.Context(t)
	key := "GRAFANA_SSO_GROUP_ID"
	groupID := os.Getenv(key)
	if groupID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	role := string(awstypes.RoleEditor)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(ctx, t)
				acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
				acctest.PreCheckSSOAdminInstances(ctx, t)
			},
			ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
			CheckDestroy:             testAccCheckRoleAssociationDestroy(ctx),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccRoleAssociationConfig_workspaceGroups(rName, role, groupID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(ctx, resourceName),
						resource.TestCheckResourceAttr(resourceName, "group_ids.#", acctest.Ct1),
						resource.TestCheckTypeSetElemAttr(resourceName, "group_ids.*", groupID),
						resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
					),
				},
			},
		})
}

func testAccRoleAssociation_usersAndGroupsAdmin(t *testing.T) {
	ctx := acctest.Context(t)
	key := "GRAFANA_SSO_USER_ID"
	userID := os.Getenv(key)
	if userID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	key = "GRAFANA_SSO_GROUP_ID"
	groupID := os.Getenv(key)
	if groupID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	role := string(awstypes.RoleAdmin)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(ctx, t)
				acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
				acctest.PreCheckSSOAdminInstances(ctx, t)
			},
			ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
			CheckDestroy:             testAccCheckRoleAssociationDestroy(ctx),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccRoleAssociationConfig_workspaceUsersAndGroups(rName, role, userID, groupID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(ctx, resourceName),
						resource.TestCheckResourceAttr(resourceName, "group_ids.#", acctest.Ct1),
						resource.TestCheckTypeSetElemAttr(resourceName, "group_ids.*", groupID),
						resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", acctest.Ct1),
						resource.TestCheckTypeSetElemAttr(resourceName, "user_ids.*", userID),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
					),
				},
			},
		})
}

func testAccRoleAssociation_usersAndGroupsEditor(t *testing.T) {
	ctx := acctest.Context(t)
	key := "GRAFANA_SSO_USER_ID"
	userID := os.Getenv(key)
	if userID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	key = "GRAFANA_SSO_GROUP_ID"
	groupID := os.Getenv(key)
	if groupID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	role := string(awstypes.RoleEditor)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(ctx, t)
				acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
				acctest.PreCheckSSOAdminInstances(ctx, t)
			},
			ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
			CheckDestroy:             testAccCheckRoleAssociationDestroy(ctx),
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccRoleAssociationConfig_workspaceUsersAndGroups(rName, role, userID, groupID),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(ctx, resourceName),
						resource.TestCheckResourceAttr(resourceName, "group_ids.#", acctest.Ct1),
						resource.TestCheckTypeSetElemAttr(resourceName, "group_ids.*", groupID),
						resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", acctest.Ct1),
						resource.TestCheckTypeSetElemAttr(resourceName, "user_ids.*", userID),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
					),
				},
			},
		})
}

func testAccRoleAssociationConfig_workspaceUsers(rName, role, userID string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_authenticationProvider(rName, "AWS_SSO"), fmt.Sprintf(`
resource "aws_grafana_role_association" "test" {
  role         = %[1]q
  user_ids     = [%[2]q]
  workspace_id = aws_grafana_workspace.test.id
}
`, role, userID))
}

func testAccRoleAssociationConfig_workspaceGroups(rName, role, groupID string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_authenticationProvider(rName, "AWS_SSO"), fmt.Sprintf(`
resource "aws_grafana_role_association" "test" {
  role         = %[1]q
  group_ids    = [%[2]q]
  workspace_id = aws_grafana_workspace.test.id
}
`, role, groupID))
}

func testAccRoleAssociationConfig_workspaceUsersAndGroups(rName, role, userID, groupID string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_authenticationProvider(rName, "AWS_SSO"), fmt.Sprintf(`
resource "aws_grafana_role_association" "test" {
  role         = %[1]q
  user_ids     = [%[2]q]
  group_ids    = [%[3]q]
  workspace_id = aws_grafana_workspace.test.id
}
`, role, userID, groupID))
}

func testAccCheckRoleAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaClient(ctx)

		_, err := tfgrafana.FindRoleAssociationsByTwoPartKey(ctx, conn, awstypes.Role(rs.Primary.Attributes[names.AttrRole]), rs.Primary.Attributes["workspace_id"])

		return err
	}
}

func testAccCheckRoleAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_grafana_role_association" {
				continue
			}

			_, err := tfgrafana.FindRoleAssociationsByTwoPartKey(ctx, conn, awstypes.Role(rs.Primary.Attributes[names.AttrRole]), rs.Primary.Attributes["workspace_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Grafana Workspace Role Association %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
