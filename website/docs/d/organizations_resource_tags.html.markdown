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

* `resource_id` - (Required) ID of the resource with the tags to list. See details below.

### resource_id

You can specify any of the following taggable resources.

* AWS account – specify the account ID number.
* Organizational unit – specify the OU ID that begins with `ou-` and looks similar to: `ou-1a2b-34uvwxyz`
* Root – specify the root ID that begins with `r-` and looks similar to: `r-1a2b`
* Policy – specify the policy ID that begins with `p-` and looks similar to: `p-12abcdefg3`

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `tags` - Map of key=value pairs for each tag set on the resource.
