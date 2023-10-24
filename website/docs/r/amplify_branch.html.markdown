---
subcategory: "Amplify"
layout: "aws"
page_title: "AWS: aws_amplify_branch"
description: |-
  Provides an Amplify Branch resource.
---

# Resource: aws_amplify_branch

Provides an Amplify Branch resource.

## Example Usage

```terraform
resource "aws_amplify_app" "example" {
  name = "app"
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.example.id
  branch_name = "master"

  framework = "React"
  stage     = "PRODUCTION"

  environment_variables = {
    REACT_APP_API_SERVER = "https://api.example.com"
  }
}
```

### Basic Authentication

```terraform
resource "aws_amplify_app" "example" {
  name = "app"
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.example.id
  branch_name = "master"

  enable_basic_auth      = true
  basic_auth_credentials = base64encode("username:password")
}
```

### Notifications

Amplify Console uses EventBridge (formerly known as CloudWatch Events) and SNS for email notifications.  To implement the same functionality, you need to set `enable_notification` in a `aws_amplify_branch` resource, as well as creating an EventBridge Rule, an SNS topic, and SNS subscriptions.

```terraform
resource "aws_amplify_app" "example" {
  name = "app"
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.example.id
  branch_name = "master"

  # Enable SNS notifications.
  enable_notification = true
}

# EventBridge Rule for Amplify notifications

resource "aws_cloudwatch_event_rule" "amplify_app_master" {
  name        = "amplify-${aws_amplify_app.app.id}-${aws_amplify_branch.master.branch_name}-branch-notification"
  description = "AWS Amplify build notifications for :  App: ${aws_amplify_app.app.id} Branch: ${aws_amplify_branch.master.branch_name}"

  event_pattern = jsonencode({
    "detail" = {
      "appId" = [
        aws_amplify_app.example.id
      ]
      "branchName" = [
        aws_amplify_branch.master.branch_name
      ],
      "jobStatus" = [
        "SUCCEED",
        "FAILED",
        "STARTED"
      ]
    }
    "detail-type" = [
      "Amplify Deployment Status Change"
    ]
    "source" = [
      "aws.amplify"
    ]
  })
}

resource "aws_cloudwatch_event_target" "amplify_app_master" {
  rule      = aws_cloudwatch_event_rule.amplify_app_master.name
  target_id = aws_amplify_branch.master.branch_name
  arn       = aws_sns_topic.amplify_app_master.arn

  input_transformer {
    input_paths = {
      jobId  = "$.detail.jobId"
      appId  = "$.detail.appId"
      region = "$.region"
      branch = "$.detail.branchName"
      status = "$.detail.jobStatus"
    }

    input_template = "\"Build notification from the AWS Amplify Console for app: https://<branch>.<appId>.amplifyapp.com/. Your build status is <status>. Go to https://console.aws.amazon.com/amplify/home?region=<region>#<appId>/<branch>/<jobId> to view details on your build. \""
  }
}

# SNS Topic for Amplify notifications

resource "aws_sns_topic" "amplify_app_master" {
  name = "amplify-${aws_amplify_app.app.id}_${aws_amplify_branch.master.branch_name}"
}

data "aws_iam_policy_document" "amplify_app_master" {
  statement {
    sid = "Allow_Publish_Events ${aws_amplify_branch.master.arn}"

    effect = "Allow"

    actions = [
      "SNS:Publish",
    ]

    principals {
      type = "Service"
      identifiers = [
        "events.amazonaws.com",
      ]
    }

    resources = [
      aws_sns_topic.amplify_app_master.arn,
    ]
  }
}

resource "aws_sns_topic_policy" "amplify_app_master" {
  arn    = aws_sns_topic.amplify_app_master.arn
  policy = data.aws_iam_policy_document.amplify_app_master.json
}

resource "aws_sns_topic_subscription" "this" {
  topic_arn = aws_sns_topic.amplify_app_master.arn
  protocol  = "email"
  endpoint  = "user@acme.com"
}
```

## Argument Reference

This resource supports the following arguments:

* `app_id` - (Required) Unique ID for an Amplify app.
* `branch_name` - (Required) Name for the branch.
* `backend_environment_arn` - (Optional) ARN for a backend environment that is part of an Amplify app.
* `basic_auth_credentials` - (Optional) Basic authorization credentials for the branch.
* `description` - (Optional) Description for the branch.
* `display_name` - (Optional) Display name for a branch. This is used as the default domain prefix.
* `enable_auto_build` - (Optional) Enables auto building for the branch.
* `enable_basic_auth` - (Optional) Enables basic authorization for the branch.
* `enable_notification` - (Optional) Enables notifications for the branch.
* `enable_performance_mode` - (Optional) Enables performance mode for the branch.
* `enable_pull_request_preview` - (Optional) Enables pull request previews for this branch.
* `environment_variables` - (Optional) Environment variables for the branch.
* `framework` - (Optional) Framework for the branch.
* `pull_request_environment_name` - (Optional) Amplify environment name for the pull request.
* `stage` - (Optional) Describes the current stage for the branch. Valid values: `PRODUCTION`, `BETA`, `DEVELOPMENT`, `EXPERIMENTAL`, `PULL_REQUEST`.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `ttl` - (Optional) Content Time To Live (TTL) for the website in seconds.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN for the branch.
* `associated_resources` - A list of custom resources that are linked to this branch.
* `custom_domains` - Custom domains for the branch.
* `destination_branch` - Destination branch if the branch is a pull request branch.
* `source_branch` - Source branch if the branch is a pull request branch.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amplify branch using `app_id` and `branch_name`. For example:

```terraform
import {
  to = aws_amplify_branch.master
  id = "d2ypk4k47z8u6/master"
}
```

Using `terraform import`, import Amplify branch using `app_id` and `branch_name`. For example:

```console
% terraform import aws_amplify_branch.master d2ypk4k47z8u6/master
```
