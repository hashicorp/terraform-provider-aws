// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_msk_cluster.test"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers", resourceName, "bootstrap_brokers"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_public_sasl_iam", resourceName, "bootstrap_brokers_public_sasl_iam"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_public_sasl_scram", resourceName, "bootstrap_brokers_public_sasl_scram"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_public_tls", resourceName, "bootstrap_brokers_public_tls"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_sasl_iam", resourceName, "bootstrap_brokers_sasl_iam"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_sasl_scram", resourceName, "bootstrap_brokers_sasl_scram"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_tls", resourceName, "bootstrap_brokers_tls"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.az_distribution", string(types.BrokerAZDistributionDefault)),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.client_subnets.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.test.1", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.test.2", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.0.type", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.vpc_connectivity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.vpc_connectivity.0.client_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.vpc_connectivity.0.client_authentication.0.sasl.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.vpc_connectivity.0.client_authentication.0.sasl.0.iam", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.vpc_connectivity.0.client_authentication.0.sasl.0.scram", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.vpc_connectivity.0.client_authentication.0.tls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.instance_type", "kafka.t3.small"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.security_groups.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.security_groups.*", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.volume_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrClusterName, resourceName, names.AttrClusterName),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_uuid", resourceName, "cluster_uuid"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kafka_version", resourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_of_broker_nodes", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string", resourceName, "zookeeper_connect_string"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string_tls", resourceName, "zookeeper_connect_string_tls"),
				),
			},
		},
	})
}

func testAccClusterDataSourceConfig_basic(rName string) string {
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
    Name = %[1]q
  }
}

data "aws_msk_cluster" "test" {
  cluster_name = aws_msk_cluster.test.cluster_name
}
`, rName))
}
