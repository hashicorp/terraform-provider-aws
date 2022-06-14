package eks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEKSAddonVersionDataSource_basic(t *testing.T) {
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	versionDataSourceName := "data.aws_eks_addon_version.test"
	addonDataSourceName := "data.aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t); testAccPreCheckAddon(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAddonVersionDataSourceConfig_basic(rName, addonName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, addonDataSourceName, &addon),
					resource.TestCheckResourceAttrPair(versionDataSourceName, "version", addonDataSourceName, "addon_version"),
					resource.TestCheckResourceAttrPair(versionDataSourceName, "addon_name", addonDataSourceName, "addon_name"),
					resource.TestCheckResourceAttr(versionDataSourceName, "most_recent", "true"),
				),
			},
			{
				Config: testAccAddonVersionDataSourceConfig_basic(rName, addonName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddonExists(ctx, addonDataSourceName, &addon),
					resource.TestCheckResourceAttrPair(versionDataSourceName, "version", addonDataSourceName, "addon_version"),
					resource.TestCheckResourceAttrPair(versionDataSourceName, "addon_name", addonDataSourceName, "addon_name"),
					resource.TestCheckResourceAttr(versionDataSourceName, "most_recent", "false"),
				),
			},
		},
	})
}

func testAccAddonVersionDataSourceConfig_basic(rName, addonName string, mostRecent bool) string {
	return acctest.ConfigCompose(testAccAddonBaseConfig(rName), fmt.Sprintf(`
data "aws_eks_addon_version" "test" {
  addon_name         = %[2]q
  kubernetes_version = aws_eks_cluster.test.version
  most_recent        = %[3]t
}

resource "aws_eks_addon" "test" {
  addon_name    = %[2]q
  cluster_name  = aws_eks_cluster.test.name
  addon_version = data.aws_eks_addon_version.test.version

  resolve_conflicts = "OVERWRITE"
}

data "aws_eks_addon" "test" {
  addon_name   = %[2]q
  cluster_name = aws_eks_cluster.test.name

  depends_on = [
    data.aws_eks_addon_version.test,
    aws_eks_addon.test,
    aws_eks_cluster.test,
  ]
}
`, rName, addonName, mostRecent))
}
