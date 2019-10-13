package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSKinesisStreamConsumer_basic(t *testing.T) {
	var stream kinesis.StreamDescription
	var consumer kinesis.ConsumerDescription
	config := createAccKinesisStreamConsumerConfig()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config.single(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(config.stream.getName(), &stream),
					testAccCheckKinesisStreamConsumerExists(config, -1, &consumer),
					testAccCheckAWSKinesisStreamConsumerAttributes(config, &consumer),
				),
			},
		},
	})
}

func TestAccAWSKinesisStreamConsumer_createMultipleConcurrentStreamConsumers(t *testing.T) {
	var stream kinesis.StreamDescription
	var consumer kinesis.ConsumerDescription
	config := createAccKinesisStreamConsumerConfig()

	var checkFunctions []resource.TestCheckFunc

	checkFunctions = append(
		checkFunctions,
		testAccCheckKinesisStreamExists(config.stream.getName(), &stream))

	for i := 0; i < config.count; i++ {
		checkFunctions = append(
			checkFunctions,
			testAccCheckKinesisStreamConsumerExists(config, i, &consumer))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config.multiple(),
				Check:  resource.ComposeTestCheckFunc(checkFunctions...),
			},
		},
	})
}

func testAccCheckKinesisStreamConsumerExists(c *accKinesisStreamConsumerConfig,
	index int, consumer *kinesis.ConsumerDescription) resource.TestCheckFunc {
	var resourceName string
	if index > -1 {
		resourceName = c.getIndexedName(index)
	} else {
		resourceName = c.getName()
	}

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Stream Consumer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisconn
		describeOpts := &kinesis.DescribeStreamConsumerInput{
			ConsumerName: aws.String(rs.Primary.Attributes["name"]),
			StreamARN:    aws.String(rs.Primary.Attributes["stream_arn"]),
		}
		resp, err := conn.DescribeStreamConsumer(describeOpts)
		if err != nil {
			return err
		}

		*consumer = *resp.ConsumerDescription

		return nil
	}
}

func testAccCheckAWSKinesisStreamConsumerAttributes(
	c *accKinesisStreamConsumerConfig, consumer *kinesis.ConsumerDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*consumer.ConsumerName, c.getConsumerName()) {
			return fmt.Errorf("Bad Stream Consumer name: %s", *consumer.ConsumerName)
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_stream_consumer" {
				continue
			}

			if *consumer.ConsumerARN != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Stream Consumer(%s) ARN\n\t expected: %s\n\tgot: %s\n", rs.Type, rs.Primary.Attributes["arn"], *consumer.ConsumerARN)
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
			StreamARN:    aws.String(rs.Primary.Attributes["stream_arn"]),
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

type accKinesisStreamConfig struct {
	resourceType      string
	resourceLocalName string
	streamBasename    string
	randInt           int
}

func (r *accKinesisStreamConfig) getName() string {
	return fmt.Sprintf("%s.%s", r.resourceType, r.resourceLocalName)
}

func (r *accKinesisStreamConfig) getStreamName() string {
	return fmt.Sprintf("%s-%d", r.streamBasename, r.randInt)
}

func (r *accKinesisStreamConfig) single() string {

	return fmt.Sprintf(`
resource "%s" "%s" {
	name = "%s"
	shard_count               = 2
  	enforce_consumer_deletion = true

  	tags = {
    	Name = "tf-test"
  	}
}`, r.resourceType, r.resourceLocalName, r.getStreamName())
}

type accKinesisStreamConsumerConfig struct {
	stream            *accKinesisStreamConfig
	resourceType      string
	resourceLocalName string
	consumerBasename  string
	count             int
	randInt           int
}

func (r *accKinesisStreamConsumerConfig) getName() string {
	return fmt.Sprintf("%s.%s", r.resourceType, r.resourceLocalName)
}

func (r *accKinesisStreamConsumerConfig) getIndexedName(index int) string {
	return fmt.Sprintf("%s.%s.%d", r.resourceType, r.resourceLocalName, index)
}

func (r *accKinesisStreamConsumerConfig) getConsumerName() string {
	return fmt.Sprintf("%s-%d", r.consumerBasename, r.randInt)
}

func (r *accKinesisStreamConsumerConfig) data() string {
	return fmt.Sprintf(`
%s

data "%s" "%s" {
	name = "${%s.name}"
	stream_arn = "${%s.arn}"
}`, r.single(), r.resourceType, r.resourceLocalName, r.getName(), r.stream.getName())
}

func (r *accKinesisStreamConsumerConfig) single() string {

	return fmt.Sprintf(`
%s

resource "%s" "%s" {
	name = "%s"
	stream_arn = "${%s.arn}"
}

`, r.stream.single(), r.resourceType, r.resourceLocalName, r.getConsumerName(), r.stream.getName())
}

func (r *accKinesisStreamConsumerConfig) multiple() string {

	return fmt.Sprintf(`
%s

resource "%s" "%s" {
	count = %d
	name = "%s-${count.index}"
	stream_arn = "${%s.arn}"
}`, r.stream.single(), r.resourceType, r.resourceLocalName, r.count, r.getConsumerName(), r.stream.getName())
}

func createAccKinesisStreamConsumerConfig() *accKinesisStreamConsumerConfig {
	iRnd := acctest.RandInt()
	return &accKinesisStreamConsumerConfig{
		stream: &accKinesisStreamConfig{
			resourceType:      "aws_kinesis_stream",
			resourceLocalName: "test_stream",
			streamBasename:    "terraform-kinesis-stream-test",
			randInt:           iRnd,
		},
		resourceType:      "aws_kinesis_stream_consumer",
		resourceLocalName: "test_stream_consumer",
		consumerBasename:  "terraform-kinesis-stream-consumer-test",
		count:             2,
		randInt:           iRnd,
	}
}
