package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSKinesisFirehoseDeliveryStream_s3basic(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3basic,
		ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     testAccKinesisFirehosePreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3WithCloudwatchLogging(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     testAccKinesisFirehosePreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3WithCloudwatchLogging(os.Getenv("AWS_ACCOUNT_ID"), ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3ConfigUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription

	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3basic,
		ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri)
	postConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3Updates,
		ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri)

	updatedS3DestinationConfig := &firehose.S3DestinationDescription{
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64(400),
			SizeInMBs:         aws.Int64(10),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     testAccKinesisFirehosePreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, updatedS3DestinationConfig, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3basic(t *testing.T) {

	rSt := acctest.RandString(5)
	rName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rSt)

	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()
	config := testAccFirehoseAWSLambdaConfigBasic(rName, rSt) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic,
			ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     testAccKinesisFirehosePreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3InvalidProcessorType(t *testing.T) {

	rSt := acctest.RandString(5)
	rName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rSt)

	ri := acctest.RandInt()
	config := testAccFirehoseAWSLambdaConfigBasic(rName, rSt) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3InvalidProcessorType,
			ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     testAccKinesisFirehosePreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("must be 'Lambda'"),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3InvalidParameterName(t *testing.T) {

	rSt := acctest.RandString(5)
	rName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rSt)

	ri := acctest.RandInt()
	config := testAccFirehoseAWSLambdaConfigBasic(rName, rSt) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3InvalidParameterName,
			ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     testAccKinesisFirehosePreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("must be one of 'LambdaArn', 'NumberOfRetries'"),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3Updates(t *testing.T) {

	rSt := acctest.RandString(5)
	rName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rSt)

	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()

	preConfig := testAccFirehoseAWSLambdaConfigBasic(rName, rSt) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic,
			ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri)
	postConfig := testAccFirehoseAWSLambdaConfigBasic(rName, rSt) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3Updates,
			ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri)

	updatedExtendedS3DestinationConfig := &firehose.ExtendedS3DestinationDescription{
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64(400),
			SizeInMBs:         aws.Int64(10),
		},
		ProcessingConfiguration: &firehose.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []*firehose.Processor{
				&firehose.Processor{
					Type: aws.String("Lambda"),
					Parameters: []*firehose.ProcessorParameter{
						&firehose.ProcessorParameter{
							ParameterName:  aws.String("LambdaArn"),
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     testAccKinesisFirehosePreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, updatedExtendedS3DestinationConfig, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_RedshiftConfigUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription

	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_RedshiftBasic,
		ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri, ri)
	postConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_RedshiftUpdates,
		ri, os.Getenv("AWS_ACCOUNT_ID"), ri, ri, ri, ri)

	updatedRedshiftConfig := &firehose.RedshiftDestinationDescription{
		CopyCommand: &firehose.CopyCommand{
			CopyOptions: aws.String("GZIP"),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     testAccKinesisFirehosePreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, updatedRedshiftConfig, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ElasticsearchConfigUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription

	ri := acctest.RandInt()
	awsAccountId := os.Getenv("AWS_ACCOUNT_ID")
	preConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchBasic,
		ri, awsAccountId, ri, ri, ri, awsAccountId, awsAccountId, ri, ri)
	postConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchUpdate,
		ri, awsAccountId, ri, ri, ri, awsAccountId, awsAccountId, ri, ri)

	updatedElasticSearchConfig := &firehose.ElasticsearchDestinationDescription{
		BufferingHints: &firehose.ElasticsearchBufferingHints{
			IntervalInSeconds: aws.Int64(500),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     testAccKinesisFirehosePreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream_es", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test_stream_es", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, updatedElasticSearchConfig),
				),
			},
		},
	})
}

