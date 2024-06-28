// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSQSQueuesDataSource_queueNamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[types.QueueAttributeName]string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_sqs_queues.test"
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueuesDataSourceConfig_queueNamePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(dataSourceName, "queue_urls.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "queue_urls.0", resourceName, names.AttrURL),
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
