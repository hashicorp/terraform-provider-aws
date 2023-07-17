// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccDataSourceCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudhsm_v2_cluster.test"
	dataSourceName := "data.aws_cloudhsm_v2_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cluster_state", "UNINITIALIZED"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_id", resourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_state", resourceName, "cluster_state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_id", resourceName, "security_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccClusterDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

data "aws_cloudhsm_v2_cluster" "test" {
  cluster_id = aws_cloudhsm_v2_cluster.test.cluster_id
}
`, rName))
}
