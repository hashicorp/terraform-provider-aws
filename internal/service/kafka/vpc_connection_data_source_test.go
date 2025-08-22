// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaVPCConnectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_msk_vpc_connection.test"
	resourceName := "aws_msk_vpc_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "authentication", resourceName, "authentication"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_subnets.#", resourceName, "client_subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.% ", resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "target_cluster_arn", resourceName, "target_cluster_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func testAccVPCConnectionDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCConnectionConfig_basic(rName), `
data "aws_msk_vpc_connection" "test" {
  arn = aws_msk_vpc_connection.test.arn
}
`)
}
