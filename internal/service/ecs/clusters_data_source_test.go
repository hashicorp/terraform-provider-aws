// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSClustersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_ecs_clusters.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceResourceName, "cluster_arns.#", 1),
				),
			},
		},
	})
}

func testAccClustersDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), `
data "aws_ecs_clusters" "test" {
  depends_on = [aws_ecs_cluster.test]
}
`)
}
