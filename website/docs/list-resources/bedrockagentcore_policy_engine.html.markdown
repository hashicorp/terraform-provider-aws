---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_policy_engine"
description: |-
  Lists Bedrock AgentCore Policy Engine resources.
---

# List Resource: aws_bedrockagentcore_policy_engine

Lists Bedrock AgentCore Policy Engine resources.

## Example Usage

```terraform
list "aws_bedrockagentcore_policy_engine" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