func testAccCheckKinesisFirehoseDeliveryStreamExists(n string, stream *firehose.DeliveryStreamDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		log.Printf("State: %#v", s.RootModule().Resources)
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Firehose ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).firehoseconn
		describeOpts := &firehose.DescribeDeliveryStreamInput{
			DeliveryStreamName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeDeliveryStream(describeOpts)
		if err != nil {
			return err
		}

		*stream = *resp.DeliveryStreamDescription

		return nil
	}
}

func testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(stream *firehose.DeliveryStreamDescription, s3config interface{}, extendedS3config interface{}, redshiftConfig interface{}, elasticsearchConfig interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*stream.DeliveryStreamName, "terraform-kinesis-firehose") {
			return fmt.Errorf("Bad Stream name: %s", *stream.DeliveryStreamName)
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_firehose_delivery_stream" {
				continue
			}
			if *stream.DeliveryStreamARN != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Delivery Stream ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["arn"], *stream.DeliveryStreamARN)
			}

			if s3config != nil {
				s := s3config.(*firehose.S3DestinationDescription)
				// Range over the Stream Destinations, looking for the matching S3
				// destination. For simplicity, our test only have a single S3 or
				// Redshift destination, so at this time it's safe to match on the first
				// one
				var match bool
				for _, d := range stream.Destinations {
					if d.S3DestinationDescription != nil {
						if *d.S3DestinationDescription.BufferingHints.SizeInMBs == *s.BufferingHints.SizeInMBs {
							match = true
						}
					}
				}
				if !match {
					return fmt.Errorf("Mismatch s3 buffer size, expected: %s, got: %s", s, stream.Destinations)
				}
			}

			if extendedS3config != nil {
				es := extendedS3config.(*firehose.ExtendedS3DestinationDescription)

				// Range over the Stream Destinations, looking for the matching S3
				// destination. For simplicity, our test only have a single S3 or
				// Redshift destination, so at this time it's safe to match on the first
				// one
				var match, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.ExtendedS3DestinationDescription != nil {
						if *d.ExtendedS3DestinationDescription.BufferingHints.SizeInMBs == *es.BufferingHints.SizeInMBs {
							match = true
						}

						processingConfigMatch = len(es.ProcessingConfiguration.Processors) == len(d.ExtendedS3DestinationDescription.ProcessingConfiguration.Processors)
					}
				}
				if !match {
					return fmt.Errorf("Mismatch extended s3 buffer size, expected: %s, got: %s", es, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch extended s3 ProcessingConfiguration.Processors count, expected: %s, got: %s", es, stream.Destinations)
				}
			}

			if redshiftConfig != nil {
				r := redshiftConfig.(*firehose.RedshiftDestinationDescription)
				// Range over the Stream Destinations, looking for the matching Redshift
				// destination
				var match bool
				for _, d := range stream.Destinations {
					if d.RedshiftDestinationDescription != nil {
						if *d.RedshiftDestinationDescription.CopyCommand.CopyOptions == *r.CopyCommand.CopyOptions {
							match = true
						}
					}
				}
				if !match {
					return fmt.Errorf("Mismatch Redshift CopyOptions, expected: %s, got: %s", r, stream.Destinations)
				}
			}

			if elasticsearchConfig != nil {
				es := elasticsearchConfig.(*firehose.ElasticsearchDestinationDescription)
				// Range over the Stream Destinations, looking for the matching Elasticsearch destination
				var match bool
				for _, d := range stream.Destinations {
					if d.ElasticsearchDestinationDescription != nil {
						match = true
					}
				}
				if !match {
					return fmt.Errorf("Mismatch Elasticsearch Buffering Interval, expected: %s, got: %s", es, stream.Destinations)
				}
			}
		}
		return nil
	}
}

func testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3(s *terraform.State) error {
	err := testAccCheckKinesisFirehoseDeliveryStreamDestroy(s)

	if err == nil {
		err = testAccCheckFirehoseLambdaFunctionDestroy(s)
	}

	return err
}

