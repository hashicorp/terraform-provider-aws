---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_evaluator"
description: |-
  Lists Bedrock AgentCore Evaluator resources.
---

# List Resource: aws_bedrockagentcore_evaluator

Lists Bedrock AgentCore Evaluator resources.

## Example Usage

```terraform
list "aws_bedrockagentcore_evaluator" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
