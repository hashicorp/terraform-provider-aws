---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_policy"
description: |-
  Lists Bedrock AgentCore Policy resources.
---

# List Resource: aws_bedrockagentcore_policy

Lists Bedrock AgentCore Policy resources.

## Example Usage

```terraform
list "aws_bedrockagentcore_policy" "example" {
  provider = aws

  config {
    policy_engine_id = aws_bedrockagentcore_policy_engine.example.policy_engine_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `policy_engine_id` - (Required) ID of the policy engine whose policies to list.
* `region` - (Optional) Region to query. Defaults to provider region.
