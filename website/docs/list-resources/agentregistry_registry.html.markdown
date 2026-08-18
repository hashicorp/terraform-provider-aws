---
subcategory: "Agent Registry"
layout: "aws"
page_title: "AWS: aws_agentregistry_registry"
description: |-
  Lists Agent Registry Registry resources.
---

# List Resource: aws_agentregistry_registry

Lists Agent Registry Registry resources.

## Example Usage

```terraform
list "aws_agentregistry_registry" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
