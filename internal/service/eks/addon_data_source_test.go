// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEKSAddonDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_eks_addon.test"
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonDataSourceConfig_basic(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "addon_version", dataSourceResourceName, "addon_version"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_values", dataSourceResourceName, "configuration_values"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceResourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "modified_at", dataSourceResourceName, "modified_at"),
					resource.TestCheckResourceAttrPair(resourceName, "service_account_role_arn", dataSourceResourceName, "service_account_role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceResourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccEKSAddonDataSource_configurationValues(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_eks_addon.test"
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	addonVersion := "v1.15.3-eksbuild.1"
	configurationValues := "{\"env\": {\"WARM_ENI_TARGET\":\"2\",\"ENABLE_POD_ENI\":\"true\"},\"resources\": {\"limits\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"}}}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAddonDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAddonDataSourceConfig_configurationValues(rName, addonName, addonVersion, configurationValues, eks.ResolveConflictsOverwrite),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "addon_version", dataSourceResourceName, "addon_version"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_values", dataSourceResourceName, "configuration_values"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceResourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "modified_at", dataSourceResourceName, "modified_at"),
					resource.TestCheckResourceAttrPair(resourceName, "service_account_role_arn", dataSourceResourceName, "service_account_role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceResourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccAddonDataSourceConfig_basic(rName, addonName string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  addon_name   = %[2]q
  cluster_name = aws_eks_cluster.test.name
}

data "aws_eks_addon" "test" {
  addon_name   = %[2]q
  cluster_name = aws_eks_cluster.test.name

  depends_on = [
    aws_eks_addon.test,
    aws_eks_cluster.test,
  ]
}
`, rName, addonName))
}

func testAccAddonDataSourceConfig_configurationValues(rName, addonName, addonVersion, configurationValues, resolveConflicts string) string {
	return acctest.ConfigCompose(testAccAddonConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name         = aws_eks_cluster.test.name
  addon_name           = %[2]q
  addon_version        = %[3]q
  configuration_values = %[4]q
  resolve_conflicts    = %[5]q
}

data "aws_eks_addon" "test" {
  addon_name   = %[2]q
  cluster_name = aws_eks_cluster.test.name

  depends_on = [
    aws_eks_addon.test,
    aws_eks_cluster.test,
  ]
}
`, rName, addonName, addonVersion, configurationValues, resolveConflicts))
}
