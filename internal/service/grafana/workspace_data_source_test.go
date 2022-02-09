package grafana_test

import (
	"fmt"
	"testing"
)

import (
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccGrafanaWorkspaceDataSource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"
	dataSourceName := "data.aws_grafana_workspace.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(dataSourceName),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "account_access_type", dataSourceName, "account_access_type"),
					resource.TestCheckResourceAttrPair(resourceName, "authentication_providers.0", dataSourceName, "authentication_providers.0"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_type", dataSourceName, "permission_type"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", dataSourceName, "role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", dataSourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "grafana_version", dataSourceName, "grafana_version"),
				),
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig(name string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfigSaml(name),
		fmt.Sprint(`
data "aws_grafana_workspace" "test" {
  id = aws_grafana_workspace.test.id
}
`))
}
