package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSKinesisStreamConsumerDataSource(t *testing.T) {
	var stream kinesis.StreamDescription

	cn := fmt.Sprintf("terraform-kinesis-stream-consumer-test-%d", acctest.RandInt())
	config := fmt.Sprintf(testAccCheckAwsKinesisStreamConsumerDataSourceConfig, cn)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream_consumer.test_stream_consumer", &stream),
					resource.TestCheckResourceAttrSet("data.aws_kinesis_stream_consumer.test_stream_consumer", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_kinesis_stream_consumer.test_stream_consumer", "stream_arn"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream_consumer.test_stream_consumer", "name", cn),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream_consumer.test_stream_consumer", "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet("data.aws_kinesis_stream_consumer.test_stream_consumer", "creation_timestamp"),
				),
			},
		},
	})
}

var testAccCheckAwsKinesisStreamConsumerDataSourceConfig = `
resource "aws_kinesis_stream_consumer" "test_stream_consumer" {
	name = "%s"
	stream_arn = "${aws_kinesis_stream.stream.arn}"
}

data "aws_kinesis_stream_consumer" "test_stream_consumer" {
	name = "${aws_kinesis_stream.test_stream_consumer.name}"
}
`
