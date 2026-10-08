---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_agent_space"
description: |-
  Lists AWS DevOps Agent Spaces.
---

# List Resource: aws_devopsagent_agent_space

Lists AWS DevOps Agent Spaces.

## Example Usage

### Basic Usage

```terraform
list "aws_devopsagent_agent_space" "example" {
  provider = aws
}
```

### Include Resource Data

```terraform
list "aws_devopsagent_agent_space" "example" {
  provider         = aws
  include_resource = true
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
