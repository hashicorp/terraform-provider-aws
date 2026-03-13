// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaTopicDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var topic kafka.DescribeTopicOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_topic.test"
	dataSourceName := "data.aws_msk_topic.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicDataSourceConfig_basic(rName, clusterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &topic),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrARN), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("configs"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrName), resourceName, tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("partition_count"), resourceName, tfjsonpath.New("partition_count"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("replication_factor"), resourceName, tfjsonpath.New("replication_factor"), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccTopicDataSourceConfig_basic(rName, clusterName string) string {
	return acctest.ConfigCompose(testAccTopicConfig_basic(rName, clusterName, 2, 2), `
data "aws_msk_topic" "test" {
  name               = aws_msk_topic.test.name
  cluster_arn        = aws_msk_topic.test.cluster_arn
}
`)
}
