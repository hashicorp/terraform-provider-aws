---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_registry"
description: |-
  Lists Bedrock AgentCore Registry resources.
---

# List Resource: aws_bedrockagentcore_registry

Lists Bedrock AgentCore Registry resources.

!> **Note:** AWS Agent Registry is currently available in public preview. [On August 6, 2026]((https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-faq.html#registry-faq-what-is-changing)) functionality will move from the `bedrock-agentcore` namespace to the `agent-registry` namespace. This resource will continue to work until [September 17, 2026](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-faq.html#registry-faq-continue-using).

## Example Usage

```terraform
list "aws_bedrockagentcore_registry" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
