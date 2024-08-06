---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl_logging_configuration"
description: |-
  Create a resource for WAFv2 Web ACL Logging Configuration.
---

# Resource: aws_wafv2_web_acl_logging_configuration

This resource creates a WAFv2 Web ACL Logging Configuration.

!> **WARNING:** When logging from a WAFv2 Web ACL to a CloudWatch Log Group, the WAFv2 service tries to create or update a generic Log Resource Policy named `AWSWAF-LOGS`. However, if there are a large number of Web ACLs or if the account frequently creates and deletes Web ACLs, this policy may exceed the maximum policy size. As a result, this resource type will fail to be created. More details about this issue can be found in [this issue](https://github.com/hashicorp/terraform-provider-aws/issues/25296). To prevent this issue, you can manage a specific resource policy. Please refer to the [example](#with-cloudwatch-log-group-and-managed-cloudwatch-log-resource-policy) below for managing a CloudWatch Log Group with a managed CloudWatch Log Resource Policy.

## Example Usage

### With Redacted Fields

```terraform
resource "aws_wafv2_web_acl_logging_configuration" "example" {
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.example.arn]
  resource_arn            = aws_wafv2_web_acl.example.arn
  redacted_fields {
    single_header {
      name = "user-agent"
    }
  }
}
```

### With Logging Filter

```terraform
resource "aws_wafv2_web_acl_logging_configuration" "example" {
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.example.arn]
  resource_arn            = aws_wafv2_web_acl.example.arn

  logging_filter {
    default_behavior = "KEEP"

    filter {
      behavior = "DROP"

      condition {
        action_condition {
          action = "COUNT"
        }
      }

      condition {
        label_name_condition {
          label_name = "awswaf:111122223333:rulegroup:testRules:LabelNameZ"
        }
      }

      requirement = "MEETS_ALL"
    }

    filter {
      behavior = "KEEP"

      condition {
        action_condition {
          action = "ALLOW"
        }
      }

      requirement = "MEETS_ANY"
    }
  }
}
```

### With CloudWatch Log Group and managed CloudWatch Log Resource Policy

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "aws-waf-logs-some-uniq-suffix"
}

resource "aws_wafv2_web_acl_logging_configuration" "example" {
  log_destination_configs = [aws_cloudwatch_log_group.example.arn]
  resource_arn            = aws_wafv2_web_acl.example.arn
}

resource "aws_cloudwatch_log_resource_policy" "example" {
  policy_document = data.aws_iam_policy_document.example.json
  policy_name     = "webacl-policy-uniq-name"
}

