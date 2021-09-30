---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_bot_association"
description: |-
  Provides details about a specific Connect Bot Association.
---

# Data Source: aws_connect_bot_association

Provides details about a specific Connect Bot Association.

## Example Usage
By name

```hcl
data "aws_connect_bot_association" "test" {
  bot_name    = "Test"
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

The following arguments are supported:

* `bot_name` - (Required) The name of the Amazon Lex bot.
* `instance_id` - (Required) The identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.

## Attributes Reference
