---
subcategory: "Amplify"
layout: "aws"
page_title: "AWS: aws_amplify_branch"
description: |-
  Provides an Amplify branch resource.
---

# Resource: aws_amplify_branch

Provides an Amplify branch resource.

## Example Usage

```hcl
resource "aws_amplify_app" "app" {
  name = "app"
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.app.id
  branch_name = "master"

  framework   = "React"
  stage       = "PRODUCTION"

  environment_variables = {
    REACT_APP_API_SERVER = "https://api.example.com"
  }
}
```

### Basic Authentication

```hcl
resource "aws_amplify_app" "app" {
  name = "app"
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.app.id
  branch_name = "master"

  basic_auth_config {
    // Enable basic authentication.
    enable_basic_auth = true

    username = "username"
    password = "password"
  }
}
```

### Notifications

Amplify uses SNS for email notifications.  In order to enable notifications, you have to set `enable_notification` in a `aws_amplify_branch` resource, as well as creating a SNS topic and subscriptions.  Use `sns_topic_name` to get a SNS topic name.  You can select any subscription protocol, such as `lambda`.

```hcl
resource "aws_amplify_app" "app" {
  name = "app"
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.app.id
  branch_name = "master"

  // Enable SNS notifications.
  enable_notification = true
}

resource "aws_sns_topic" "amplify_app_master" {
  name = aws_amplify_branch.master.sns_topic_name
}

resource "aws_sns_topic_subscription" "amplify_app_master_lambda" {
  topic_arn = aws_sns_topic.amplify_app_master.arn
  protocol  = "lambda"
  endpoint  = "arn:aws:lambda:..."
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) Unique Id for an Amplify App.
* `branch_name` - (Required) Name for the branch.
* `backend_environment_arn` - (Optional) ARN for a Backend Environment, part of an Amplify App.
* `basic_auth_config` - (Optional) Basic Authentication config for the branch. A `basic_auth_config` block is documented below.
* `build_spec` - (Optional) BuildSpec for the branch.
* `description` - (Optional) Description for the branch.
* `display_name` - (Optional) Display name for a branch, will use as the default domain prefix.
* `enable_auto_build` - (Optional) Enables auto building for the branch.
* `enable_notifications` - (Optional) Enables notifications for the branch.
* `enable_pull_request_preview` - (Optional) Enables Pull Request Preview for this branch.
* `environment_variables` - (Optional) Environment Variables for the branch.
* `framework` - (Optional) Framework for the branch.
* `pull_request_environment_name` - (Optional) The Amplify Environment name for the pull request.
* `stage` - (Optional) Stage for the branch. Possible values: "PRODUCTION", "BETA", "DEVELOPMENT", "EXPERIMENTAL", or "PULL_REQUEST".
* `tags` - (Optional) Key-value mapping of resource tags.
* `ttl` - (Optional) The content TTL for the website in seconds.

An `basic_auth_config` block supports the following arguments:

* `enable_basic_auth` - (Optional) Enables Basic Authorization.
* `username` - (Optional) Basic Authorization username.
* `password` - (Optional) Basic Authorization password.

## Attribute Reference

The following attributes are exported:

* `arn` - ARN for the Amplify App.
* `associated_resources` - List of custom resources that are linked to this branch.
* `custom_domains` - Custom domains for a branch, part of an Amplify App.
* `destination_branch` - The destination branch if the branch is a pull request branch.
* `sns_topic_name` - SNS topic name for notifications.
* `source_branch` - The source branch if the branch is a pull request branch.

## Import

Amplify branch can be imported using `app_id` and `branch_name`, e.g.

```
$ terraform import aws_amplify_branch.master d2ypk4k47z8u6/branches/master
```
