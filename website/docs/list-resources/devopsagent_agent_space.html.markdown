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

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
