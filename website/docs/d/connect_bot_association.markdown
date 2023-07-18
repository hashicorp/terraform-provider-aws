---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_bot_association"
description: |-
  Provides details about a specific Lex (V1) Bot associated with an Amazon Connect instance
---

# Data Source: aws_connect_bot_association

Provides details about a specific Lex (V1) Bot associated with an Amazon Connect instance.

## Example Usage

### By name

```terraform
data "aws_connect_bot_association" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  lex_bot {
    name = "Test"
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `instance_id` - (Required) Identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.
* `lex_bot` - (Required) Configuration information of an Amazon Lex (V1) bot. Detailed below.

### lex_bot

The `lex_bot` configuration block supports the following:

* `name` - (Required) Name of the Amazon Lex (V1) bot.
* `lex_region` - (Optional) Region that the Amazon Lex (V1) bot was created in.
