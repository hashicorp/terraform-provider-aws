package kinesis_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kinesis"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKinesisStreamDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kinesis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckStreamDataSourceConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_timestamp"),
					resource.TestCheckResourceAttr(dataSourceName, "closed_shards.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "open_shards.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "retention_period", "72"),
					resource.TestCheckResourceAttr(dataSourceName, "shard_level_metrics.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(dataSourceName, "stream_mode_details.0.stream_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccCheckStreamDataSourceConfig(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "closed_shards.#", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "open_shards.#", "3"),
				),
			},
		},
	})
}

func testAccCheckStreamDataSourceConfig(rName string, shardCount int) string {
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
