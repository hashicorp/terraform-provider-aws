---
subcategory: "Chatbot"
layout: "aws"
page_title: "AWS: aws_chatbot_teams_channel_configuration"
description: |-
  Terraform resource for managing an AWS Chatbot Microsoft Teams Channel Configuration.
---

# Resource: aws_chatbot_teams_channel_configuration

Terraform resource for managing an AWS Chatbot Microsoft Teams Channel Configuration.

~> **NOTE:** We provide this resource on a best-effort basis. If you are able to test it and find it useful, we welcome your input at [GitHub](https://github.com/hashicorp/terraform-provider-aws).

## Example Usage

### Basic Usage

```terraform
resource "aws_chatbot_teams_channel_configuration" "test" {
  channel_id         = "C07EZ1ABC23"
  configuration_name = "mitt-lags-kanal"
  iam_role_arn       = aws_iam_role.test.arn
  team_id            = "74361522-da01-538d-aa2e-ac7918c6bb92"
  tenant_id          = "1234"

  tags = {
    Name = "mitt-lags-kanal"
  }
}
```

## Argument Reference

The following arguments are required:

* `channel_id` - (Required) ID of the Microsoft Teams channel.
* `configuration_name` - (Required) Name of the Microsoft Teams channel configuration.
* `iam_role_arn` - (Required) ARN of the IAM role that defines the permissions for AWS Chatbot. This is a user-defined role that AWS Chatbot will assume. This is not the service-linked role.
* `team_id` - (Required) ID of the Microsoft Team authorized with AWS Chatbot. To get the team ID, you must perform the initial authorization flow with Microsoft Teams in the AWS Chatbot console. Then you can copy and paste the team ID from the console.
* `tenant_id` - (Required) ID of the Microsoft Teams tenant.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `channel_name` - (Optional) Name of the Microsoft Teams channel.
* `guardrail_policy_arns` - (Optional) List of IAM policy ARNs that are applied as channel guardrails. The AWS managed `AdministratorAccess` policy is applied by default if this is not set.
* `logging_level` - (Optional) Logging levels include `ERROR`, `INFO`, or `NONE`.
* `sns_topic_arns` - (Optional) ARNs of the SNS topics that deliver notifications to AWS Chatbot.
* `tags` - (Optional) Map of tags assigned to the resource.
* `team_name` - (Optional) Name of the Microsoft Teams team.
* `user_authorization_required` - (Optional) Enables use of a user role requirement in your chat configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `chat_configuration_arn` - ARN of the Microsoft Teams channel configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Chatbot Microsoft Teams Channel Configuration using the `team_id`. For example:

```terraform
import {
  to = aws_chatbot_teams_channel_configuration.example
  id = "5f4f15d2-b958-522a-8333-124aa8bf0925"
}
```

Using `terraform import`, import Chatbot Microsoft Teams Channel Configuration using the `team_id`. For example:

```console
% terraform import aws_chatbot_teams_channel_configuration.example 5f4f15d2-b958-522a-8333-124aa8bf0925
```
