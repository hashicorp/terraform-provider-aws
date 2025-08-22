---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_logging_options"
description: |-
    Provides a resource to manage default logging options.
---

# Resource: aws_iot_logging_options

Provides a resource to manage [default logging options](https://docs.aws.amazon.com/iot/latest/developerguide/configure-logging.html#configure-logging-console).

## Example Usage

```terraform
resource "aws_iot_logging_options" "example" {
  default_log_level = "WARN"
  role_arn          = aws_iam_role.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `default_log_level` - (Optional) The default logging level. Valid Values: `"DEBUG"`, `"INFO"`, `"ERROR"`, `"WARN"`, `"DISABLED"`.
* `disable_all_logs` - (Optional) If `true` all logs are disabled. The default is `false`.
* `role_arn` - (Required) The ARN of the role that allows IoT to write to Cloudwatch logs.

## Attribute Reference

This resource exports no additional attributes.
