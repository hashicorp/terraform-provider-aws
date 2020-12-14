package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsDynamodbKinesisStreamingDestination_basic(t *testing.T) {
	var conf dynamodb.DescribeKinesisStreamingDestinationOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	streamName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_dynamodb_kinesis_streaming_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSDynamodbKinesisStreamingDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDynamodbKinesisStreamingDestinationConfigBasic(tableName, streamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamodbKinesisStreamingDestinationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "table_name", tableName),
					testAccMatchResourceAttrRegionalARN(resourceName, "stream_arn", "kinesis", regexp.MustCompile(fmt.Sprintf("stream/%s", streamName))),
				),
			},
			{
				Config: testAccAwsDynamodbKinesisStreamingDestinationConfigBase(tableName, streamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamodbKinesisStreamingDestinationNotExist,
				),
			},
		},
	})
}

func TestAccAwsDynamodbKinesisStreamingDestination_disappears(t *testing.T) {
	var conf dynamodb.DescribeKinesisStreamingDestinationOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	streamName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_dynamodb_kinesis_streaming_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSDynamodbKinesisStreamingDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDynamodbKinesisStreamingDestinationConfigBasic(tableName, streamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamodbKinesisStreamingDestinationExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDynamodbKinesisStreamingDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsDynamodbKinesisStreamingDestination_disappears_DynamoTable(t *testing.T) {
	var conf dynamodb.DescribeKinesisStreamingDestinationOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	streamName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_dynamodb_kinesis_streaming_destination.test"
	parentResourceName := "aws_dynamodb_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSDynamodbKinesisStreamingDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDynamodbKinesisStreamingDestinationConfigBasic(tableName, streamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamodbKinesisStreamingDestinationExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDynamoDbTable(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsDynamodbKinesisStreamingDestinationConfigBase(tableName string, streamName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "hk"

  attribute {
    name = "hk"
    type = "S"
  }
}

resource "aws_kinesis_stream" "test" {
  name        = "%s"
  shard_count = 2
}
`, tableName, streamName)
}

func testAccAwsDynamodbKinesisStreamingDestinationConfigBasic(tableName string, streamName string) string {
	return composeConfig(testAccAwsDynamodbKinesisStreamingDestinationConfigBase(tableName, streamName),
		`
resource "aws_dynamodb_kinesis_streaming_destination" "test" {
	table_name = aws_dynamodb_table.test.name
	stream_arn = aws_kinesis_stream.test.arn
}
`)
}

func testAccCheckDynamodbKinesisStreamingDestinationExists(n string, conf *dynamodb.DescribeKinesisStreamingDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn
		describeOpts := &dynamodb.DescribeKinesisStreamingDestinationInput{
			TableName: aws.String(rs.Primary.Attributes["table_name"]),
		}
		resp, err := conn.DescribeKinesisStreamingDestination(describeOpts)
		if err != nil {
			return err
		}

		*conf = *resp

		return nil
	}
}

func testAccCheckAWSDynamodbKinesisStreamingDestinationNotExist(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_kinesis_streaming_destination" {
			continue
		}

		log.Printf("[DEBUG] Checking if DynamoDB kinesis streaming destination %s exists", rs.Primary.ID)
		describeOpts := &dynamodb.DescribeKinesisStreamingDestinationInput{
			TableName: aws.String(rs.Primary.Attributes["table_name"]),
		}

		resp, err := conn.DescribeKinesisStreamingDestination(describeOpts)
		if err != nil {
			return err
		}

		if len(resp.KinesisDataStreamDestinations) > 0 {
			return fmt.Errorf("Error: DynamoDB kinesis streaming destination still exists %s", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

func testAccCheckAWSDynamodbKinesisStreamingDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_table" {
			continue
		}

		log.Printf("[DEBUG] Checking if DynamoDB table %s exists", rs.Primary.ID)
		// Check if queue exists by checking for its attributes
		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTable(params)
		if err == nil {
			return fmt.Errorf("DynamoDB table %s still exists. Failing!", rs.Primary.ID)
		}

		// Verify the error is what we want
		if dbErr, ok := err.(awserr.Error); ok && dbErr.Code() == "ResourceNotFoundException" {
			return nil
		}

		return err
	}

	return nil
}
