---
subcategory: "Security Lake"
layout: "aws"
page_title: "AWS: aws_securitylake_subscriber_notification"
description: |-
  Terraform resource for managing an AWS Security Lake Subscriber Notification.
---

# Resource: aws_securitylake_subscriber_notification

Terraform resource for managing an AWS Security Lake Subscriber Notification.

## Example Usage

### SQS Notification

```terraform
resource "aws_securitylake_subscriber_notification" "example" {
  subscriber_id = aws_securitylake_subscriber.example.id
  configuration {
    sqs_notification_configuration {}
  }
}
```

### HTTPS Notification

```terraform
resource "aws_securitylake_subscriber_notification" "example" {
  subscriber_id = aws_securitylake_subscriber.example.id
  configuration {
    https_notification_configuration {
      endpoint        = aws_apigatewayv2_api.test.api_endpoint
      target_role_arn = aws_iam_role.event_bridge.arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `subscriber_id` - (Required) The subscriber ID for the notification subscription.
* `configuration` - (Required) Specify the configuration using which you want to create the subscriber notification..

Configuration support the following:

* `sqs_notification_configuration` - (Optional) The configurations for SQS subscriber notification.
  There are no parameters within `sqs_notification_configuration`.
* `https_notification_configuration` - (Optional) The configurations for HTTPS subscriber notification.

HTTPS Notification Configuration support the following:

* `endpoint` - (Required) The subscription endpoint in Security Lake.
  If you prefer notification with an HTTPS endpoint, populate this field.
* `target_role_arn` - (Required) The Amazon Resource Name (ARN) of the EventBridge API destinations IAM role that you created.
  For more information about ARNs and how to use them in policies, see Managing data access and AWS Managed Policies in the Amazon Security Lake User Guide.
* `authorization_api_key_name` - (Optional) The API key name for the notification subscription.
* `authorization_api_key_value` - (Optional) The API key value for the notification subscription.
* `http_method` - (Optional) The HTTP method used for the notification subscription.
  Valid values are `POST` and `PUT`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `endpoint_id` - (**Deprecated**) The subscriber endpoint to which exception messages are posted.
* `subscriber_endpoint` - The subscriber endpoint to which exception messages are posted.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)