data "aws_iam_policy_document" "example" {
  version = "2012-10-17"
  statement {
    effect = "Allow"
    principals {
      identifiers = ["delivery.logs.amazonaws.com"]
      type        = "Service"
    }
    actions   = ["logs:CreateLogStream", "logs:PutLogEvents"]
    resources = ["${aws_cloudwatch_log_group.example.arn}:*"]
    condition {
      test     = "ArnLike"
      values   = ["arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"]
      variable = "aws:SourceArn"
    }
    condition {
      test     = "StringEquals"
      values   = [tostring(data.aws_caller_identity.current.account_id)]
      variable = "aws:SourceAccount"
    }
  }
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}
```

## Argument Reference

This resource supports the following arguments:

* `log_destination_configs` - (Required) Configuration block that allows you to associate Amazon Kinesis Data Firehose, Cloudwatch Log log group, or S3 bucket Amazon Resource Names (ARNs) with the web ACL. **Note:** data firehose, log group, or bucket name **must** be prefixed with `aws-waf-logs-`, e.g. `aws-waf-logs-example-firehose`, `aws-waf-logs-example-log-group`, or `aws-waf-logs-example-bucket`.
* `logging_filter` - (Optional) Configuration block that specifies which web requests are kept in the logs and which are dropped. It allows filtering based on the rule action and the web request labels applied by matching rules during web ACL evaluation. For more details, refer to the [Logging Filter](#logging-filter) section below.
* `redacted_fields` - (Optional) Configuration for parts of the request that you want to keep out of the logs. Up to 100 `redacted_fields` blocks are supported. See [Redacted Fields](#redacted-fields) below for more details.
* `resource_arn` - (Required) Amazon Resource Name (ARN) of the web ACL that you want to associate with `log_destination_configs`.

### Logging Filter

The `logging_filter` block supports the following arguments:

* `default_behavior` - (Required) Default handling for logs that don't match any of the specified filtering conditions. Valid values for `default_behavior` are `KEEP` or `DROP`.
* `filter` - (Required) Filter(s) that you want to apply to the logs. See [Filter](#filter) below for more details.

### Filter

The `filter` block supports the following arguments:

* `behavior` - (Required) Parameter that determines how to handle logs that meet the conditions and requirements of the filter. The valid values for `behavior` are `KEEP` or `DROP`.
* `condition` - (Required) Match condition(s) for the filter. See [Condition](#condition) below for more details.
* `requirement` - (Required) Logic to apply to the filtering conditions. You can specify that a log must match all conditions or at least one condition in order to satisfy the filter. Valid values for `requirement` are `MEETS_ALL` or `MEETS_ANY`.

### Condition

The `condition` block supports the following arguments:

~> **NOTE:** Either the `action_condition` or `label_name_condition` must be specified.

* `action_condition` - (Optional) Configuration for a single action condition. See [Action Condition](#action-condition) below for more details.
* `label_name_condition` - (Optional) Condition for a single label name. See [Label Name Condition](#label-name-condition) below for more details.

### Action Condition

The `action_condition` block supports the following argument:

* `action` - (Required) Action setting that a log record must contain in order to meet the condition. Valid values for `action` are `ALLOW`, `BLOCK`, and `COUNT`.

### Label Name Condition

The `label_name_condition` block supports the following argument:

* `label_name` - (Required) Name of the label that a log record must contain in order to meet the condition. It must be a fully qualified label name, which includes a prefix, optional namespaces, and the label name itself. The prefix identifies the rule group or web ACL context of the rule that added the label.

### Redacted Fields

The `redacted_fields` block supports the following arguments:

~> **NOTE:** You can only specify one of the following: `method`, `query_string`, `single_header`, or `uri_path`.

* `method` - (Optional) HTTP method to be redacted. It must be specified as an empty configuration block `{}`. The method indicates the type of operation that the request is asking the origin to perform.
* `query_string` - (Optional) Whether to redact the query string. It must be specified as an empty configuration block `{}`. The query string is the part of a URL that appears after a `?` character, if any.
* `single_header` - (Optional) "single_header" refers to the redaction of a single header. For more information, please see the details below under [Single Header](#single-header).
* `uri_path` - (Optional) Configuration block that redacts the request URI path. It should be specified as an empty configuration block `{}`. The URI path is the part of a web request that identifies a resource, such as `/images/daily-ad.jpg`.

### Single Header

To redact a single header, provide the name of the header to be redacted. For example, use `User-Agent` or `Referer` (provided as lowercase strings).

The `single_header` block supports the following arguments:

* `name` - (Required) Name of the query header to redact. This setting must be provided in lowercase characters.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the WAFv2 Web ACL.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 Web ACL Logging Configurations using the ARN of the WAFv2 Web ACL. For example:

```terraform
import {
  to = aws_wafv2_web_acl_logging_configuration.example
  id = "arn:aws:wafv2:us-west-2:123456789012:regional/webacl/test-logs/a1b2c3d4-5678-90ab-cdef"
}
```

Using `terraform import`, import WAFv2 Web ACL Logging Configurations using the ARN of the WAFv2 Web ACL. For example:

```console
% terraform import aws_wafv2_web_acl_logging_configuration.example arn:aws:wafv2:us-west-2:123456789012:regional/webacl/test-logs/a1b2c3d4-5678-90ab-cdef
```