func testAccCheckKinesisFirehoseDeliveryStreamDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_firehose_delivery_stream" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).firehoseconn
		describeOpts := &firehose.DescribeDeliveryStreamInput{
			DeliveryStreamName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeDeliveryStream(describeOpts)
		if err == nil {
			if resp.DeliveryStreamDescription != nil && *resp.DeliveryStreamDescription.DeliveryStreamStatus != "DELETING" {
				return fmt.Errorf("Error: Delivery Stream still exists")
			}
		}

		return nil

	}

	return nil
}

func testAccCheckFirehoseLambdaFunctionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_function" {
			continue
		}

		_, err := conn.GetFunction(&lambda.GetFunctionInput{
			FunctionName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Lambda Function still exists")
		}
	}

	return nil
}

func testAccKinesisFirehosePreCheck(t *testing.T) func() {
	return func() {
		testAccPreCheck(t)
		if os.Getenv("AWS_ACCOUNT_ID") == "" {
			t.Fatal("AWS_ACCOUNT_ID must be set")
		}
	}
}

func baseAccFirehoseAWSLambdaConfig(rst string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "iam_policy_for_lambda" {
    name = "iam_policy_for_lambda_%s"
    role = "${aws_iam_role.iam_for_lambda.id}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
	{
		"Effect": "Allow",
		"Action": [
			"logs:CreateLogGroup",
			"logs:CreateLogStream",
			"logs:PutLogEvents"
		],
		"Resource": "arn:aws:logs:*:*:*"
	},
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
    name = "iam_for_lambda_%s"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, rst, rst)
}

func testAccFirehoseAWSLambdaConfigBasic(rName, rSt string) string {
	return fmt.Sprintf(baseAccFirehoseAWSLambdaConfig(rSt)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs4.3"
}
`, rName)
}

const testAccKinesisFirehoseDeliveryStreamBaseConfig = `
resource "aws_iam_role" "firehose" {
  name = "tf_acctest_firehose_delivery_role_%d"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "%s"
        }
      }
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "private"
}

resource "aws_iam_role_policy" "firehose" {
  name = "tf_acctest_firehose_delivery_policy_%d"
  role = "${aws_iam_role.firehose.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::${aws_s3_bucket.bucket.id}",
        "arn:aws:s3:::${aws_s3_bucket.bucket.id}/*"
      ]
    }
  ]
}
EOF
}

`

func testAccKinesisFirehoseDeliveryStreamConfig_s3WithCloudwatchLogging(accountId string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "firehose" {
  name = "tf_acctest_firehose_delivery_role_%d"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "%s"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose" {
  name = "tf_acctest_firehose_delivery_policy_%d"
  role = "${aws_iam_role.firehose.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::${aws_s3_bucket.bucket.id}",
        "arn:aws:s3:::${aws_s3_bucket.bucket.id}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:putLogEvents"
      ],
      "Resource": [
        "arn:aws:logs::log-group:/aws/kinesisfirehose/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "private"
}

resource "aws_cloudwatch_log_group" "test" {
  name = "example-%d"
}

resource "aws_cloudwatch_log_stream" "test" {
  name = "sample-log-stream-test-%d"
  log_group_name = "${aws_cloudwatch_log_group.test.name}"
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-cloudwatch-%d"
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    cloudwatch_logging_options {
	enabled = true
	log_group_name = "${aws_cloudwatch_log_group.test.name}"
	log_stream_name = "${aws_cloudwatch_log_stream.test.name}"
    }
  }
}
`, rInt, accountId, rInt, rInt, rInt, rInt, rInt)
}

