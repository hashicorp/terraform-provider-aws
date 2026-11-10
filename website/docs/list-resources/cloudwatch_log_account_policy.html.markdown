---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_account_policy"
description: |-
  Lists CloudWatch Logs Account Policy resources.
---

# List Resource: aws_cloudwatch_log_account_policy

Lists CloudWatch Logs Account Policy resources.

## Example Usage

```terraform
list "aws_cloudwatch_log_account_policy" "example" {
  provider = aws

  config {
    policy_type = "RETENTION_POLICY"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `policy_type` - (Required) Type of account policy to list. Valid values are `DATA_PROTECTION_POLICY`, `SUBSCRIPTION_FILTER_POLICY`, `FIELD_INDEX_POLICY`, `TRANSFORMER_POLICY`, `METRIC_EXTRACTION_POLICY`, and `RETENTION_POLICY`.
* `region` - (Optional) Region to query. Defaults to provider region.
