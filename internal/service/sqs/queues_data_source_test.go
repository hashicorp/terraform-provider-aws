package sqs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSQSQueuesDataSource_queueNamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_sqs_queues.test"
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueuesDataSourceConfig_queueNamePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(dataSourceName, "queue_urls.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "queue_urls.0", resourceName, "url"),
				),
			},
		},
	})
}

func testAccQueuesDataSourceConfig_queueNamePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "wrong" {
  name = "wrong_%[1]s"
}

data "aws_sqs_queues" "test" {
  queue_name_prefix = aws_sqs_queue.test.name
}
`, rName)
}
