---
layout: "aws"
page_title: "AWS: aws_iot_topic_rule"
sidebar_current: "docs-aws-resource-iot-topic-rule"
description: |-
    Creates and manages an AWS IoT topic rule
---

# aws_iot_topic_rule

## Example Usage

```
resource "aws_iot_topic_rule" "rule" {
  name = "MyRule"
  description = "Example rule"
  enabled = true
  sql = "SELECT * FROM 'topic/test'"
  sql_version = ""

  sns {
    message_format = "RAW"
    role_arn = "${aws_iam_role.role.arn}"
    target_arn = "${aws_sns_topic.mytopic.arn}"
  }
}

resource "aws_sns_topic" "mytopic" {
  name = "mytopic"
}

resource "aws_iam_role" "role" {
    name = "myrole"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "iot.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
    name = "mypolicy"
    role = "${aws_iam_role.role.id}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Effect": "Allow",
        "Action": [
            "sns:Publish"
        ],
        "Resource": "${aws_sns_topic.mytopic.arn}"
    }
  ]
}
EOF
}
```

## Argument Reference

* `name` - (Required) Name of the topic rule
* `description` - (Optional) Human readable description of the topic rule
* `enabled` - (Required) Boolean flag to indicate if the topic rule is enabled
* `sql` - (Required) The SQL statement of the topic rule
* `sql_version` - (Required) Version of the SQL statement

The `cloudwatch_alarm` object takes the following arguments:

* `alarm_name` - (Required) The CloudWatch alarm name
* `role_arn` - (Required) The IAM role arn that allows to access the CloudWatch alarm
* `state_reason` - (Required) The reason for the alarm change
* `state_value` - (Required) The value of the alarm state. Acceptable values are: OK, ALARM, INSUFFICIENT_DATA

The `cloudwatch_metric` object takes the following arguments:

* `metric_name` - (Required) The CloudWatch metric name
* `metric_namespace` - (Required) The CloudWatch metric namespace
* `metric_timestamp` - (Optional) The CloudWatch metric timestamp
* `metric_unit` - (Required) The CloudWatch metric unit (supported units can be found here: http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Unit)
* `metric_value` - (Required) The CloudWatch metric value
* `role_arn` - (Required) The IAM role arn that allows to access the CloudWatch metric

The `dynamodb` object takes the following arguments:

* `hash_key_field` - (Required) The hash key field
* `hash_key_type` - (Optional) The hash key type, can be STRING or NUMBER
* `hash_key_value` - (Required) The hash key value
* `payload_field` - (Optional) The action payload
* `range_key_field` - (Optional) The range key field
* `range_key_type` - (Optional) The range key type, can be STRING or NUMBer
* `range_key_value` - (Optional) The range key value
* `role_arn` - (Required) The IAM role arn that allows to access the DynamoDB table
* `table_name` - (Required) The DynamoDB table name

The `elasticsearch` object takes the following arguments:

* `endpoint` - (Required) The ElasticSearch endpoint
* `id` - (Required) Unique ID for the document
* `index` - (Required) The ElasticSearch index
* `role_arn` - (Required) The IAM role arn that allows to access the ElasticSearch domain
* `type` - (Required) The type of the document

The `firehose` object takes the following arguments:

* `delivery_stream_name` - (Required) The name of the Firehose delivery stream
* `role_arn` - (Required) The IAM role arn that allows to access the Firehose delivery stream

The `kinesis` object takes the following arguments:

* `partition_key` - (Optional) The partition key
* `role_arn` - (Required) The IAM role arn that allows to access the Kinesis stream
* `stream_name` - (Required) The Kinesis stream name

The `lambda` object takes the following arguments:

* `function_arn` - (Required) The arn of the lambda function

The `republish` object takes the following arguments:

* `role_arn` - (Required) The IAM role arn that allows to access the topic
* `topic` - (Required) The topic the message should be republished to

The `s3` object takes the following arguments:

* `bucket_name` - (Required) The name of the S3 bucket
* `key` - (Required) The key of the object
* `role_arn` - (Required) The IAM role arn that allows to access the S3 bucket

The `sns` object takes the following arguments:

* `message_format` - (Required) The message format, allowed values are "RAW" or "JSON"
* `role_arn` - (Required) The IAM role arn that allows to access the SNS topic
* `target_arn` - (Required) The arn of the SNS topic

The `sqs` object takes the following arguments:

* `queue_url` - (Required) The URL of the SQS queue
* `role_arn` - (Required) The IAM role arn that allows to access the SQS queue
* `use_base64` - (Required) Boolean to enable base64 encoding

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the topic rule
