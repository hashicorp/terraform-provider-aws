---
layout: "aws"
page_title: "AWS: aws_ssm_parameter"
sidebar_current: "docs-aws-datasource-ssm-parameter"
description: |-
  Provides a SSM Parameter datasource
---

# Data Source: aws_ssm_parameter

Provides an SSM Parameter data source.

## Example Usage

To store a basic string parameter:

```hcl
data "aws_ssm_parameter" "foo" {
  name  = "foo"
}
```
To store a StringList parameter:

```hcl
resource "aws_ssm_parameter" "subnets" {
  name      = "/my/subnets"
  type      = "StringList"
  overwrite = true
  value     = "1.1.1.1,2.2.2.2"
}

data "aws_ssm_parameter" "subnets" {
  name = "/my/subnets"
}

output "subnets_list" {
  value = "${data.aws_ssm_parameter.subnets.value_list}"
}
```

~> **Note:** The unencrypted value of a SecureString will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the parameter.
* `with_decryption` - (Optional) Whether to return decrypted `SecureString` value. Defaults to `true`.


In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the parameter.
* `name` - (Required) The name of the parameter.
* `type` - (Required) The type of the parameter. Valid types are `String`, `StringList` and `SecureString`.
* `value` - (Required) The value of the parameter.
* `value_list` - (Optional) It will provide value of the parameter in a list when `type` of the parameter is `StringList`.
