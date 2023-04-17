package amp_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/prometheusservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAMPWorkspacesDataSource_basic(t *testing.T) { // nosemgrep:ci.caps0-in-func-name
	ctx := acctest.Context(t)
	rCount := strconv.Itoa(sdkacctest.RandIntRange(1, 4))
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_prometheus_workspaces.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, prometheusservice.EndpointsID)
		},
		ErrorCheck:                acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		PreventPostDestroyRefresh: true,
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfig_base(rCount, rName),
			},
			{
				Config: testAccWorkspacesDataSourceConfig_basic(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "aliases.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "workspace_ids.#", rCount),
				),
			},
		},
	})
}

func TestAccAMPWorkspacesDataSource_aliasPrefix(t *testing.T) { // nosemgrep:ci.caps0-in-func-name
	ctx := acctest.Context(t)
	rCount := strconv.Itoa(sdkacctest.RandIntRange(1, 4))
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_prometheus_workspaces.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, prometheusservice.EndpointsID)
		},
		ErrorCheck:                acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		PreventPostDestroyRefresh: true,
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfig_aliasPrefix(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "aliases.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "workspace_ids.#", "1"),
				),
			},
		},
	})
}

func testAccWorkspacesDataSourceConfig_base(rCount, rName string) string { // nosemgrep:ci.caps0-in-func-name
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  count = %[1]s
  alias = "%[2]s-${count.index}"
}
`, rCount, rName)
}

func testAccWorkspacesDataSourceConfig_basic(rCount, rName string) string { // nosemgrep:ci.caps0-in-func-name
	return acctest.ConfigCompose(testAccWorkspacesDataSourceConfig_base(rCount, rName), `
data "aws_prometheus_workspaces" "test" {}
`)
}

func testAccWorkspacesDataSourceConfig_aliasPrefix(rCount, rName string) string { // nosemgrep:ci.caps0-in-func-name
	return acctest.ConfigCompose(testAccWorkspacesDataSourceConfig_base(rCount, rName), `
data "aws_prometheus_workspaces" "test" {
  alias_prefix = aws_prometheus_workspace.test[0].alias
}
`)
}
