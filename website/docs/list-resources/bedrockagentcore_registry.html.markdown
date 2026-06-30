---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_registry"
description: |-
  Lists Bedrock AgentCore Registry resources.
---

# List Resource: aws_bedrockagentcore_registry

Lists Bedrock AgentCore Registry resources.

## Example Usage

```terraform
list "aws_bedrockagentcore_registry" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
