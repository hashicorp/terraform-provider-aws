package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSKinesisStreamConsumer_basic(t *testing.T) {
	var consumer kinesis.ConsumerDescription

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer", &consumer),
					testAccCheckAWSKinesisStreamConsumerAttributes(&consumer),
				),
			},
		},
	})
}

func TestAccAWSKinesisStreamConsumer_createMultipleConcurrentStreamConsumers(t *testing.T) {
	var consumer kinesis.ConsumerDescription

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConsumerConfigConcurrent(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.0", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.1", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.2", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.3", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.4", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.5", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.6", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.7", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.8", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.9", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.10", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.11", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.12", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.13", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.14", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.15", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.16", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.17", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.18", &consumer),
					testAccCheckKinesisStreamConsumerExists("aws_kinesis_stream_consumer.test_stream_consumer.19", &consumer),
				),
			},
		},
	})
}

func TestAccAWSKinesisStreamConsumer_importBasic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_kinesis_stream_consumer.test_stream_consumer"
	consumerName := fmt.Sprintf("terraform-kinesis-stream-consumer-test-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConsumerConfig(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     consumerName,
			},
		},
	})
}

func testAccCheckKinesisStreamConsumerExists(n string, stream *kinesis.ConsumerDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Stream Consumer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisconn
		describeOpts := &kinesis.DescribeStreamConsumerInput{
			ConsumerName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeStreamConsumer(describeOpts)
		if err != nil {
			return err
		}

		*stream = *resp.ConsumerDescription

		return nil
	}
}

func testAccCheckAWSKinesisStreamConsumerAttributes(stream *kinesis.ConsumerDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*stream.ConsumerName, "terraform-kinesis-stream-consumer-test") {
			return fmt.Errorf("Bad Stream Consumer name: %s", *stream.ConsumerName)
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_stream_consumer" {
				continue
			}
			if *stream.ConsumerARN != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Stream Consumer ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["arn"], *stream.ConsumerARN)
			}
		}
		return nil
	}
}

func testAccCheckKinesisStreamConsumerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_stream_consumer" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).kinesisconn
		describeOpts := &kinesis.DescribeStreamConsumerInput{
			ConsumerName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeStreamConsumer(describeOpts)
		if err == nil {
			if resp.ConsumerDescription != nil && *resp.ConsumerDescription.ConsumerStatus != "DELETING" {
				return fmt.Errorf("Error: Stream Consumer still exists")
			}
		}

		return nil

	}

	return nil
}

func testAccKinesisStreamConsumerConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test_stream_consumer" {
	name = "terraform-kinesis-stream-consumer-test-%d"
	stream_arn = "${aws_kinesis_stream.stream.arn}"
}`, rInt)
}

func testAccKinesisStreamConsumerConfigConcurrent(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test_stream_consumer" {
        count = 20
	name = "terraform-kinesis-stream-consumer-test-%d-${count.index}"
	stream_arn = "${aws_kinesis_stream.stream.arn}"
}`, rInt)
}
