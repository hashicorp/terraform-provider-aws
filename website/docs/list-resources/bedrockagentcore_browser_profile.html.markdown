---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_browser_profile"
description: |-
  Lists Bedrock AgentCore Browser Profile resources.
---

# List Resource: aws_bedrockagentcore_browser_profile

Lists Bedrock AgentCore Browser Profile resources.

## Example Usage

```terraform
list "aws_bedrockagentcore_browser_profile" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
