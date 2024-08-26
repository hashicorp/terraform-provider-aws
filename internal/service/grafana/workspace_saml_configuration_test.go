// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgrafana "github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccWorkspaceSAMLConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_saml_configuration.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSAMLConfigurationConfig_providerBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceSAMLConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.0", "admin"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.0", "editor"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_metadata_xml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.SamlConfigurationStatusConfigured)),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccWorkspaceSAMLConfiguration_loginValidity(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_saml_configuration.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSAMLConfigurationConfig_providerLoginValidity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceSAMLConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.0", "admin"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.0", "editor"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_metadata_xml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.SamlConfigurationStatusConfigured)),
					resource.TestCheckResourceAttr(resourceName, "login_validity_duration", "1440"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccWorkspaceSAMLConfiguration_assertions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_saml_configuration.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSAMLConfigurationConfig_providerAssertions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceSAMLConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "admin_role_values.0", "admin"),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "editor_role_values.0", "editor"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_metadata_xml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.SamlConfigurationStatusConfigured)),
					resource.TestCheckResourceAttr(resourceName, "email_assertion", "mail"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_metadata_xml"),
					resource.TestCheckResourceAttr(resourceName, "groups_assertion", "groups"),
					resource.TestCheckResourceAttr(resourceName, "login_assertion", "mail"),
					resource.TestCheckResourceAttr(resourceName, "name_assertion", "displayName"),
					resource.TestCheckResourceAttr(resourceName, "org_assertion", "org"),
					resource.TestCheckResourceAttr(resourceName, "role_assertion", names.AttrRole),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
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

func testAccCheckWorkspaceSAMLConfigurationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaClient(ctx)

		_, err := tfgrafana.FindSAMLConfigurationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}
