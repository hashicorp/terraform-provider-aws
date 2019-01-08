---
layout: "aws"
page_title: "AWS: aws_lex_fulfillment_activity"
sidebar_current: "docs-aws-resource-lex-fulfillment-activity"
description: |-
  Definition of an Amazon Lex Fulfillment Activity used as an attribute in other Lex resources.
---

# aws_lex_fulfillment_activity

Describes how the intent is fulfilled after the user provides all of the information required for the intent.

## Example Usage

```hcl
resource "aws_lex_intent" "flowers" {
  fulfillment_activity {
    code_hook {
      message_version = "1"
      uri             = "arn:aws:lambda:us-east-1:123456789012:function:ProcessFlowerOrder"
    }

    type = "CodeHook"
  }
}
```

## Argument Reference

The following arguments are supported:

### Required

* `type`

    How the intent should be fulfilled, either by running a Lambda function or by returning the
    slot data to the client application.

    * Type: string
    * Values: ReturnIntent | CodeHook

### Optional

* `code_hook`

    A description of the Lambda function that is run to fulfill the intent. Required if type is CodeHook.

    * Type: [Lex Code Hook](/docs/providers/aws/r/lex_code_hook.html)
