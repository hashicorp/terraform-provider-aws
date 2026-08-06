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

func TestAccKafkaServerlessClusterDataSource_basic(t *testing.T) {
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
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_sasl_iam", resourceName, "bootstrap_brokers_sasl_iam"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_uuid", resourceName, "cluster_uuid"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccServleressClusterDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccServerlessClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_msk_serverless_cluster" "test" {
  cluster_name = %[1]q

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
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
