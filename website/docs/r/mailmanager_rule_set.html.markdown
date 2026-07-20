---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_rule_set"
description: |-
  Manages an AWS SES Mail Manager Rule Set.
---

# Resource: aws_mailmanager_rule_set

Manages an AWS SES Mail Manager Rule Set.

## Example Usage

### Basic Usage

```terraform
resource "aws_mailmanager_rule_set" "example" {
  name = "example"

  rule {
    name = "add-header"

    action {
      add_header {
        header_name  = "X-Example"
        header_value = "example"
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the rule set.
* `rule` - (Required) One or more rules that define filtering and action logic. Up to 40 rules are supported. See [Rule](#rule).

The following arguments are optional:

* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Rule

* `action` - (Required) One or more actions to execute when all conditions match. Between 1 and 10 actions are supported. Each action must contain exactly one action configuration. See [Action](#action).
* `condition` - (Optional) One or more conditions that must all evaluate to true for the rule to match. Up to 10 conditions are supported. See [Condition](#condition).
* `name` - (Optional) Name of the rule.
* `unless` - (Optional) One or more conditions that prevent the rule from matching when any evaluates to true. Up to 10 conditions are supported. See [Condition](#condition).

### Condition

Each `condition` and `unless` block must configure exactly one of the following expression blocks:

* `boolean_expression` - (Optional) Boolean expression evaluated against an email attribute or Add On result. See [Boolean Expression](#boolean-expression).
* `dmarc_expression` - (Optional) DMARC policy expression evaluated against the email's DMARC result. See [DMARC Expression](#dmarc-expression).
* `ip_expression` - (Optional) IP CIDR expression evaluated against the sender IP address. See [IP Expression](#ip-expression).
* `number_expression` - (Optional) Numeric expression evaluated against an email attribute such as message size. See [Number Expression](#number-expression).
* `string_expression` - (Optional) String expression evaluated against an email attribute, MIME header, client certificate field, or Add On result. See [String Expression](#string-expression).
* `verdict_expression` - (Optional) Verdict expression evaluated against email authentication results such as SPF or DKIM. See [Verdict Expression](#verdict-expression).

### Boolean Expression

* `evaluate` - (Required) Operand evaluated by the expression. Exactly one of `analysis`, `attribute`, or `is_in_address_list` must be configured.
* `operator` - (Required) Boolean matching operator. Valid values are `IS_TRUE` and `IS_FALSE`.

The `evaluate` block supports:

* `analysis` - (Optional) Add On result to evaluate. See [Analysis](#analysis).
* `attribute` - (Optional) Boolean email attribute to evaluate. Valid values are `READ_RECEIPT_REQUESTED`, `TLS`, and `TLS_WRAPPED`.
* `is_in_address_list` - (Optional) Address-list membership expression.

The `is_in_address_list` block supports:

* `address_lists` - (Required) List containing exactly one address list ARN or identifier.
* `attribute` - (Required) Email attribute to evaluate against the address list. Valid values are `MAIL_FROM`, `HELO`, `RECIPIENT`, `SENDER`, `FROM`, `SUBJECT`, `TO`, and `CC`.

### DMARC Expression

* `operator` - (Required) DMARC policy matching operator. Valid values are `EQUALS` and `NOT_EQUALS`.
* `values` - (Required) List of DMARC policy values. Valid values are `NONE`, `QUARANTINE`, and `REJECT`.

### IP Expression

* `evaluate` - (Required) Left-hand operand of the expression.
* `operator` - (Required) CIDR matching operator. Valid values are `CIDR_MATCHES` and `NOT_CIDR_MATCHES`.
* `values` - (Required) List of IP CIDR ranges against which the sender IP address is evaluated. Between 1 and 10 values are supported.

The `evaluate` block supports:

* `attribute` - (Required) IP attribute to evaluate. Valid value is `SOURCE_IP`.

### Number Expression

* `evaluate` - (Required) Left-hand operand of the expression.
* `operator` - (Required) Numeric comparison operator. Valid values are `EQUALS`, `NOT_EQUALS`, `LESS_THAN`, `GREATER_THAN`, `LESS_THAN_OR_EQUAL`, and `GREATER_THAN_OR_EQUAL`.
* `value` - (Required) Numeric value to compare against.

The `evaluate` block supports:

* `attribute` - (Required) Numeric email attribute to evaluate. Valid value is `MESSAGE_SIZE`.

### String Expression

* `evaluate` - (Required) Left-hand operand of the expression. Exactly one of `analysis`, `attribute`, `client_certificate_attribute`, or `mime_header_attribute` must be configured.
* `operator` - (Required) String matching operator. Valid values are `EQUALS`, `NOT_EQUALS`, `STARTS_WITH`, `ENDS_WITH`, and `CONTAINS`.
* `values` - (Required) List of strings against which the selected operand is evaluated. Between 1 and 10 values are supported, each up to 4096 characters.

The `evaluate` block supports:

* `analysis` - (Optional) Add On result to evaluate. See [Analysis](#analysis).
* `attribute` - (Optional) Email attribute to evaluate. Valid values are `MAIL_FROM`, `HELO`, `RECIPIENT`, `SENDER`, `FROM`, `SUBJECT`, `TO`, and `CC`.
* `client_certificate_attribute` - (Optional) Client certificate field to evaluate. Valid values are `CN`, `SAN_RFC822_NAME`, `SAN_DNS_NAME`, `SAN_DIRECTORY_NAME`, `SAN_UNIFORM_RESOURCE_IDENTIFIER`, `SAN_IP_ADDRESS`, `SAN_REGISTERED_ID`, and `SERIAL_NUMBER`.
* `mime_header_attribute` - (Optional) MIME header name to evaluate. Must contain between 1 and 256 characters and begin with `X-` or `x-`.

### Verdict Expression

* `evaluate` - (Required) Left-hand operand of the expression. Exactly one of `analysis` or `attribute` must be configured.
* `operator` - (Required) Verdict matching operator. Valid values are `EQUALS` and `NOT_EQUALS`.
* `values` - (Required) List of verdict values. Valid values are `PASS`, `FAIL`, `GRAY`, and `PROCESSING_FAILED`. Between 1 and 10 values are supported.

The `evaluate` block supports:

* `analysis` - (Optional) Add On result to evaluate. See [Analysis](#analysis).
* `attribute` - (Optional) Email authentication attribute to evaluate. Valid values are `SPF` and `DKIM`.

### Analysis

* `analyzer` - (Required) ARN of the Mail Manager Add On.
* `result_field` - (Required) Result field returned by the Add On. Must contain between 1 and 256 characters.

### Action

The `action` block supports the following blocks:

* `add_header` - (Optional) Adds a header to the email. See [Add Header](#add-header).
* `archive` - (Optional) Archives the email. See [Archive](#archive).
* `bounce` - (Optional) Sends a bounce response. See [Bounce](#bounce).
* `deliver_to_mailbox` - (Optional) Delivers the email to a WorkMail mailbox. See [Deliver to Mailbox](#deliver-to-mailbox).
* `deliver_to_q_business` - (Optional) Delivers the email to an Amazon Q Business application. See [Deliver to Q Business](#deliver-to-q-business).
* `drop` - (Optional) Stops rule evaluation and drops the email.
* `invoke_lambda` - (Optional) Invokes a Lambda function. See [Invoke Lambda](#invoke-lambda).
* `publish_to_sns` - (Optional) Publishes the email to an SNS topic. See [Publish to SNS](#publish-to-sns).
* `relay` - (Optional) Relays the email to an SMTP server. See [Relay](#relay).
* `replace_recipient` - (Optional) Replaces envelope recipients. See [Replace Recipient](#replace-recipient).
* `send` - (Optional) Sends the email to the internet. See [Send](#send).
* `write_to_s3` - (Optional) Writes the email MIME content to an S3 bucket. See [Write to S3](#write-to-s3).

### Add Header

* `header_name` - (Required) Header name. Must begin with `X-`.
* `header_value` - (Required) Header value.

### Archive

* `target_archive` - (Required) Identifier of the archive.
* `action_failure_policy` - (Optional) Policy applied when the action fails.

### Bounce

* `diagnostic_message` - (Required) Diagnostic message included in the bounce.
* `role_arn` - (Required) ARN of the IAM role used to send the bounce.
* `sender` - (Required) Sender address of the bounce.
* `smtp_reply_code` - (Required) SMTP reply code.
* `status_code` - (Required) Enhanced status code.
* `action_failure_policy` - (Optional) Policy applied when the action fails.
* `message` - (Optional) Human-readable bounce message.

### Deliver to Mailbox

* `mailbox_arn` - (Required) ARN of the WorkMail organization.
* `role_arn` - (Required) ARN of the IAM role used to deliver the email.
* `action_failure_policy` - (Optional) Policy applied when the action fails.

### Deliver to Q Business

* `application_id` - (Required) Q Business application identifier.
* `index_id` - (Required) Q Business index identifier.
* `role_arn` - (Required) ARN of the IAM role used to deliver the email.
* `action_failure_policy` - (Optional) Policy applied when the action fails.

### Invoke Lambda

* `function_arn` - (Required) ARN of the Lambda function.
* `invocation_type` - (Required) Lambda invocation type.
* `role_arn` - (Required) ARN of the IAM role used to invoke the function.
* `action_failure_policy` - (Optional) Policy applied when the action fails.
* `retry_time_minutes` - (Optional) Maximum retry time in minutes.

### Publish to SNS

* `role_arn` - (Required) ARN of the IAM role used to publish the email.
* `topic_arn` - (Required) ARN of the SNS topic.
* `action_failure_policy` - (Optional) Policy applied when the action fails.
* `encoding` - (Optional) Email encoding in the notification.
* `payload_type` - (Optional) Notification payload type.

### Relay

* `relay` - (Required) Identifier of the relay resource.
* `action_failure_policy` - (Optional) Policy applied when the action fails.
* `mail_from` - (Optional) Whether to preserve or replace the original MAIL FROM address.

### Replace Recipient

* `replace_with` - (Required) Replacement envelope recipient addresses.

### Send

* `role_arn` - (Required) ARN of the IAM role used to send the email.
* `action_failure_policy` - (Optional) Policy applied when the action fails.

### Write to S3

* `role_arn` - (Required) ARN of the IAM role used to write to S3.
* `s3_bucket` - (Required) Name of the S3 bucket.
* `action_failure_policy` - (Optional) Policy applied when the action fails.
* `s3_prefix` - (Optional) S3 object key prefix.
* `s3_sse_kms_key_id` - (Optional) KMS key identifier used to encrypt the email.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the rule set.
* `created_date` - Date and time when the rule set was created.
* `id` - Identifier of the rule set.
* `last_modification_date` - Date and time when the rule set was last modified.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_mailmanager_rule_set.example
  identity = {
    id = "rule-set-id"
  }
}
```

### Identity Schema

#### Required

* `id` (String) Identifier of the rule set.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an SES Mail Manager Rule Set using its identifier. For example:

```terraform
import {
  to = aws_mailmanager_rule_set.example
  id = "rule-set-id"
}
```

Using `terraform import`, import an SES Mail Manager Rule Set using its identifier. For example:

```console
% terraform import aws_mailmanager_rule_set.example rule-set-id
```
