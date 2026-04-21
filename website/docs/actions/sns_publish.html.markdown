---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_publish"
description: |-
  Publishes a message to an Amazon SNS topic.
---

# Action: aws_sns_publish

Publishes a message to an Amazon SNS topic. This action allows for imperative message publishing with full control over message attributes and structure.

For information about Amazon SNS, see the [Amazon SNS Developer Guide](https://docs.aws.amazon.com/sns/latest/dg/). For specific information about publishing messages, see the [Publish](https://docs.aws.amazon.com/sns/latest/api/API_Publish.html) page in the Amazon SNS API Reference.

## Example Usage

### Basic Usage

```terraform
resource "aws_sns_topic" "example" {
  name = "example-topic"
}

action "aws_sns_publish" "example" {
  config {
    topic_arn = aws_sns_topic.example.arn
    message   = "Hello from Terraform!"
  }
}

resource "terraform_data" "example" {
  input = "trigger-message"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_sns_publish.example]
    }
  }
}
```

### Message with Subject

```terraform
action "aws_sns_publish" "notification" {
  config {
    topic_arn = aws_sns_topic.alerts.arn
    subject   = "System Alert"
    message   = "Critical system event detected at ${timestamp()}"
  }
}
```

### JSON Message Structure

```terraform
action "aws_sns_publish" "structured" {
  config {
    topic_arn         = aws_sns_topic.mobile.arn
    message_structure = "json"
    message = jsonencode({
      default = "Default message"
      email   = "Email version of the message"
      sms     = "SMS version"
      GCM = jsonencode({
        data = {
          message = "Push notification message"
        }
      })
    })
  }
}
```

### Message with Attributes

```terraform
action "aws_sns_publish" "with_attributes" {
  config {
    topic_arn = aws_sns_topic.processing.arn
    message   = "Process this data"

    message_attributes {
      map_block_key = "priority"
      data_type     = "String"
      string_value  = "high"
    }

    message_attributes {
      map_block_key = "source"
      data_type     = "String"
      string_value  = "terraform"
    }
  }
}
```

### Deployment Notification

```terraform
action "aws_sns_publish" "deploy_complete" {
  config {
    topic_arn = aws_sns_topic.deployments.arn
    subject   = "Deployment Complete"
    message = jsonencode({
      environment = var.environment
      version     = var.app_version
      timestamp   = timestamp()
      resources = {
        instances = length(aws_instance.app)
        databases = length(aws_db_instance.main)
      }
    })
  }
}

resource "terraform_data" "deploy_trigger" {
  input = var.deployment_id

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_sns_publish.deploy_complete]
    }
  }

  depends_on = [aws_instance.app, aws_db_instance.main]
}
```

## Argument Reference

This action supports the following arguments:

* `message` - (Required) Message to publish. For JSON message structure, this should be a JSON object with protocol-specific messages. Maximum size is 256 KB.
* `message_attributes` - (Optional) Message attributes to include with the message. Each attribute consists of a name, data type, and value. Up to 10 attributes are allowed. [See below.](#message-attributes)
* `message_structure` - (Optional) Set to `json` if you want to send different messages for each protocol. If not specified, the message will be sent as-is to all protocols.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subject` - (Optional) Optional subject for the message. Only used for email and email-json protocols. Maximum length is 100 characters.
* `topic_arn` - (Required) ARN of the SNS topic to publish the message to.

### Message Attributes

The `message_attributes` block supports:

* `data_type` - (Required) Data type of the message attribute. Valid values are `String`, `Number`, and `Binary`.
* `map_block_key` - (Required) Name of the message attribute (used as map key). Must be unique within the message.
* `string_value` - (Required) Value of the message attribute.
