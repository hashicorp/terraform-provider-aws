---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_connection"
description: |-
  Provides an EventBridge connection data source.
---

# Data Source: aws_cloudwatch_event_connection

Use this data source to retrieve information about an EventBridge connection.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```terraform
data "aws_cloudwatch_event_connection" "test" {
  name = "test"
}
```

## Argument Reference

* `name` - Name of the connection.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `name` - Name of the connection.

* `arn` - ARN (Amazon Resource Name) for the connection.

* `secret_arn` - ARN (Amazon Resource Name) for the secret created from the authorization parameters specified for the connection.

* `authorization_type` - Type of authorization to use to connect. One of `API_KEY`,`BASIC`,`OAUTH_CLIENT_CREDENTIALS`.
