// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaBrokerNodesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_msk_broker_nodes.test"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerNodesDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_info_list.#", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttr(dataSourceName, "node_info_list.0.broker_id", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "node_info_list.1.broker_id", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, "node_info_list.2.broker_id", acctest.Ct3),
				),
			},
		},
	})
}

func testAccBrokerNodesDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.8.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.test.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }

  tags = {
    foo = "bar"
  }
}

data "aws_msk_broker_nodes" "test" {
  cluster_arn = aws_msk_cluster.test.arn
}
`, rName))
}
