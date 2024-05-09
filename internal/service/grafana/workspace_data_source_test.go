// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/managedgrafana"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccWorkspaceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"
	dataSourceName := "data.aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, managedgrafana.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
		CheckDestroy:             nil,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "account_access_type", dataSourceName, "account_access_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "authentication_providers.#", dataSourceName, "authentication_providers.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "data_sources.#", dataSourceName, "data_sources.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", dataSourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "grafana_version", dataSourceName, "grafana_version"),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "notification_destinations.#", dataSourceName, "notification_destinations.#"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_role_name", dataSourceName, "organization_role_name"),
					resource.TestCheckResourceAttrPair(resourceName, "organizational_units.#", dataSourceName, "organizational_units.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_type", dataSourceName, "permission_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, dataSourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrPair(resourceName, "saml_configuration_status", dataSourceName, "saml_configuration_status"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", dataSourceName, "stack_set_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_authenticationProvider(rName, "SAML"), `
data "aws_grafana_workspace" "test" {
  workspace_id = aws_grafana_workspace.test.id
}
`)
}
