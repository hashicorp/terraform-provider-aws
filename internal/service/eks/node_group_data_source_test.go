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

func TestAccEKSNodeGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_eks_node_group.test"
	resourceName := "aws_eks_node_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeGroupConfig_dataSourceName(rName),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccNodeGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "ami_type", dataSourceResourceName, "ami_type"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "capacity_type", dataSourceResourceName, "capacity_type"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_name", dataSourceResourceName, "cluster_name"),
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
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceResourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.#", dataSourceResourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "taint.#", dataSourceResourceName, "taints.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceResourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "version", dataSourceResourceName, "version"),
				),
			},
		},
	})
}

func testAccNodeGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNodeGroupConfig_dataSourceName(rName), fmt.Sprintf(`
data "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
}
`, rName))
}