var testAccKinesisFirehoseDeliveryStreamConfig_s3basic = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_s3Updates = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-s3test-%d"
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    buffer_size = 10
    buffer_interval = 400
    compression_format = "GZIP"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    processing_configuration = [{
    	enabled = "false",
    	processors = [{
    		type = "Lambda"
    		parameters = [{
    			parameter_name = "LambdaArn"
    			parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
    		}]
    	}]
    }]
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3InvalidProcessorType = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    processing_configuration = [{
    	enabled = "false",
    	processors = [{
    		type = "NotLambda"
    		parameters = [{
    			parameter_name = "LambdaArn"
    			parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
    		}]
    	}]
    }]
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3InvalidParameterName = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    processing_configuration = [{
    	enabled = "false",
    	processors = [{
    		type = "Lambda"
    		parameters = [{
    			parameter_name = "NotLambdaArn"
    			parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
    		}]
    	}]
    }]
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3Updates = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    processing_configuration = [{
    	enabled = "false",
    	processors = [{
    		type = "Lambda"
    		parameters = [{
    			parameter_name = "LambdaArn"
    			parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
    		}]
    	}]
    }]
    buffer_size = 10
    buffer_interval = 400
    compression_format = "GZIP"
  }
}
`

var testAccKinesisFirehoseDeliveryStreamBaseRedshiftConfig = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_redshift_cluster" "test_cluster" {
  cluster_identifier = "tf-redshift-cluster-%d"
  database_name = "test"
  master_username = "testuser"
  master_password = "T3stPass"
  node_type = "dc1.large"
  cluster_type = "single-node"
	skip_final_snapshot = true
}`

var testAccKinesisFirehoseDeliveryStreamConfig_RedshiftBasic = testAccKinesisFirehoseDeliveryStreamBaseRedshiftConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  depends_on = ["aws_iam_role_policy.firehose", "aws_redshift_cluster.test_cluster"]
  name = "terraform-kinesis-firehose-basicredshifttest-%d"
  destination = "redshift"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
  redshift_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    cluster_jdbcurl = "jdbc:redshift://${aws_redshift_cluster.test_cluster.endpoint}/${aws_redshift_cluster.test_cluster.database_name}"
    username = "testuser"
    password = "T3stPass"
    data_table_name = "test-table"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_RedshiftUpdates = testAccKinesisFirehoseDeliveryStreamBaseRedshiftConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  depends_on = ["aws_iam_role_policy.firehose", "aws_redshift_cluster.test_cluster"]
  name = "terraform-kinesis-firehose-basicredshifttest-%d"
  destination = "redshift"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    buffer_size = 10
    buffer_interval = 400
    compression_format = "GZIP"
  }
  redshift_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    cluster_jdbcurl = "jdbc:redshift://${aws_redshift_cluster.test_cluster.endpoint}/${aws_redshift_cluster.test_cluster.database_name}"
    username = "testuser"
    password = "T3stPass"
    data_table_name = "test-table"
    copy_options = "GZIP"
    data_table_columns = "test-col"
  }
}`

var testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = "es-test-%d"
  cluster_config {
    instance_type = "r3.large.elasticsearch"
  }

  access_policies = <<CONFIG
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::%s:root"
      },
      "Action": "es:*",
      "Resource": "arn:aws:es:us-east-1:%s:domain/es-test-%d/*"
    }
  ]
}
CONFIG
}`

var testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchBasic = testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream_es" {
  depends_on = ["aws_iam_role_policy.firehose", "aws_elasticsearch_domain.test_cluster"]
  name = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
  elasticsearch_configuration {
    domain_arn = "${aws_elasticsearch_domain.test_cluster.arn}"
    role_arn = "${aws_iam_role.firehose.arn}"
    index_name = "test"
    type_name = "test"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchUpdate = testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test_stream_es" {
  depends_on = ["aws_iam_role_policy.firehose", "aws_elasticsearch_domain.test_cluster"]
  name = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
  elasticsearch_configuration {
    domain_arn = "${aws_elasticsearch_domain.test_cluster.arn}"
    role_arn = "${aws_iam_role.firehose.arn}"
    index_name = "test"
    type_name = "test"
    buffering_interval = 500
  }
}`
