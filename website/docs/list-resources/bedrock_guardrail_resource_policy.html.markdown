---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_guardrail_resource_policy"
description: |-
  Lists Bedrock Guardrail Resource Policy resources.
---

# List Resource: aws_bedrock_guardrail_resource_policy

Lists Bedrock Guardrail Resource Policy resources.

Results include resource policies attached to guardrails. Resource policies attached to system-defined guardrail profiles are not listed.

## Example Usage

```terraform
list "aws_bedrock_guardrail_resource_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
