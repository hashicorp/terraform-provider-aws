---
layout: "aws"
page_title: "AWS: aws_lex_code_hook"
sidebar_current: "docs-aws-resource-lex-code-hook"
description: |-
  Definition of an Amazon Lex Code Hook used as an attribute in other Lex resources.
---

# aws_lex_code_hook

Specifies a Lambda function that verifies requests to a bot or fulfills the user's request to a bot.

## Example Usage

```hcl
resource "aws_lex_intent" "flowers" {
  dialog_code_hook {
    message_version = "1"
    uri             = "arn:aws:lambda:us-east-1:123456789012:function:RetrieveAvailableFlowers"
  }

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

* `message_version`

	The version of the request-response that you want Amazon Lex to use to invoke your Lambda
	function. For more information, see
	[Using Lambda Functions](https://docs.aws.amazon.com/lex/latest/dg/using-lambda.html).

    * Type: string
    * Min: 1
    * Max: 5

* `uri`

    The Amazon Resource Name (ARN) of the Lambda function.

    * Type: string
    * Min: 20
    * Max: 2048
    * Pattern: arn:aws:lambda:[a-z]+-[a-z]+-[0-9]:[0-9]{12}:function:[a-zA-Z0-9-_]+(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})?(:[a-zA-Z0-9-_]+)?
