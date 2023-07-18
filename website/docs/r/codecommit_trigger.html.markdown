---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_trigger"
description: |-
  Provides a CodeCommit Trigger Resource.
---

# Resource: aws_codecommit_trigger

Provides a CodeCommit Trigger Resource.

~> **NOTE:** Terraform currently can create only one trigger per repository, even if multiple aws_codecommit_trigger resources are defined. Moreover, creating triggers with Terraform will delete all other triggers in the repository (also manually-created triggers).

## Example Usage

```terraform
resource "aws_codecommit_repository" "test" {
  repository_name = "test"
}

resource "aws_codecommit_trigger" "test" {
  repository_name = aws_codecommit_repository.test.repository_name

  trigger {
    name            = "all"
    events          = ["all"]
    destination_arn = aws_sns_topic.test.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `repository_name` - (Required) The name for the repository. This needs to be less than 100 characters.
* `name` - (Required) The name of the trigger.
* `destination_arn` - (Required) The ARN of the resource that is the target for a trigger. For example, the ARN of a topic in Amazon Simple Notification Service (SNS).
* `custom_data` - (Optional) Any custom data associated with the trigger that will be included in the information sent to the target of the trigger.
* `branches` - (Optional) The branches that will be included in the trigger configuration. If no branches are specified, the trigger will apply to all branches.
* `events` - (Required) The repository events that will cause the trigger to run actions in another service, such as sending a notification through Amazon Simple Notification Service (SNS). If no events are specified, the trigger will run for all repository events. Event types include: `all`, `updateReference`, `createReference`, `deleteReference`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `configuration_id` - System-generated unique identifier.
