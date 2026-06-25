---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_registry_record"
description: |-
  Lists Bedrock AgentCore Registry Record resources.
---

# List Resource: aws_bedrockagentcore_registry_record

Lists Bedrock AgentCore Registry Record resources.

## Example Usage

```terraform
list "aws_bedrockagentcore_registry_record" "example" {
  provider = aws

  config {
    registry_id = aws_bedrockagentcore_registry.example.registry_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `registry_id` - (Required) Registry ID.
