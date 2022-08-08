package amp_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/prometheusservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAMPWorkspaceDataSource_basic(t *testing.T) {
	randName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amp_workspace.test"
	dataSourceName := "data.aws_amp_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(prometheusservice.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		CheckDestroy:             nil,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig_alias(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(dataSourceName),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "prometheus_endpoint", dataSourceName, "prometheus_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "alias", dataSourceName, "alias"),
				),
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig_alias(randName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias = %q
}
`, randName)
}
