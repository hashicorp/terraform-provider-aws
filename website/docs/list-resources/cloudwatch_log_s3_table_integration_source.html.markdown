---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_logs_s3_table_integration_source"
description: |-
  Lists CloudWatch Logs S3 Table Integration Source resources.
---

# List Resource: aws_cloudwatch_log_s3_table_integration_source

Lists CloudWatch Logs S3 Table Integration Source resources.

## Example Usage

```terraform
list "aws_cloudwatch_log_s3_table_integration_source" "example" {
  provider = aws

  config {
    integration_arn = aws_observabilityadmin_s3_table_integration.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `integration_arn` - (Required) ARN of the integration.
* `region` - (Optional) Region to query. Defaults to provider region.
