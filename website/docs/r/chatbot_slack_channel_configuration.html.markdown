---
subcategory: "Chatbot"
layout: "aws"
page_title: "AWS: aws_chatbot_slack_channel_configuration"
description: |-
  Terraform resource for managing an AWS Chatbot Slack Channel Configuration.
---

# Resource: aws_chatbot_slack_channel_configuration

Terraform resource for managing an AWS Chatbot Slack Channel Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_chatbot_slack_channel_configuration" "example" {
  configuration_name = "example_aws_slack_channel_config"
  iam_role_arn       = aws_iam_role.test_chatbot_role.arn
  slack_channel_id   = "ABCD1234"
  slack_team_id      = data.aws_chatbot_slack_workspace.example.slack_team_id
}
```

### Usage with all the arguments

```terraform
resource "aws_chatbot_slack_channel_configuration" "example" {
  configuration_name          = "example_aws_slack_channel_config"
  guardrail_policy_arns       = [aws_iam_policy.test_guardrail_policy.arn]
  iam_role_arn                = aws_iam_role.test_chatbot_role.arn
  logging_level               = "INFO"
  slack_channel_id            = "ABCD1234"
  slack_channel_name          = "example-channel"
  slack_team_id               = data.aws_chatbot_slack_workspace.test.slack_team_id
  sns_topic_arns              = [aws_sns_topic.test_sns_topic.arn]
  user_authorization_required = false
  tags = {
    key1 = value1
  }
}
```

## Argument Reference

The following arguments are required:

* `configuration_name` - (Required) Configuration Name.

* `iam_role_arn` - (Required) ARN of the user-defined IAM role that will be assumed by AWS Chatbot.

* `slack_channel_id` - (Required) ID of the Slack channel. To get the ID, open Slack, right click on the channel name in the left pane, then choose Copy Link.

* `slack_team_id` - (Required) ID of the Slack workspace authorized with AWS Chatbot.

The following arguments are optional:

* `guardrail_policy_arns` - (Optional) ARNs of IAM policies that are applied as channel guardrails.

* `logging_level` - (Optional) Logging level. Valid values are ERROR, INFO, AND NONE.

* `slack_channel_name` - (Optional) Name of the Slack channel. if a slack channel name is not provided, the servie tries to get the name from the channel itself. The service will be unable to get a name if the channel is private and the @aws bot is not added to the channel, in which case, the slack channel name is left as blank. If a slack channel name is provided, it will override any name received from Slack.

* `sns_topic_arns` - (Optional) ARNs of SNS topics that deliver notifications to AWS Chatbot.

* `user_authorization_required` - (Optional) Flag whether user roles are required in the Slack Channel. Please refer to [Understanding Permissions](https://docs.aws.amazon.com/chatbot/latest/adminguide/understanding-permissions.html) page in the documentation for more details.

* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `chat_configuration_arn` - ARN of the Slack Channel Configuration.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Chatbot Slack Channel Configuration using the `id`. For example:

```terraform
import {
  to = aws_chatbot_slack_channel_configuration.example
  id = "arn:aws:chatbot::111222333444:chat-configuration/slack-channel/example_aws_slack_channel_config"
}
```

Using `terraform import`, import Chatbot Slack Channel Configuration using the `id`. For example:

```console
% terraform import aws_chatbot_slack_channel_configuration.example arn:aws:chatbot::111222333444:chat-configuration/slack-channel/example_aws_slack_channel_config
```
