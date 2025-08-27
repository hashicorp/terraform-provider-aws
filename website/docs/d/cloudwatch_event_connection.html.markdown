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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - Name of the connection.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN (Amazon Resource Name) of the connection.
* `authorization_type` - Type of authorization specified for the connection. One of `API_KEY`,`BASIC`,`OAUTH_CLIENT_CREDENTIALS`.
* `kms_key_identifier` - (Optional) Identifier of the AWS KMS customer managed key for EventBridge to use to encrypt the connection, if one has been specified.
* `secret_arn` - ARN of the secret created from the authorization parameters specified for the connection.
