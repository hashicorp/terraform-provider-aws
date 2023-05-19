---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_extension_association"
description: |-
  Associates an AppConfig Extension with a Resource.
---

# Resource: aws_appconfig_extension_association

Associates an AppConfig Extension with a Resource.

## Example Usage

```terraform
resource "aws_sns_topic" "test" {
  name = "test"
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["appconfig.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = "test"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_appconfig_extension" "test" {
  name        = "test"
  description = "test description"
  action_point {
    point = "ON_DEPLOYMENT_COMPLETE"
    action {
      name     = "test"
      role_arn = aws_iam_role.test.arn
      uri      = aws_sns_topic.test.arn
    }
  }
  tags = {
    Type = "AppConfig Extension"
  }
}

resource "aws_appconfig_application" "test" {
  name = "test"
}

resource "aws_appconfig_extension_association" "test" {
  extension_arn = aws_appconfig_extension.test.arn
  resource_arn  = aws_appconfig_application.test.arn
}
```

## Argument Reference

The following arguments are supported:

* `extension_arn` - (Required) The ARN of the extension defined in the association.
* `resource_arn` - (Optional) The ARN of the application, configuration profile, or environment to associate with the extension.
* `parameters` - (Optional) The parameter names and values defined for the association.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the AppConfig Extension Association.
* `id` - AppConfig Extension Association ID.
* `extension_version` - The version number for the extension defined in the association.

## Import

AppConfig Extension Associations can be imported using their extension association ID, e.g.,

```
$ terraform import aws_appconfig_extension_association.example 71rxuzt
```
