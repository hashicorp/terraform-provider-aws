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

* `data_source` - (Required, Forces new resource) Data source to associate with the S3 Table Integration. See [`data_source` Block](#data_source-block) below.
* `integration_arn` - (Required, Forces new resource) ARN of the `aws_observabilityadmin_s3_table_integration` to associate the data source with.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `data_source` Block

The `data_source` block supports the following arguments:

* `name` - (Required, Forces new resource) Name of the data source. Use `"*"` to match all sources.
* `type` - (Required, Forces new resource) Type of the data source. Use `"*"` to match all types.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier (UUID) for the association between the data source and S3 Table Integration. To find this value for an existing association created outside Terraform, run `aws logs list-sources-for-s3-table-integration --integration-arn <ARN>` and use the `Identifier` field from the response.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_log_s3_table_integration_source.example
  identity = {
    integration_arn = "arn:aws:observabilityadmin:us-west-2:123456789012:s3tableintegration/3g5043wqe54nmw5apiugwkn1a"
    id              = "a8928b36-ab82-4ae2-ae5c-fcb910ec4237"
  }
}

resource "aws_cloudwatch_log_s3_table_integration_source" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `integration_arn` (String) ARN of the integration.
* `id` (String) ID of the association.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Table Integration data source associations using `integration_arn` and `id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_cloudwatch_log_s3_table_integration_source.example
  id = "arn:aws:observabilityadmin:us-west-2:123456789012:s3tableintegration/3g5043wqe54nmw5apiugwkn1a,a8928b36-ab82-4ae2-ae5c-fcb910ec4237"
}
```

Using `terraform import`, import S3 Table Integration data source associations using `integration_arn` and `id` separated by a comma (`,`). For example:

```console
% terraform import aws_cloudwatch_log_s3_table_integration_source.example arn:aws:observabilityadmin:us-west-2:123456789012:s3tableintegration/3g5043wqe54nmw5apiugwkn1a,a8928b36-ab82-4ae2-ae5c-fcb910ec4237
```
