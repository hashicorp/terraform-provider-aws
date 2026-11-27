---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_resource_tags"
description: |-
  Get tags attached to the specified AWS Organizations resource.
---

# Data Source: aws_organizations_resource_tags

Get tags attached to the specified AWS Organizations resource.

## Example Usage

```terraform
data "aws_organizations_resource_tags" "account" {
  resource_id = "123456123846"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `resource_id` - (Required) ID of the resource with the tags to list. You can specify any of the following taggable resources: an AWS account (specify the account ID number); an organizational unit (specify the OU ID that begins with `ou-` and looks similar to `ou-1a2b-34uvwxyz`); a root (specify the root ID that begins with `r-` and looks similar to `r-1a2b`); or a policy (specify the policy ID that begins with `p-` and looks similar to `p-12abcdefg3`).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `tags` - Map of key=value pairs for each tag set on the resource.
