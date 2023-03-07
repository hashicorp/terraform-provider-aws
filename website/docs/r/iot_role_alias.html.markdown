---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_role_alias"
description: |-
  Provides an IoT role alias.
---

# Resource: aws_iot_role_alias

Provides an IoT role alias.

## Example Usage

```terraform
data "aws_iam_policy_document" "assume_role" {
  effect = "Allow"

  principals {
    type        = "Service"
    identifiers = ["credentials.iot.amazonaws.com"]
  }

  actions = ["sts:AssumeRole"]
}

resource "aws_iam_role" "role" {
  name               = "dynamodb-access-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iot_role_alias" "alias" {
  alias    = "Thermostat-dynamodb-access-role-alias"
  role_arn = aws_iam_role.role.arn
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Required) The name of the role alias.
* `role_arn` - (Required) The identity of the role to which the alias refers.
* `credential_duration` - (Optional) The duration of the credential, in seconds. If you do not specify a value for this setting, the default maximum of one hour is applied. This setting can have a value from 900 seconds (15 minutes) to 43200 seconds (12 hours).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN assigned by AWS to this role alias.

## Import

IOT Role Alias can be imported via the alias, e.g.,

```sh
$ terraform import aws_iot_role_alias.example myalias
```
