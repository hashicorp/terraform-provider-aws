---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_remediation_configuration"
description: |-
  Provides an AWS Config Remediation Configuration.
---

# Resource: aws_config_remediation_configuration

Provides an AWS Config Remediation Configuration.

~> **Note:** Config Remediation Configuration requires an existing [Config Rule](/docs/providers/aws/r/config_config_rule.html) to be present.

## Example Usage

AWS managed rules can be used by setting the source owner to `AWS` and the source identifier to the name of the managed rule. More information about AWS managed rules can be found in the [AWS Config Developer Guide](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html).

```terraform
resource "aws_config_config_rule" "this" {
  name = "example"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }
}

resource "aws_config_remediation_configuration" "this" {
  config_rule_name = aws_config_config_rule.this.name
  resource_type    = "AWS::S3::Bucket"
  target_type      = "SSM_DOCUMENT"
  target_id        = "AWS-EnableS3BucketEncryption"
  target_version   = "1"

  parameter {
    name         = "AutomationAssumeRole"
    static_value = "arn:aws:iam::875924563244:role/security_config"
  }
  parameter {
    name           = "BucketName"
    resource_value = "RESOURCE_ID"
  }
  parameter {
    name         = "SSEAlgorithm"
    static_value = "AES256"
  }

  automatic                  = true
  maximum_automatic_attempts = 10
  retry_attempt_seconds      = 600

  execution_controls {
    ssm_controls {
      concurrent_execution_rate_percentage = 25
      error_percentage                     = 20
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `config_rule_name` - (Required) Name of the AWS Config rule.
* `target_id` - (Required) Target ID is the name of the public document.
* `target_type` - (Required) Type of the target. Target executes remediation. For example, SSM document.

The following arguments are optional:

* `automatic` - (Optional) Remediation is triggered automatically if `true`.
* `execution_controls` - (Optional) Configuration block for execution controls. See below.
* `maximum_automatic_attempts` - (Optional) Maximum number of failed attempts for auto-remediation. If you do not select a number, the default is 5.
* `parameter` - (Optional) Can be specified multiple times for each parameter. Each parameter block supports arguments below.
* `resource_type` - (Optional) Type of resource.
* `retry_attempt_seconds` - (Optional) Maximum time in seconds that AWS Config runs auto-remediation. If you do not select a number, the default is 60 seconds.
* `target_version` - (Optional) Version of the target. For example, version of the SSM document

### `execution_controls`

* `ssm_controls` - (Required) Configuration block for SSM controls. See below.

#### `ssm_controls`

One or both of these values are required.

* `concurrent_execution_rate_percentage` - (Optional) Maximum percentage of remediation actions allowed to run in parallel on the non-compliant resources for that specific rule. The default value is 10%.
* `error_percentage` - (Optional) Percentage of errors that are allowed before SSM stops running automations on non-compliant resources for that specific rule. The default is 50%.

### `parameter`

The value is either a dynamic (resource) value or a static value. You must select either a dynamic value or a static value.

* `name` - (Required) Name of the attribute.
* `resource_value` - (Optional) Value is dynamic and changes at run-time.
* `static_value` - (Optional) Value is static and does not change at run-time.
* `static_values` - (Optional) List of static values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Config Remediation Configuration.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Remediation Configurations using the name config_rule_name. For example:

```terraform
import {
  to = aws_config_remediation_configuration.this
  id = "example"
}
```

Using `terraform import`, import Remediation Configurations using the name config_rule_name. For example:

```console
% terraform import aws_config_remediation_configuration.this example
```
