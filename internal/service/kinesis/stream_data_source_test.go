// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisStreamDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamDataSourceConfig_basic(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_timestamp"),
					resource.TestCheckResourceAttr(dataSourceName, "closed_shards.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "open_shards.#", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRetentionPeriod, "72"),
					resource.TestCheckResourceAttr(dataSourceName, "shard_level_metrics.#", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(dataSourceName, "stream_mode_details.0.stream_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccStreamDataSourceConfig_basic(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "closed_shards.#", acctest.Ct4),
					resource.TestCheckResourceAttr(dataSourceName, "open_shards.#", acctest.Ct3),
				),
			},
		},
	})
}

func testAccStreamDataSourceConfig_basic(rName string, shardCount int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name             = %[1]q
  shard_count      = %[2]d
  retention_period = 72

  tags = {
    Name = %[1]q
  }

  shard_level_metrics = [
    "IncomingBytes",
    "OutgoingBytes"
  ]
}

data "aws_kinesis_stream" "test" {
  name = aws_kinesis_stream.test.name
}
`, rName, shardCount)
}
