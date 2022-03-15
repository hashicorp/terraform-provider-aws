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

func testAccGrafanaWorkspaceSamlConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_saml_configuration.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy: testAccCheckWorkspaceSamlConfigurationDestroy,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSamlConfigurationConfigProvider_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceSamlConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.0", "admin"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.0", "editor"),
					resource.TestCheckResourceAttr(resourceName, "status", managedgrafana.SamlConfigurationStatusConfigured),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
				),
			},
		},
	})
}

func testAccGrafanaWorkspaceSamlConfiguration_loginValidity(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_saml_configuration.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy: testAccCheckWorkspaceSamlConfigurationDestroy,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSamlConfigurationConfigProvider_loginValidity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceSamlConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.0", "admin"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.0", "editor"),
					resource.TestCheckResourceAttr(resourceName, "status", managedgrafana.SamlConfigurationStatusConfigured),
					resource.TestCheckResourceAttr(resourceName, "login_validity_duration", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
				),
			},
		},
	})
}

func testAccGrafanaWorkspaceSamlConfiguration_assertions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_saml_configuration.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy: testAccCheckWorkspaceSamlConfigurationDestroy,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSamlConfigurationConfigProvider_assertions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceSamlConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.0", "admin"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.0", "editor"),
					resource.TestCheckResourceAttr(resourceName, "status", managedgrafana.SamlConfigurationStatusConfigured),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
				),
			},
		},
	})
}

func testAccWorkspaceSamlConfigurationConfigProvider_basic(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfigAuthenticationProvider(rName, "SAML"), fmt.Sprintf(`
resource "aws_grafana_workspace_saml_configuration" "test" {
  admin_role_values  = ["admin"]
  editor_role_values = ["editor"]
  idp_metadata_xml   = file("test-fixtures/idp_metadata.xml")
  workspace_id       = aws_grafana_workspace.test.id
}
`))
}

func testAccWorkspaceSamlConfigurationConfigProvider_loginValidity(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfigAuthenticationProvider(rName, "SAML"), fmt.Sprintf(`
resource "aws_grafana_workspace_saml_configuration" "test" {
  admin_role_values       = ["admin"]
  editor_role_values      = ["editor"]
  idp_metadata_xml        = file("test-fixtures/idp_metadata.xml")
  login_validity_duration = 100
  workspace_id            = aws_grafana_workspace.test.id
}
`))
}

func testAccWorkspaceSamlConfigurationConfigProvider_assertions(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfigAuthenticationProvider(rName, "SAML"), fmt.Sprintf(`
resource "aws_grafana_workspace_saml_configuration" "test" {
  admin_role_values  = ["admin"]
  editor_role_values = ["editor"]
  email_assertion    = "email"
  groups_assertion   = "groups"
  login_assertion    = "login"
  name_assertion     = "name"
  org_assertion      = "org"
  role_assertion     = "role"
  idp_metadata_xml   = file("test-fixtures/idp_metadata.xml")
  workspace_id       = aws_grafana_workspace.test.id
}
`))
}

func testAccCheckWorkspaceSamlConfigurationExists(name string) resource.TestCheckFunc {
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

func testAccCheckWorkspaceSamlConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_grafana_workspace_saml_configuration" {
			continue
		}

		_, err := tfgrafana.FindSamlConfigurationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Grafana Workspace Saml Configuration %s still exists", rs.Primary.ID)
	}
	return nil
}
