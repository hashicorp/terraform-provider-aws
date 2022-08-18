package grafana_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/managedgrafana"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccWorkspaceApiKey_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_api_key.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:             nil,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceApiKey_providerBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "key_name", rName),
					resource.TestCheckResourceAttr(resourceName, "key_role", "EDITOR"),
					resource.TestCheckResourceAttr(resourceName, "seconds_to_live", "3600"),
				),
			},
		},
	})
}

func testAccWorkspaceApiKey_providerBasic(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceSAMLConfigurationConfig_providerBasic(rName), fmt.Sprintf(`
resource "aws_grafana_workspace_api_key" "test" {
  workspace_id    = aws_grafana_workspace.test.id
  key_name        = %[1]q
  key_role        = "EDITOR"
  seconds_to_live = 3600
}`, rName))
}
