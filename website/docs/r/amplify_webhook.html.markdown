---
subcategory: "Amplify"
layout: "aws"
page_title: "AWS: aws_amplify_webhook"
description: |-
  Provides an Amplify webhook resource.
---

# Resource: aws_amplify_webhook

Provides an Amplify webhook resource.

## Example Usage

```hcl
resource "aws_amplify_app" "app" {
  name = "app"
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.app.id
  branch_name = "master"
}

resource "aws_amplify_webhook" "master" {
  app_id      = aws_amplify_app.app.id
  branch_name = aws_amplify_branch.master.branch_name
  description = "triggermaster"
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) Unique Id for an Amplify App.
* `branch_name` - (Required) Name for a branch.
* `description` - (Optional) Description for a webhook

## Attribute Reference

The following attributes are exported:

* `arn` - ARN for the webhook.
* `url` - Url of the webhook.

## Import

Amplify webhook can be imported using a webhook ID (webhookId), e.g.

```
$ terraform import aws_amplify_webhook.master a26b22a0-748b-4b57-b9a0-ae7e601fe4b1
```
