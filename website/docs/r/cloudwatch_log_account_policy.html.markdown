---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_account_policy"
description: |-
  Provides a CloudWatch Log Account Policy resource.
---

# Resource: aws_cloudwatch_log_account_policy

Provides a CloudWatch Log Account Policy resource.

## Example Usage

### Account Data Protection Policy

```terraform
resource "aws_cloudwatch_log_account_policy" "data_protection" {
  policy_name = "data-protection"
  policy_type = "DATA_PROTECTION_POLICY"
  policy_document = jsonencode({
    Name    = "DataProtection"
    Version = "2021-06-01"

    Statement = [
      {
        Sid            = "Audit"
        DataIdentifier = ["arn:aws:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Audit = {
            FindingsDestination = {}
          }
        }
      },
      {
        Sid            = "Redact"
        DataIdentifier = ["arn:aws:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Deidentify = {
            MaskConfig = {}
          }
        }
      }
    ]
  })
}
```

### Subscription Filter Policy

```terraform
resource "aws_cloudwatch_log_account_policy" "subscription_filter" {
  policy_name = "subscription-filter"
  policy_type = "SUBSCRIPTION_FILTER_POLICY"
  policy_document = jsonencode(
    {
      DestinationArn = aws_lambda_function.test.arn
      FilterPattern  = "test"
    }
  )
  selection_criteria = "LogGroupName NOT IN [\"excluded_log_group_name\"]"
}
```

### Field Index Policy

```terraform
resource "aws_cloudwatch_log_account_policy" "field_index" {
  policy_name = "field-index"
  policy_type = "FIELD_INDEX_POLICY"
  policy_document = jsonencode(
    {
      Fields = [
        "field1",
        "field2"
      ]
    }
  )
}
```

### Metric Extraction Policy

```terraform
resource "aws_cloudwatch_log_account_policy" "metric_extraction" {
  policy_name = "disable-emf-metrics"
  policy_type = "METRIC_EXTRACTION_POLICY"
  policy_document = jsonencode(
    {
      EmbeddedMetricFormat = {
        Status = "Disabled"
      }
    }
  )
  selection_criteria = "LogGroupNamePrefix NOT IN [\"/aws/containerinsights\", \"/aws/ecs/containerinsights\", \"/aws/application-signals/data\"]"
}
```

### Retention Policy

```terraform
resource "aws_cloudwatch_log_account_policy" "retention" {
  policy_name = "lambda-retention"
  policy_type = "RETENTION_POLICY"
  policy_document = jsonencode(
    {
      retentionInDays = 90
      enforcementMode = "RETROACTIVE"
    }
  )
  selection_criteria = "LogGroupNamePrefix = \"/aws/lambda\""
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `policy_document` - (Required) Text of the account policy. Refer to the [AWS docs](https://docs.aws.amazon.com/cli/latest/reference/logs/put-account-policy.html) for more information.
* `policy_type` - (Required) Type of account policy. Valid values are `DATA_PROTECTION_POLICY`, `SUBSCRIPTION_FILTER_POLICY`, `FIELD_INDEX_POLICY`, `TRANSFORMER_POLICY`, `METRIC_EXTRACTION_POLICY`, and `RETENTION_POLICY`. Refer to the [AWS docs](https://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_PutAccountPolicy.html) for the number of policies of each type allowed in an account.
* `policy_name` - (Required) Name of the account policy.
* `scope` - (Optional) Currently defaults to and only accepts the value: `ALL`.
* `selection_criteria` - (Optional) Criteria for applying the policy to a selection of log groups. Valid only when `policy_type` is `SUBSCRIPTION_FILTER_POLICY`, `FIELD_INDEX_POLICY`, `TRANSFORMER_POLICY`, `METRIC_EXTRACTION_POLICY`, or `RETENTION_POLICY`. For `SUBSCRIPTION_FILTER_POLICY`, the only allowable criteria selector is `LogGroupName NOT IN []`. For `RETENTION_POLICY`, the only allowable criteria selector is `LogGroupNamePrefix`, for example `LogGroupNamePrefix = "/aws/lambda"`. For `METRIC_EXTRACTION_POLICY`, `LogGroupName` and `LogGroupNamePrefix` can be used with `IN` and `NOT IN`, for example `LogGroupNamePrefix NOT IN ["/aws/containerinsights"]`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_log_account_policy.example
  identity = {
    policy_name = "my-account-policy"
    policy_type = "SUBSCRIPTION_FILTER_POLICY"
  }
}

resource "aws_cloudwatch_log_account_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `policy_name` (String) Name of the account policy.
* `policy_type` (String) Type of account policy.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Account Policies using `policy_name` and `policy_type` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_cloudwatch_log_account_policy.example
  id = "my-account-policy:SUBSCRIPTION_FILTER_POLICY"
}
```

Using `terraform import`, import Account Policies using `policy_name` and `policy_type` separated by a colon (`:`). For example:

```console
% terraform import aws_cloudwatch_log_account_policy.example "my-account-policy:SUBSCRIPTION_FILTER_POLICY"
```
