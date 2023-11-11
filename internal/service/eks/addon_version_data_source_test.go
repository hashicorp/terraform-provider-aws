// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSAddonVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	versionDataSourceName := "data.aws_eks_addon_version.test"
	addonDataSourceName := "data.aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAddonVersionDataSourceConfig_basic(rName, addonName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(versionDataSourceName, "version", addonDataSourceName, "addon_version"),
					resource.TestCheckResourceAttrPair(versionDataSourceName, "addon_name", addonDataSourceName, "addon_name"),
					resource.TestCheckResourceAttr(versionDataSourceName, "most_recent", "true"),
				),
			},
			{
				Config: testAccAddonVersionDataSourceConfig_basic(rName, addonName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(versionDataSourceName, "version", addonDataSourceName, "addon_version"),
					resource.TestCheckResourceAttrPair(versionDataSourceName, "addon_name", addonDataSourceName, "addon_name"),
					resource.TestCheckResourceAttr(versionDataSourceName, "most_recent", "false"),
				),
			},
		},
	})
}

func testAccAddonVersionDataSourceConfig_basic(rName, addonName string, mostRecent bool) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
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
