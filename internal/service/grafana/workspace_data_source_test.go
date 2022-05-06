package grafana_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/managedgrafana"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccWorkspaceDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"
	dataSourceName := "data.aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(dataSourceName),
					resource.TestCheckResourceAttrPair(resourceName, "account_access_type", dataSourceName, "account_access_type"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "authentication_providers.#", dataSourceName, "authentication_providers.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "data_sources.#", dataSourceName, "data_sources.#"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", dataSourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "grafana_version", dataSourceName, "grafana_version"),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "notification_destinations.#", dataSourceName, "notification_destinations.#"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_role_name", dataSourceName, "organization_role_name"),
					resource.TestCheckResourceAttrPair(resourceName, "organizational_units.#", dataSourceName, "organizational_units.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_type", dataSourceName, "permission_type"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", dataSourceName, "role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "saml_configuration_status", dataSourceName, "saml_configuration_status"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", dataSourceName, "stack_set_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfigAuthenticationProvider(rName, "SAML"), `
data "aws_grafana_workspace" "test" {
  workspace_id = aws_grafana_workspace.test.id
}
`)
}
