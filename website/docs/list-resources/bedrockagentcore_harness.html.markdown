---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_harness"
description: |-
  Lists Bedrock AgentCore Harness resources.
---

# List Resource: aws_bedrockagentcore_harness

Lists Bedrock AgentCore Harness resources.

## Example Usage

```terraform
list "aws_bedrockagentcore_harness" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
