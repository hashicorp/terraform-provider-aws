// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSNodeGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_eks_node_group.test"
	resourceName := "aws_eks_node_group.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeGroupConfig_name(rName),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccNodeGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "ami_type", dataSourceResourceName, "ami_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "capacity_type", dataSourceResourceName, "capacity_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterName, dataSourceResourceName, names.AttrClusterName),
					resource.TestCheckResourceAttrPair(resourceName, "disk_size", dataSourceResourceName, "disk_size"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_types.#", dataSourceResourceName, "instance_types.#"),
					resource.TestCheckResourceAttrPair(resourceName, "labels.%", dataSourceResourceName, "labels.%"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.#", dataSourceResourceName, "launch_template.#"),
					resource.TestCheckResourceAttrPair(resourceName, "node_group_name", dataSourceResourceName, "node_group_name"),
					resource.TestCheckResourceAttrPair(resourceName, "node_role_arn", dataSourceResourceName, "node_role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "release_version", dataSourceResourceName, "release_version"),
					resource.TestCheckResourceAttrPair(resourceName, "remote_access.#", dataSourceResourceName, "remote_access.#"),
					resource.TestCheckResourceAttrPair(resourceName, "resources.#", dataSourceResourceName, "resources.#"),
					resource.TestCheckResourceAttrPair(resourceName, "scaling_config.#", dataSourceResourceName, "scaling_config.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceResourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.#", dataSourceResourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "taint.#", dataSourceResourceName, "taints.#"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceResourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, "update_config.#", dataSourceResourceName, "update_config.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVersion, dataSourceResourceName, names.AttrVersion),
				),
			},
		},
	})
}

func testAccNodeGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNodeGroupConfig_name(rName), fmt.Sprintf(`
data "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
}
`, rName))
}
