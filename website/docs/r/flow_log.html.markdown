---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_flow_log"
description: |-
  Provides a VPC/Subnet/ENI Flow Log
---

# Resource: aws_flow_log

Provides a VPC/Subnet/ENI/Transit Gateway/Transit Gateway Attachment Flow Log to capture IP traffic for a specific network
interface, subnet, or VPC. Logs are sent to a CloudWatch Log Group, a S3 Bucket, or Amazon Kinesis Data Firehose

## Example Usage

### CloudWatch Logging

```terraform
resource "aws_flow_log" "example" {
  iam_role_arn    = aws_iam_role.example.arn
  log_destination = aws_cloudwatch_log_group.example.arn
  traffic_type    = "ALL"
  vpc_id          = aws_vpc.example.id
}

resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "vpc-flow-logs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "example" {
  name = "example"
  role = aws_iam_role.example.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
```

### Amazon Kinesis Data Firehose logging

```terraform
resource "aws_flow_log" "example" {
  log_destination      = aws_kinesis_firehose_delivery_stream.example.arn
  log_destination_type = "kinesis-data-firehose"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.example.id
}

resource "aws_kinesis_firehose_delivery_stream" "example" {
  name        = "kinesis_firehose_test"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.example.arn
    bucket_arn = aws_s3_bucket.example.arn
  }

  tags = {
    "LogDeliveryEnabled" = "true"
  }
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}

resource "aws_iam_role" "example" {
  name               = "firehose_test_role"
  assume_role_policy = <<EOF
 {
   "Version":"2012-10-17",
   "Statement": [
     {
       "Action":"sts:AssumeRole",
       "Principal":{
         "Service":"firehose.amazonaws.com"
       },
       "Effect":"Allow",
       "Sid":""
     }
   ]
 }
 EOF
}

resource "aws_iam_role_policy" "example" {
  name   = "test"
  role   = aws_iam_role.example.id
  policy = <<EOF
 {
   "Version":"2012-10-17",
   "Statement":[
     {
       "Action": [
         "logs:CreateLogDelivery",
         "logs:DeleteLogDelivery",
         "logs:ListLogDeliveries",
         "logs:GetLogDelivery",
         "firehose:TagDeliveryStream"
       ],
       "Effect":"Allow",
       "Resource":"*"
     }
   ]
 }
 EOF
}
```

### S3 Logging

```terraform
resource "aws_flow_log" "example" {
  log_destination      = aws_s3_bucket.example.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.example.id
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}
```

### S3 Logging in Apache Parquet format with per-hour partitions

```terraform
resource "aws_flow_log" "example" {
  log_destination      = aws_s3_bucket.example.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.example.id
  destination_options {
    file_format        = "parquet"
    per_hour_partition = true
  }
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}
```

## Argument Reference

~> **NOTE:** One of `eni_id`, `subnet_id`, `transit_gateway_id`, `transit_gateway_attachment_id`, or `vpc_id` must be specified.

The following arguments are supported:

* `traffic_type` - (Required) The type of traffic to capture. Valid values: `ACCEPT`,`REJECT`, `ALL`.
* `eni_id` - (Optional) Elastic Network Interface ID to attach to
* `iam_role_arn` - (Optional) The ARN for the IAM role that's used to post flow logs to a CloudWatch Logs log group
* `log_destination_type` - (Optional) The type of the logging destination. Valid values: `cloud-watch-logs`, `s3`, `kinesis-data-firehose`. Default: `cloud-watch-logs`.
* `log_destination` - (Optional) The ARN of the logging destination. Either `log_destination` or `log_group_name` must be set.
* `log_group_name` - (Optional) *Deprecated:* Use `log_destination` instead. The name of the CloudWatch log group. Either `log_group_name` or `log_destination` must be set.
* `subnet_id` - (Optional) Subnet ID to attach to
* `transit_gateway_id` - (Optional) Transit Gateway ID to attach to
* `transit_gateway_attachment_id` - (Optional) Transit Gateway Attachment ID to attach to
* `vpc_id` - (Optional) VPC ID to attach to
* `log_format` - (Optional) The fields to include in the flow log record, in the order in which they should appear.
* `max_aggregation_interval` - (Optional) The maximum interval of time
  during which a flow of packets is captured and aggregated into a flow
  log record. Valid Values: `60` seconds (1 minute) or `600` seconds (10
  minutes). Default: `600`. When `transit_gateway_id` or `transit_gateway_attachment_id` is specified, `max_aggregation_interval` _must_ be 60 seconds (1 minute).
* `destination_options` - (Optional) Describes the destination options for a flow log. More details below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### destination_options

Describes the destination options for a flow log.

* `file_format` - (Optional) The format for the flow log. Default value: `plain-text`. Valid values: `plain-text`, `parquet`.
* `hive_compatible_partitions` - (Optional) Indicates whether to use Hive-compatible prefixes for flow logs stored in Amazon S3. Default value: `false`.
* `per_hour_partition` - (Optional) Indicates whether to partition the flow log per hour. This reduces the cost and response time for queries. Default value: `false`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Flow Log ID
* `arn` - The ARN of the Flow Log.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Flow Logs can be imported using the `id`, e.g.,

```
$ terraform import aws_flow_log.test_flow_log fl-1a2b3c4d
```
