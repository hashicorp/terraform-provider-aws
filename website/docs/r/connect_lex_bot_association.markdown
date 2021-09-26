---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_lex_bot_association"
description: |-
  Provides details about a specific Connect Lex Bot Association.
---

# Resource: aws_connect_lex_bot_association

Allows the specified Amazon Connect instance to access the specified Amazon Lex bot. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

## Example Usage
### Basic
```hcl
resource "aws_connect_lex_bot_association" "test" {
  instance_id = aws_connect_instance.test.id
  name        = "Test"
  region      = "us-west-2"
}
```
## Argument Reference

The following arguments are supported:
* `instance_id` - (Required) The identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.
* `name` - (Required) The name of the Amazon Lex bot.
* `region` - (Required) The Region in which the Amazon Lex bot has been created.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 5 mins) Used when creating the association.
* `delete` - (Defaults to 5 mins) Used when creating the association.

## Attributes Reference
