---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_flow_log"
description: |-
  Provides a VPC/Subnet/ENI Flow Log
---

# Resource: aws_flow_log

Provides a VPC/Subnet/ENI/Transit Gateway/Transit Gateway Attachment Flow Log to capture IP traffic for a specific network
interface, subnet, or VPC. Logs are sent to a CloudWatch Log Group, a S3 Bucket, or Amazon Data Firehose

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

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["vpc-flow-logs.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "example" {
  name   = "example"
  role   = aws_iam_role.example.id
  policy = data.aws_iam_policy_document.example.json
}
```

### Amazon Data Firehose logging

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

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["firehose.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "firehose_test_role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "example" {
  effect = "Allow"

  actions = [
    "logs:CreateLogDelivery",
    "logs:DeleteLogDelivery",
    "logs:ListLogDeliveries",
    "logs:GetLogDelivery",
    "firehose:TagDeliveryStream",
  ]

  resources = ["*"]
}

resource "aws_iam_role_policy" "example" {
  name   = "test"
  role   = aws_iam_role.example.id
  policy = data.aws_iam_policy_document.example.json
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

### Cross-Account Amazon Data Firehose Logging

The following example shows how to set up a flow log in one AWS account (source) that sends logs to an Amazon Data Firehose delivery stream in another AWS account (destination).
See the [AWS Documentation](https://docs.aws.amazon.com/vpc/latest/userguide/flow-logs-firehose.html).

```terraform
# Provider configurations
provider "aws" {
  profile = "admin-src"
}

provider "aws" {
  alias   = "destination_account"
  profile = "admin-dst"
}

# For source account
resource "aws_vpc" "src" {
  # config...
}

data "aws_iam_policy_document" "src_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "src" {
  name               = "tf-example-mySourceRole"
  assume_role_policy = data.aws_iam_policy_document.src_assume_role_policy.json
}

data "aws_iam_policy_document" "src_role_policy" {
  statement {
    effect    = "Allow"
    actions   = ["iam:PassRole"]
    resources = [aws_iam_role.src.arn]

    condition {
      test     = "StringEquals"
      variable = "iam:PassedToService"
      values   = ["delivery.logs.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "iam:AssociatedResourceARN"
      values   = [aws_vpc.src.arn]
    }
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogDelivery",
      "logs:DeleteLogDelivery",
      "logs:ListLogDeliveries",
      "logs:GetLogDelivery"
    ]
    resources = ["*"]
  }

  statement {
    effect    = "Allow"
    actions   = ["sts:AssumeRole"]
    resources = [aws_iam_role.dst.arn]
  }
}

resource "aws_iam_role_policy" "src_policy" {
  name   = "tf-example-mySourceRolePolicy"
  role   = aws_iam_role.src.name
  policy = data.aws_iam_policy_document.src_role_policy.json
}

resource "aws_flow_log" "src" {
  log_destination_type       = "kinesis-data-firehose"
  log_destination            = aws_kinesis_firehose_delivery_stream.dst.arn
  traffic_type               = "ALL"
  vpc_id                     = aws_vpc.src.id
  iam_role_arn               = aws_iam_role.src.arn
  deliver_cross_account_role = aws_iam_role.dst.arn
}

# For destination account
data "aws_iam_policy_document" "dst_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.src.arn]
    }
  }
}

resource "aws_iam_role" "dst" {
  provider           = aws.destination_account
  name               = "AWSLogDeliveryFirehoseCrossAccountRole" # must start with "AWSLogDeliveryFirehoseCrossAccountRolePolicy"
  assume_role_policy = data.aws_iam_policy_document.dst_assume_role_policy.json
}

data "aws_iam_policy_document" "dst_role_policy" {
  statement {
    effect = "Allow"
    actions = [
      "iam:CreateServiceLinkedRole",
      "firehose:TagDeliveryStream"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "dst" {
  provider = aws.destination_account
  name     = "AWSLogDeliveryFirehoseCrossAccountRolePolicy"
  role     = aws_iam_role.dst.name
  policy   = data.aws_iam_policy_document.dst_role_policy.json
}

resource "aws_kinesis_firehose_delivery_stream" "dst" {
  provider = aws.destination_account
  # The tag named "LogDeliveryEnabled" must be set to "true" to allow the service-linked role "AWSServiceRoleForLogDelivery"
  # to perform permitted actions on your behalf.
  # See: https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/AWS-logs-infrastructure-Firehose.html
  tags = {
    LogDeliveryEnabled = "true"
  }
  # other config...
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `deliver_cross_account_role` - (Optional) ARN of the IAM role in the destination account used for cross-account delivery of flow logs.
* `destination_options` - (Optional) Describes the destination options for a flow log. More details below.
* `eni_id` - (Optional) Elastic Network Interface ID to attach to.
* `iam_role_arn` - (Optional) ARN of the IAM role used to post flow logs. Corresponds to `DeliverLogsPermissionArn` in the [AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateFlowLogs.html).
* `log_destination_type` - (Optional) Logging destination type. Valid values: `cloud-watch-logs`, `s3`, `kinesis-data-firehose`. Default: `cloud-watch-logs`.
* `log_destination` - (Optional) ARN of the logging destination.
* `log_format` - (Optional) The fields to include in the flow log record. Accepted format example: `"$${interface-id} $${srcaddr} $${dstaddr} $${srcport} $${dstport}"`.
* `max_aggregation_interval` - (Optional) The maximum interval of time during which a flow of packets is captured and aggregated into a flow log record.
  Valid Values: `60` seconds (1 minute) or `600` seconds (10 minutes). Default: `600`.
  When `transit_gateway_id` or `transit_gateway_attachment_id` is specified, `max_aggregation_interval` *must* be 60 seconds (1 minute).
* `regional_nat_gateway_id` - (Optional) Regional NAT Gateway ID to attach to.
* `subnet_id` - (Optional) Subnet ID to attach to.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `traffic_type` - (Optional) The type of traffic to capture. Valid values: `ACCEPT`,`REJECT`, `ALL`. Required if `eni_id`, `regional_nat_gateway_id`, `subnet_id`, or `vpc_id` is specified.
* `transit_gateway_id` - (Optional) Transit Gateway ID to attach to.
* `transit_gateway_attachment_id` - (Optional) Transit Gateway Attachment ID to attach to.
* `vpc_id` - (Optional) VPC ID to attach to.

~> **NOTE:** One of `eni_id`, `regional_nat_gateway_id`, `subnet_id`, `transit_gateway_id`, `transit_gateway_attachment_id`, or `vpc_id` must be specified.

### destination_options

Describes the destination options for a flow log.

* `file_format` - (Optional) File format for the flow log. Default value: `plain-text`. Valid values: `plain-text`, `parquet`.
* `hive_compatible_partitions` - (Optional) Indicates whether to use Hive-compatible prefixes for flow logs stored in Amazon S3. Default value: `false`.
* `per_hour_partition` - (Optional) Indicates whether to partition the flow log per hour. This reduces the cost and response time for queries. Default value: `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Flow Log ID.
* `arn` - ARN of the Flow Log.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Flow Logs using the `id`. For example:

```terraform
import {
  to = aws_flow_log.test_flow_log
  id = "fl-1a2b3c4d"
}
```

Using `terraform import`, import Flow Logs using the `id`. For example:

```console
% terraform import aws_flow_log.test_flow_log fl-1a2b3c4d
```
