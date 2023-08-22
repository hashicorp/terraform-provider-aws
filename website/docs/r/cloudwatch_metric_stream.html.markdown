---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_metric_stream"
description: |-
  Provides a CloudWatch Metric Stream resource.
---

# Resource: aws_cloudwatch_metric_stream

Provides a CloudWatch Metric Stream resource.

## Example Usage

### Filters

```terraform
resource "aws_cloudwatch_metric_stream" "main" {
  name          = "my-metric-stream"
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"

  include_filter {
    namespace    = "AWS/EC2"
    metric_names = ["CPUUtilization", "NetworkOut"]
  }

  include_filter {
    namespace    = "AWS/EBS"
    metric_names = []
  }
}

# https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-metric-streams-trustpolicy.html
data "aws_iam_policy_document" "streams_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["streams.metrics.cloudwatch.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "metric_stream_to_firehose" {
  name               = "metric_stream_to_firehose_role"
  assume_role_policy = data.aws_iam_policy_document.streams_assume_role.json
}

# https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-metric-streams-trustpolicy.html
data "aws_iam_policy_document" "metric_stream_to_firehose" {
  statement {
    effect = "Allow"

    actions = [
      "firehose:PutRecord",
      "firehose:PutRecordBatch",
    ]

    resources = [aws_kinesis_firehose_delivery_stream.s3_stream.arn]
  }
}
resource "aws_iam_role_policy" "metric_stream_to_firehose" {
  name   = "default"
  role   = aws_iam_role.metric_stream_to_firehose.id
  policy = data.aws_iam_policy_document.metric_stream_to_firehose.json
}

resource "aws_s3_bucket" "bucket" {
  bucket = "metric-stream-test-bucket"
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

data "aws_iam_policy_document" "firehose_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["firehose.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "firehose_to_s3" {
  assume_role_policy = data.aws_iam_policy_document.firehose_assume_role.json
}

data "aws_iam_policy_document" "firehose_to_s3" {
  statement {
    effect = "Allow"

    actions = [
      "s3:AbortMultipartUpload",
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:ListBucket",
      "s3:ListBucketMultipartUploads",
      "s3:PutObject",
    ]

    resources = [
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*",
    ]
  }
}

resource "aws_iam_role_policy" "firehose_to_s3" {
  name   = "default"
  role   = aws_iam_role.firehose_to_s3.id
  policy = data.aws_iam_policy_document.firehose_to_s3.json
}

resource "aws_kinesis_firehose_delivery_stream" "s3_stream" {
  name        = "metric-stream-test-stream"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose_to_s3.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
```

### Additional Statistics

```terraform
resource "aws_cloudwatch_metric_stream" "main" {
  name          = "my-metric-stream"
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"

  statistics_configuration {
    additional_statistics = [
      "p1", "tm99"
    ]

    include_metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
    }
  }

  statistics_configuration {
    additional_statistics = [
      "TS(50.5:)"
    ]

    include_metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `firehose_arn` - (Required) ARN of the Amazon Kinesis Firehose delivery stream to use for this metric stream.
* `role_arn` - (Required) ARN of the IAM role that this metric stream will use to access Amazon Kinesis Firehose resources. For more information about role permissions, see [Trust between CloudWatch and Kinesis Data Firehose](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-metric-streams-trustpolicy.html).
* `output_format` - (Required) Output format for the stream. Possible values are `json` and `opentelemetry0.7`. For more information about output formats, see [Metric streams output formats](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-metric-streams-formats.html).

The following arguments are optional:

* `exclude_filter` - (Optional) List of exclusive metric filters. If you specify this parameter, the stream sends metrics from all metric namespaces except for the namespaces and the conditional metric names that you specify here. If you don't specify metric names or provide empty metric names whole metric namespace is excluded. Conflicts with `include_filter`.
* `include_filter` - (Optional) List of inclusive metric filters. If you specify this parameter, the stream sends only the conditional metric names from the metric namespaces that you specify here. If you don't specify metric names or provide empty metric names whole metric namespace is included. Conflicts with `exclude_filter`.
* `name` - (Optional, Forces new resource) Friendly name of the metric stream. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique friendly name beginning with the specified prefix. Conflicts with `name`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `statistics_configuration` - (Optional) For each entry in this array, you specify one or more metrics and the list of additional statistics to stream for those metrics. The additional statistics that you can stream depend on the stream's `output_format`. If the OutputFormat is `json`, you can stream any additional statistic that is supported by CloudWatch, listed in [CloudWatch statistics definitions](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html.html). If the OutputFormat is `opentelemetry0.7`, you can stream percentile statistics (p99 etc.). See details below.
* `include_linked_accounts_metrics` (Optional) If you are creating a metric stream in a monitoring account, specify true to include metrics from source accounts that are linked to this monitoring account, in the metric stream. The default is false. For more information about linking accounts, see [CloudWatch cross-account observability](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-Unified-Cross-Account.html).

### Nested Fields

#### `exclude_filter`

* `namespace` - (Required) Name of the metric namespace in the filter.
* `metric_names` - (Optional) An array that defines the metrics you want to exclude for this metric namespace

#### `include_filter`

* `namespace` - (Required) Name of the metric namespace in the filter.
* `metric_names` - (Optional) An array that defines the metrics you want to include for this metric namespace

#### `statistics_configurations`

* `additional_statistics` - (Required) The additional statistics to stream for the metrics listed in `include_metrics`.
* `include_metric` - (Required) An array that defines the metrics that are to have additional statistics streamed. See details below.

#### `include_metrics`

* `metric_name` - (Required) The name of the metric.
* `namespace` - (Required) The namespace of the metric.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the metric stream.
* `creation_date` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the metric stream was created.
* `last_update_date` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the metric stream was last updated.
* `state` - State of the metric stream. Possible values are `running` and `stopped`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch metric streams using the `name`. For example:

```terraform
import {
  to = aws_cloudwatch_metric_stream.sample
  id = "sample-stream-name"
}
```

Using `terraform import`, import CloudWatch metric streams using the `name`. For example:

```console
% terraform import aws_cloudwatch_metric_stream.sample sample-stream-name
```
