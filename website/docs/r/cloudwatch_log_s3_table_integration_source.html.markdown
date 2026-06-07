---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_s3_table_integration_source"
description: |-
  Manages a CloudWatch Logs S3 Table Integration data source association.
---

# Resource: aws_cloudwatch_log_s3_table_integration_source

Manages a CloudWatch Logs S3 Table Integration data source association.
This resource associates a data source (such as a CloudWatch log type) with an S3 Table Integration, enabling CloudWatch logs to be automatically written to S3 Tables for analytics.

For more information, see the [CloudWatch Logs S3 Tables integration documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/s3-tables-integration.html).

## Example Usage

### Associate All Sources (Wildcard)

```terraform
resource "aws_cloudwatch_log_s3_table_integration_source" "example" {
  integration_arn = aws_observabilityadmin_s3_table_integration.example.arn

  data_source {
    name = "*"
    type = "*"
  }
}
```

### Associate a Custom Data Source

To route log stream messages from a specific custom data source into a dedicated S3 Table, tag the CloudWatch log group with `cw:datasource:name` and `cw:datasource:type`. The integration then writes matching log streams into a table named `{name}__{type}` inside the `aws-cloudwatch` table bucket.

```terraform
# Tag the log group to declare it as a custom data source.
# Log stream messages are written to the "myapp__events" table
# in the aws-cloudwatch table bucket.
resource "aws_cloudwatch_log_group" "example" {
  name = "/example/myapp"

  tags = {
    "cw:datasource:name" = "myapp"
    "cw:datasource:type" = "events"
  }
}

resource "aws_cloudwatch_log_s3_table_integration_source" "example" {
  integration_arn = aws_observabilityadmin_s3_table_integration.example.arn

  data_source {
    name = "myapp"
    type = "events"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `data_source` - (Required, Forces new resource) Data source to associate with the S3 Table Integration. See below.
* `integration_arn` - (Required, Forces new resource) ARN of the `aws_observabilityadmin_s3_table_integration` to associate the data source with.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

The `data_source` block supports the following arguments:

* `name` - (Required, Forces new resource) Name of the data source.
* `type` - (Required, Forces new resource) Type of the data source.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - unique identifier for the association between the data source and S3 Table Integration.
