---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_gateway_rule"
description: |-
  Lists Bedrock AgentCore Gateway Rule resources.
---

# List Resource: aws_bedrockagentcore_gateway_rule

Lists Bedrock AgentCore Gateway Rule resources.

## Example Usage

```terraform
list "aws_bedrockagentcore_gateway_rule" "example" {
  provider = aws

  config {
    gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `gateway_identifier` - (Required) ID of the gateway whose rules to list.
* `region` - (Optional) Region to query. Defaults to provider region.
