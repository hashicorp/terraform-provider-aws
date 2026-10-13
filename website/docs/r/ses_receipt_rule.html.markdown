---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_receipt_rule"
description: |-
  Provides an SES receipt rule resource
---

# Resource: aws_ses_receipt_rule

Provides an SES receipt rule resource

## Example Usage

```terraform
# Add a header to the email and store it in S3
resource "aws_ses_receipt_rule" "store" {
  name          = "store"
  rule_set_name = "default-rule-set"
  recipients    = ["karen@example.com"]
  enabled       = true
  scan_enabled  = true

  add_header_action {
    header_name  = "Custom-Header"
    header_value = "Added by SES"
    position     = 1
  }

  s3_action {
    bucket_name = "emails"
    position    = 2
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the rule.
* `rule_set_name` - (Required) Name of the rule set.

The following arguments are optional:

* `add_header_action` - (Optional) Configuration block for adding a header to received emails. Detailed below.
* `after` - (Optional) Name of the rule to place this rule after.
* `bounce_action` - (Optional) Configuration block for rejecting received emails. Detailed below.
* `enabled` - (Optional) If true, the rule will be enabled.
* `lambda_action` - (Optional) Configuration block for calling an AWS Lambda function. Detailed below.
* `recipients` - (Optional) List of email addresses.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `s3_action` - (Optional) Configuration block for storing received emails in an S3 bucket. Detailed below.
* `scan_enabled` - (Optional) If true, incoming emails will be scanned for spam and viruses.
* `sns_action` - (Optional) Configuration block for publishing to an SNS topic. Detailed below.
* `stop_action` - (Optional) Configuration block for terminating the evaluation of the receipt rule set. Detailed below.
* `tls_policy` - (Optional) `Require` or `Optional`.
* `workmail_action` - (Optional) Configuration block for calling Amazon WorkMail. Detailed below.

### `add_header_action` Block

* `header_name` - (Required) Name of the header to add.
* `header_value` - (Required) Value of the header to add.
* `position` - (Required) Position of the action in the receipt rule.

### `bounce_action` Block

* `message` - (Required) Message to send.
* `position` - (Required) Position of the action in the receipt rule.
* `sender` - (Required) Email address of the sender.
* `smtp_reply_code` - (Required) RFC 5321 SMTP reply code.
* `status_code` - (Optional) RFC 3463 SMTP enhanced status code.
* `topic_arn` - (Optional) ARN of an SNS topic to notify.

### `lambda_action` Block

* `function_arn` - (Required) ARN of the Lambda function to invoke.
* `invocation_type` - (Optional) `Event` or `RequestResponse`.
* `position` - (Required) Position of the action in the receipt rule.
* `topic_arn` - (Optional) ARN of an SNS topic to notify.

### `s3_action` Block

* `bucket_name` - (Required) Name of the S3 bucket.
* `iam_role_arn` - (Optional) ARN of the IAM role to be used by Amazon Simple Email Service while writing to the Amazon S3 bucket, optionally encrypting your mail via the provided customer managed key, and publishing to the Amazon SNS topic.
* `kms_key_arn` - (Optional) ARN of the KMS key.
* `object_key_prefix` - (Optional) Key prefix of the S3 bucket.
* `position` - (Required) Position of the action in the receipt rule.
* `topic_arn` - (Optional) ARN of an SNS topic to notify.

### `sns_action` Block

* `encoding` - (Optional) Encoding to use for the email within the Amazon SNS notification. Default value is `UTF-8`.
* `position` - (Required) Position of the action in the receipt rule.
* `topic_arn` - (Required) ARN of an SNS topic to notify.

### `stop_action` Block

* `position` - (Required) Position of the action in the receipt rule.
* `scope` - (Required) Scope to apply. The only acceptable value is `RuleSet`.
* `topic_arn` - (Optional) ARN of an SNS topic to notify.

### `workmail_action` Block

* `organization_arn` - (Required) ARN of the WorkMail organization.
* `position` - (Required) Position of the action in the receipt rule.
* `topic_arn` - (Optional) ARN of an SNS topic to notify.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - SES receipt rule ARN.
* `id` - SES receipt rule name.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SES receipt rules using the ruleset name and rule name separated by `:`. For example:

```terraform
import {
  to = aws_ses_receipt_rule.my_rule
  id = "my_rule_set:my_rule"
}
```

Using `terraform import`, import SES receipt rules using the ruleset name and rule name separated by `:`. For example:

```console
% terraform import aws_ses_receipt_rule.my_rule my_rule_set:my_rule
```
