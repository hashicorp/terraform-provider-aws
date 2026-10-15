---
subcategory: "Chatbot"
layout: "aws"
page_title: "AWS: aws_chatbot_slack_workspace"
description: |-
  Terraform data source for managing an AWS Chatbot Slack Workspace.
---

# Data Source: aws_chatbot_slack_workspace

Terraform data source for managing an AWS Chatbot Slack Workspace.

## Example Usage

### Basic Usage

```terraform
data "aws_chatbot_slack_workspace" "example" {
  slack_team_name = "abc"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `slack_team_name` - (Required) Slack workspace name configured with AWS Chatbot.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `slack_team_id` - ID of the Slack Workspace assigned by AWS Chatbot.
