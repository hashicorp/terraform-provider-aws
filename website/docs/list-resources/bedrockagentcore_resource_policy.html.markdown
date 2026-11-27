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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
