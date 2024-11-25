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

This resource supports the following arguments:

* `name` - (Required) The name of the rule
* `rule_set_name` - (Required) The name of the rule set
* `after` - (Optional) The name of the rule to place this rule after
* `enabled` - (Optional) If true, the rule will be enabled
* `recipients` - (Optional) A list of email addresses
* `scan_enabled` - (Optional) If true, incoming emails will be scanned for spam and viruses
* `tls_policy` - (Optional) `Require` or `Optional`
* `add_header_action` - (Optional) A list of Add Header Action blocks. Documented below.
* `bounce_action` - (Optional) A list of Bounce Action blocks. Documented below.
* `lambda_action` - (Optional) A list of Lambda Action blocks. Documented below.
* `s3_action` - (Optional) A list of S3 Action blocks. Documented below.
* `sns_action` - (Optional) A list of SNS Action blocks. Documented below.
* `stop_action` - (Optional) A list of Stop Action blocks. Documented below.
* `workmail_action` - (Optional) A list of WorkMail Action blocks. Documented below.

Add header actions support the following:

* `header_name` - (Required) The name of the header to add
* `header_value` - (Required) The value of the header to add
* `position` - (Required) The position of the action in the receipt rule

Bounce actions support the following:

* `message` - (Required) The message to send
* `sender` - (Required) The email address of the sender
* `smtp_reply_code` - (Required) The RFC 5321 SMTP reply code
* `status_code` - (Optional) The RFC 3463 SMTP enhanced status code
* `topic_arn` - (Optional) The ARN of an SNS topic to notify
* `position` - (Required) The position of the action in the receipt rule

Lambda actions support the following:

* `function_arn` - (Required) The ARN of the Lambda function to invoke
* `invocation_type` - (Optional) `Event` or `RequestResponse`
* `topic_arn` - (Optional) The ARN of an SNS topic to notify
* `position` - (Required) The position of the action in the receipt rule

S3 actions support the following:

* `bucket_name` - (Required) The name of the S3 bucket
* `iam_role_arn` - (Optional) The ARN of the IAM role to be used by Amazon Simple Email Service while writing to the Amazon S3 bucket, optionally encrypting your mail via the provided customer managed key, and publishing to the Amazon SNS topic
* `kms_key_arn` - (Optional) The ARN of the KMS key
* `object_key_prefix` - (Optional) The key prefix of the S3 bucket
* `topic_arn` - (Optional) The ARN of an SNS topic to notify
* `position` - (Required) The position of the action in the receipt rule

SNS actions support the following:

* `topic_arn` - (Required) The ARN of an SNS topic to notify
* `position` - (Required) The position of the action in the receipt rule
* `encoding` - (Optional) The encoding to use for the email within the Amazon SNS notification. Default value is `UTF-8`.

Stop actions support the following:

* `scope` - (Required) The scope to apply. The only acceptable value is `RuleSet`.
* `topic_arn` - (Optional) The ARN of an SNS topic to notify
* `position` - (Required) The position of the action in the receipt rule

WorkMail actions support the following:

* `organization_arn` - (Required) The ARN of the WorkMail organization
* `topic_arn` - (Optional) The ARN of an SNS topic to notify
* `position` - (Required) The position of the action in the receipt rule

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The SES receipt rule name.
* `arn` - The SES receipt rule ARN.

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
