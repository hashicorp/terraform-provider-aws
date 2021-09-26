---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_lex_bot_association"
description: |-
  Provides details about a specific Connect Lex Bot Association.
---

# Data Source: aws_connect_lex_bot_association

Provides details about a specific Connect Lex Bot Association.

## Example Usage
By name

```hcl
data "aws_connect_contact_flow" "test" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Test"
}
```

By name with region

```hcl
data "aws_connect_contact_flow" "test" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Test"
  region      = "us-west-2"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) The identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.
* `name` - (Required) The name of the Amazon Lex bot.
* `region` - (Optional) The Region in which the Amazon Lex bot has been created.

## Attributes Reference
