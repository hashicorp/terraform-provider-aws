// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"fmt"
	"testing"
	"github.com/aws/aws-sdk-go/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKafkaBootstrapBrokersDataSourcebasic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_msk_bootstrap_brokers.test"
	resourceName := "aws_msk_serverless_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kafka.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBootstrapBrokersDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_string_sasl_iam", resourceName, "bootstrap_brokers_string_sasl_iam"),

				),
			},
		},
	})
}

func testAccBootstrapBrokersataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_msk_serverless_cluster" "test" {
  cluster_name = %[1]q

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
    security_groups = [aws_security_group.test.id]
  }

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  tags = {
    foo = "bar"
  }
}

data "aws_msk_bootstrap_brokers" "test" {
  bootstrap_broker_string_sasl_iam = aws_msk_serverless_cluster.test.bootstrap_broker_string_sasl_iam
}
`, rName))
}
