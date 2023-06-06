---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_extension"
description: |-
  Provides an AppConfig Extension resource.
---

# Resource: aws_appconfig_extension

Provides an AppConfig Extension resource.

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
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the extension. Each extension name in your account must be unique. Extension versions use the same name.
* `description` - (Optional) Information about the extension.
* `action_point` - (Required) The action points defined in the extension. [Detailed below](#action_point).
* `parameter` - (Optional) The parameters accepted by the extension. You specify parameter values when you associate the extension to an AppConfig resource by using the CreateExtensionAssociation API action. For Lambda extension actions, these parameters are included in the Lambda request object. [Detailed below](#parameter).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `action_point`

Defines the actions the extension performs during the AppConfig workflow and at which point those actions are performed. The `action_point` configuration block supports the following arguments:

* `point` - (Required) The point at which to perform the defined actions. Valid points are `PRE_CREATE_HOSTED_CONFIGURATION_VERSION`, `PRE_START_DEPLOYMENT`, `ON_DEPLOYMENT_START`, `ON_DEPLOYMENT_STEP`, `ON_DEPLOYMENT_BAKING`, `ON_DEPLOYMENT_COMPLETE`, `ON_DEPLOYMENT_ROLLED_BACK`.
* `action` - (Required) An action defines the tasks the extension performs during the AppConfig workflow. [Detailed below](#action).

#### `action`

The `action` configuration block supports configuring any number of the following arguments:

* `name` - (Required) The action name.
* `role_arn` - (Required) An Amazon Resource Name (ARN) for an Identity and Access Management assume role.
* `uri` - (Required) The extension URI associated to the action point in the extension definition. The URI can be an Amazon Resource Name (ARN) for one of the following: an Lambda function, an Amazon Simple Queue Service queue, an Amazon Simple Notification Service topic, or the Amazon EventBridge default event bus.
* `description` - (Optional) Information about the action.

#### `parameter`

The `parameter` configuration block supports configuring any number of the following arguments:

* `name` - (Required) The parameter name.
* `required` - (Required) Determines if a parameter value must be specified in the extension association.
* `description` - (Optional) Information about the parameter.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the AppConfig Extension.
* `id` - AppConfig Extension ID.
* `version` - The version number for the extension.

## Import

AppConfig Extensions can be imported using their extension ID, e.g.,

```
$ terraform import aws_appconfig_extension.example 71rxuzt
```
