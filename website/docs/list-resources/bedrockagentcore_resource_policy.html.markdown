---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_resource_policy"
description: |-
  Lists Bedrock AgentCore Resource Policy resources.
---

# List Resource: aws_bedrockagentcore_resource_policy

Lists Bedrock AgentCore Resource Policy resources.

## Example Usage

```terraform
list "aws_bedrockagentcore_resource_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
