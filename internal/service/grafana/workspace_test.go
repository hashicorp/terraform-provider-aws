package grafana_test

import (
	"fmt"
	"regexp"
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

func TestAccGrafana_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Workspace": {
			"saml":                     testAccWorkspace_saml,
			"sso":                      testAccWorkspace_sso,
			"disappears":               testAccWorkspace_disappears,
			"organization":             testAccWorkspace_organization,
			"dataSources":              testAccWorkspace_dataSources,
			"permissionType":           testAccWorkspace_permissionType,
			"notificationDestinations": testAccWorkspace_notificationDestinations,
			"tags":                     testAccWorkspace_tags,
		},
		"DataSource": {
			"basic": testAccWorkspaceDataSource_basic,
		},
		"LicenseAssociation": {
			"enterpriseFreeTrial": testAccLicenseAssociation_freeTrial,
		},
		"SamlConfiguration": {
			"basic":         testAccWorkspaceSamlConfiguration_basic,
			"loginValidity": testAccWorkspaceSamlConfiguration_loginValidity,
			"assertions":    testAccWorkspaceSamlConfiguration_assertions,
		},
		"RoleAssociation": {
			"usersAdmin":           testAccRoleAssociation_usersAdmin,
			"usersEditor":          testAccRoleAssociation_usersEditor,
			"groupsAdmin":          testAccRoleAssociation_groupsAdmin,
			"groupsEditor":         testAccRoleAssociation_groupsEditor,
			"usersAndGroupsAdmin":  testAccRoleAssociation_usersAndGroupsAdmin,
			"usersAndGroupsEditor": testAccRoleAssociation_usersAndGroupsEditor,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccWorkspace_saml(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      testAccCheckWorkspaceDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigAuthenticationProvider(rName, "SAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "grafana", regexp.MustCompile(`/workspaces/.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_access_type", managedgrafana.AccountAccessTypeCurrentAccount),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.0", managedgrafana.AuthenticationProviderTypesSaml),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "grafana_version"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_role_name", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_units.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeServiceManaged),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "saml_configuration_status", managedgrafana.SamlConfigurationStatusNotConfigured),
					resource.TestCheckResourceAttr(resourceName, "stack_set_name", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccWorkspace_sso(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t)
			acctest.PreCheckSSOAdminInstances(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      testAccCheckWorkspaceDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigAuthenticationProvider(rName, "AWS_SSO"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "grafana", regexp.MustCompile(`/workspaces/.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_access_type", managedgrafana.AccountAccessTypeCurrentAccount),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.0", managedgrafana.AuthenticationProviderTypesAwsSso),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "grafana_version"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_role_name", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_units.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeServiceManaged),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "saml_configuration_status", ""),
					resource.TestCheckResourceAttr(resourceName, "stack_set_name", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccWorkspace_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      testAccCheckWorkspaceDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigAuthenticationProvider(rName, "SAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfgrafana.ResourceWorkspace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccWorkspace_organization(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t)
			acctest.PreCheckOrganizationsEnabled(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      testAccCheckWorkspaceDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigOrganization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_access_type", managedgrafana.AccountAccessTypeOrganization),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.0", managedgrafana.AuthenticationProviderTypesSaml),
					resource.TestCheckResourceAttr(resourceName, "organization_role_name", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_units.#", "1"),
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

func testAccWorkspace_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      testAccCheckWorkspaceDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
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
				Config: testAccWorkspaceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkspaceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccWorkspace_dataSources(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      testAccCheckWorkspaceDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigDataSources(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "grafana", regexp.MustCompile(`/workspaces/.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_access_type", managedgrafana.AccountAccessTypeCurrentAccount),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_providers.0", managedgrafana.AuthenticationProviderTypesSaml),
					resource.TestCheckResourceAttr(resourceName, "data_sources.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "data_sources.0", "CLOUDWATCH"),
					resource.TestCheckResourceAttr(resourceName, "data_sources.1", "PROMETHEUS"),
					resource.TestCheckResourceAttr(resourceName, "data_sources.2", "XRAY"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "grafana_version"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_role_name", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_units.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeServiceManaged),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "saml_configuration_status", managedgrafana.SamlConfigurationStatusNotConfigured),
					resource.TestCheckResourceAttr(resourceName, "stack_set_name", ""),
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

func testAccWorkspace_permissionType(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      testAccCheckWorkspaceDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigPermissionType(rName, "CUSTOMER_MANAGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeCustomerManaged),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfigPermissionType(rName, "SERVICE_MANAGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeServiceManaged),
				),
			},
		},
	})
}

func testAccWorkspace_notificationDestinations(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      testAccCheckWorkspaceDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigAuthenticationProvider(rName, "SAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "0"),
				),
			},
			{
				Config: testAccWorkspaceConfigNotificationDestinations(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_destinations.0", "SNS"),
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

func testAccWorkspaceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "grafana.amazonaws.com"
        }
      },
    ]
  })
}
`, rName)
}

func testAccWorkspaceConfigAuthenticationProvider(rName, authenticationProvider string) string {
	return acctest.ConfigCompose(testAccWorkspaceRole(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = [%[1]q]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.test.arn
}
`, authenticationProvider))
}

func testAccWorkspaceConfigOrganization(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceRole(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "ORGANIZATION"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.test.arn
  organizational_units     = [aws_organizations_organizational_unit.test.id]
}

data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.test.roots[0].id
}
`, rName))
}

func testAccWorkspaceConfigTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccWorkspaceRole(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = %[1]q
  role_arn                 = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }

}
  `, rName, tagKey1, tagValue1))
}

func testAccWorkspaceConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccWorkspaceRole(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = %[1]q
  role_arn                 = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

}
  `, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccWorkspaceConfigDataSources(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceRole(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = %[1]q
  description              = %[1]q
  data_sources             = ["CLOUDWATCH", "PROMETHEUS", "XRAY"]
  role_arn                 = aws_iam_role.test.arn
}
`, rName))
}

func testAccWorkspaceConfigPermissionType(rName, permissionType string) string {
	return acctest.ConfigCompose(testAccWorkspaceRole(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = %[1]q
  role_arn                 = aws_iam_role.test.arn
}
`, permissionType))
}

func testAccWorkspaceConfigNotificationDestinations(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceRole(rName), fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
  account_access_type       = "CURRENT_ACCOUNT"
  authentication_providers  = ["SAML"]
  permission_type           = "SERVICE_MANAGED"
  name                      = %[1]q
  description               = %[1]q
  notification_destinations = ["SNS"]
  role_arn                  = aws_iam_role.test.arn
}
`, rName))
}

func testAccCheckWorkspaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Grafana Workspace ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn

		_, err := tfgrafana.FindWorkspaceByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckWorkspaceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_grafana_workspace" {
			continue
		}

		_, err := tfgrafana.FindWorkspaceByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Grafana Workspace %s still exists", rs.Primary.ID)
	}
	return nil
}
