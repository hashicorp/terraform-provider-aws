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
)

func testAccWorkspaceSAMLConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_saml_configuration.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSAMLConfigurationConfig_providerBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceSAMLConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.0", "admin"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.0", "editor"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_metadata_xml"),
					resource.TestCheckResourceAttr(resourceName, "status", managedgrafana.SamlConfigurationStatusConfigured),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
				),
			},
		},
	})
}

func testAccWorkspaceSAMLConfiguration_loginValidity(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_saml_configuration.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSAMLConfigurationConfig_providerLoginValidity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceSAMLConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.0", "admin"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.0", "editor"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_metadata_xml"),
					resource.TestCheckResourceAttr(resourceName, "status", managedgrafana.SamlConfigurationStatusConfigured),
					resource.TestCheckResourceAttr(resourceName, "login_validity_duration", "1440"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
				),
			},
		},
	})
}

func testAccWorkspaceSAMLConfiguration_assertions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_saml_configuration.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSAMLConfigurationConfig_providerAssertions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceSAMLConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.0", "admin"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.0", "editor"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_metadata_xml"),
					resource.TestCheckResourceAttr(resourceName, "status", managedgrafana.SamlConfigurationStatusConfigured),
					resource.TestCheckResourceAttr(resourceName, "email_assertion", "mail"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_metadata_xml"),
					resource.TestCheckResourceAttr(resourceName, "groups_assertion", "groups"),
					resource.TestCheckResourceAttr(resourceName, "login_assertion", "mail"),
					resource.TestCheckResourceAttr(resourceName, "name_assertion", "displayName"),
					resource.TestCheckResourceAttr(resourceName, "org_assertion", "org"),
					resource.TestCheckResourceAttr(resourceName, "role_assertion", "role"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
				),
			},
		},
	})
}

func testAccWorkspaceSAMLConfigurationConfig_providerBasic(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_authenticationProvider(rName, "SAML"), `
resource "aws_grafana_workspace_saml_configuration" "test" {
  admin_role_values  = ["admin"]
  editor_role_values = ["editor"]
  idp_metadata_xml   = file("test-fixtures/idp_metadata.xml")
  workspace_id       = aws_grafana_workspace.test.id
}
`)
}

func testAccWorkspaceSAMLConfigurationConfig_providerLoginValidity(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_authenticationProvider(rName, "SAML"), `
resource "aws_grafana_workspace_saml_configuration" "test" {
  admin_role_values       = ["admin"]
  editor_role_values      = ["editor"]
  idp_metadata_xml        = file("test-fixtures/idp_metadata.xml")
  login_validity_duration = 1440
  workspace_id            = aws_grafana_workspace.test.id
}
`)
}

func testAccWorkspaceSAMLConfigurationConfig_providerAssertions(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_authenticationProvider(rName, "SAML"), `
resource "aws_grafana_workspace_saml_configuration" "test" {
  admin_role_values  = ["admin"]
  editor_role_values = ["editor"]
  email_assertion    = "mail"
  groups_assertion   = "groups"
  login_assertion    = "mail"
  name_assertion     = "displayName"
  org_assertion      = "org"
  role_assertion     = "role"
  idp_metadata_xml   = file("test-fixtures/idp_metadata.xml")
  workspace_id       = aws_grafana_workspace.test.id
}
`)
}

func testAccCheckWorkspaceSAMLConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Grafana Workspace ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn

		_, err := tfgrafana.FindSamlConfigurationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}
