package grafana_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/managedgrafana"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgrafana "github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccGrafanaRoleAssociation_usersAdmin(t *testing.T) {
	role := managedgrafana.RoleAdmin
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"
	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(t)
				acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t)
				acctest.PreCheckSSOAdminInstances(t)
			},
			ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
			CheckDestroy: testAccCheckRoleAssociationDestroy,
			Providers:    acctest.Providers,
			Steps: []resource.TestStep{
				{
					Config: testAccWorkspaceGrafanaRoleAssociationUsers(rName, role),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "role", role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "user_ids.0", "USER_ID"),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					),
				},
			},
		})
}

func testAccGrafanaRoleAssociation_usersEditor(t *testing.T) {
	role := managedgrafana.RoleEditor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"
	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(t)
				acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t)
				acctest.PreCheckSSOAdminInstances(t)
			},
			ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
			CheckDestroy: testAccCheckRoleAssociationDestroy,
			Providers:    acctest.Providers,
			Steps: []resource.TestStep{
				{
					Config: testAccWorkspaceGrafanaRoleAssociationUsers(rName, role),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "role", role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "user_ids.0", "USER_ID"),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					),
				},
			},
		})
}

func testAccGrafanaRoleAssociation_groupsAdmin(t *testing.T) {
	role := managedgrafana.RoleAdmin
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"
	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(t)
				acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t)
				acctest.PreCheckSSOAdminInstances(t)
			},
			ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
			CheckDestroy: testAccCheckRoleAssociationDestroy,
			Providers:    acctest.Providers,
			Steps: []resource.TestStep{
				{
					Config: testAccWorkspaceGrafanaRoleAssociationGroups(rName, role),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "role", role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "user_ids.0", "USER_ID"),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					),
				},
			},
		})
}

func testAccGrafanaRoleAssociation_groupsEditor(t *testing.T) {
	role := managedgrafana.RoleEditor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"
	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(t)
				acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t)
				acctest.PreCheckSSOAdminInstances(t)
			},
			ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
			CheckDestroy: testAccCheckRoleAssociationDestroy,
			Providers:    acctest.Providers,
			Steps: []resource.TestStep{
				{
					Config: testAccWorkspaceGrafanaRoleAssociationGroups(rName, role),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "role", role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "user_ids.0", "USER_ID"),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					),
				},
			},
		})
}

func testAccGrafanaRoleAssociation_usersAndGroupsAdmin(t *testing.T) {
	role := managedgrafana.RoleAdmin
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"
	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(t)
				acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t)
				acctest.PreCheckSSOAdminInstances(t)
			},
			ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
			CheckDestroy: testAccCheckRoleAssociationDestroy,
			Providers:    acctest.Providers,
			Steps: []resource.TestStep{
				{
					Config: testAccWorkspaceGrafanaRoleAssociationUsersAndGroups(rName, role),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "role", role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "user_ids.0", "USER_ID"),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					),
				},
			},
		})
}

func testAccGrafanaRoleAssociation_usersAndGroupsEditor(t *testing.T) {
	role := managedgrafana.RoleEditor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_role_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"
	resource.Test(t,
		resource.TestCase{
			PreCheck: func() {
				acctest.PreCheck(t)
				acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t)
				acctest.PreCheckSSOAdminInstances(t)
			},
			ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
			CheckDestroy: testAccCheckRoleAssociationDestroy,
			Providers:    acctest.Providers,
			Steps: []resource.TestStep{
				{
					Config: testAccWorkspaceGrafanaRoleAssociationUsersAndGroups(rName, role),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckRoleAssociationExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "role", role),
						resource.TestCheckResourceAttr(resourceName, "user_ids.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "user_ids.0", "USER_ID"),
						resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					),
				},
			},
		})
}

func testAccWorkspaceGrafanaRoleAssociationUsers(rName string, role string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfigAuthenticationProvider(rName, "AWS_SSO"), fmt.Sprintf(`
resource "aws_grafana_role_association" "test" {
  role         = %[1]q
  user_ids     = ["9a67126fc1-ad12499e-78c0-4476-b180-f68a772b8be9"]
  workspace_id = aws_grafana_workspace.test.id
}
`, role))
}

func testAccWorkspaceGrafanaRoleAssociationGroups(rName string, role string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfigAuthenticationProvider(rName, "AWS_SSO"), fmt.Sprintf(`
resource "aws_grafana_role_association" "test" {
  role         = %[1]q
  group_ids    = ["9a67126fc1-6dbc70ac-546e-4191-af47-b0dd9fe827ce"]
  workspace_id = aws_grafana_workspace.test.id
}
`, role))
}

func testAccWorkspaceGrafanaRoleAssociationUsersAndGroups(rName string, role string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfigAuthenticationProvider(rName, "AWS_SSO"), fmt.Sprintf(`
resource "aws_grafana_role_association" "test" {
  role         = %[1]q
  user_ids     = ["9a67126fc1-ad12499e-78c0-4476-b180-f68a772b8be9"]
  group_ids    = ["9a67126fc1-6dbc70ac-546e-4191-af47-b0dd9fe827ce"]
  workspace_id = aws_grafana_workspace.test.id
}
`, role))
}

func testAccCheckRoleAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn

		usersMap, err := tfgrafana.FindRoleAssociationByRoleAndWorkspaceID(conn, rs.Primary.Attributes["role"], rs.Primary.Attributes["workspace_id"])

		if err != nil {
			return err
		}

		if len(usersMap[managedgrafana.UserTypeSsoUser]) == 0 && len(usersMap[managedgrafana.UserTypeSsoGroup]) == 0 {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccCheckRoleAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_grafana_role_association" {
			continue
		}

		role := rs.Primary.Attributes["role"]
		workspaceID := rs.Primary.Attributes["workspace_id"]
		userMap, err := tfgrafana.FindRoleAssociationByRoleAndWorkspaceID(conn, role, workspaceID)

		if tfresource.NotFound(err) {
			continue
		}

		if len(userMap[managedgrafana.UserTypeSsoUser]) > 0 || len(userMap[managedgrafana.UserTypeSsoGroup]) > 0 {
			return fmt.Errorf("Grafana Workspace Role Association %s-%s still exists", role, workspaceID)
		}

		if err != nil {
			return err
		}
	}
	return nil
}
