---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_s3_table_source_association"
description: |-
  Manages a CloudWatch Logs S3 Table Source Association.
---

# Resource: aws_cloudwatch_log_s3_table_source_association

Manages a CloudWatch Logs S3 Table Source Association. This resource associates a data source (such as a CloudWatch log type) with an S3 Table Integration, enabling CloudWatch logs to be automatically written to S3 Tables for analytics.

For more information, see the [CloudWatch Logs S3 Tables integration documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/s3-tables-integration.html).

## Example Usage

### Associate All Sources (Wildcard)

```terraform
resource "aws_cloudwatch_log_s3_table_source_association" "example" {
  integration_arn = aws_observabilityadmin_s3_table_integration.example.arn
  # datasource_name and datasource_type default to "*" (all sources)
}
```

### Associate a Custom Data Source

To route log stream messages from a specific custom data source into a dedicated S3 Table, tag the CloudWatch log group with `cw:datasource:name` and `cw:datasource:type`. The integration then writes matching log streams into a table named `{name}__{type}` inside the `aws-cloudwatch` table bucket.

```terraform
resource "aws_iam_role" "example" {
  name = "example-s3-table-integration"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "logs.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "example" {
  role = aws_iam_role.example.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3tables:CreateTableBucket",
          "s3tables:ListTableBuckets",
          "s3tables:GetTableBucket",
          "s3tables:CreateNamespace",
          "s3tables:GetNamespace",
          "s3tables:ListNamespaces",
          "s3tables:CreateTable",
          "s3tables:GetTable",
          "s3tables:ListTables",
          "s3tables:PutTableData",
          "s3tables:GetTableData",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_observabilityadmin_s3_table_integration" "example" {
  role_arn = aws_iam_role.example.arn

  encryption {
    sse_algorithm = "AES256"
  }
}

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

resource "aws_cloudwatch_log_s3_table_source_association" "example" {
  integration_arn = aws_observabilityadmin_s3_table_integration.example.arn
  datasource_name = "myapp"
  datasource_type = "events"
}
```

## Argument Reference

This resource supports the following arguments:

* `datasource_name` - (Optional, Forces new resource) Name of the data source. Defaults to `*` to match all data source names. When specifying a custom value, only lowercase letters, numbers, and underscores are allowed (no uppercase, hyphens, or spaces). Must match the value of the `cw:datasource:name` tag on the source log group.
* `datasource_type` - (Optional, Forces new resource) Type of the data source. Defaults to `*` to match all data source types. When specifying a custom value, only lowercase letters, numbers, and underscores are allowed (no uppercase, hyphens, or spaces). Must match the value of the `cw:datasource:type` tag on the source log group.
* `integration_arn` - (Required, Forces new resource) ARN of the `aws_observabilityadmin_s3_table_integration` to associate the data source with.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for this source association.
* `status` - Current status of the association. Valid values: `ACTIVE`, `UNHEALTHY`, `FAILED`, `DATA_SOURCE_DELETE_IN_PROGRESS`.
* `status_reason` - Additional information about the current status.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs S3 Table Source Associations using `integration_arn,id`. For example:

```terraform
import {
  to = aws_cloudwatch_log_s3_table_source_association.example
  id = "arn:aws:observabilityadmin:us-east-1:123456789012:s3-table-integration/example-id,source-identifier"
}
```

Using `terraform import`, import CloudWatch Logs S3 Table Source Associations using `integration_arn,id`. For example:

```console
% terraform import aws_cloudwatch_log_s3_table_source_association.example arn:aws:observabilityadmin:us-east-1:123456789012:s3-table-integration/example-id,source-identifier
```
