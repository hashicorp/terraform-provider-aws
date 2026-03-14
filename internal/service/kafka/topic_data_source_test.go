// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaTopicDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_msk_topic.test"
	resourceName := "aws_msk_topic.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_arn", resourceName, "cluster_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "topic_name", resourceName, "topic_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "partition_count", resourceName, "partition_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_factor", resourceName, "replication_factor"),
				),
			},
		},
	})
}

func testAccTopicDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTopicConfig_basic(rName), `
data "aws_msk_topic" "test" {
  cluster_arn = aws_msk_cluster.test.arn
  topic_name  = aws_msk_topic.test.topic_name
}
`)
}
