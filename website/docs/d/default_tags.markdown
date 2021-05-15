---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_default_tags"
description: |-
Expose the default tags configured on the provider.
---

# Data Source: aws_default_Tags

Use this data source to get the default tags configured on the provider.

It is intended to be used to optionally add the default tags to resources not _directly_ managed by the Terraform
resource - such as the instances underneath an autoscaling group or the volumes created for an instance.

## Example Usage

```terraform
provider "aws" {
  default_tags {
    tags = {
      Environment = "Test"
      Name = "Provider Tag"
    }
  }
}
data "aws_default_tags" "tags" {}

resource "aws_autoscaling_group" "group" {
  # ...
  dynamic "tag" {
    for_each = data.aws_default_tags.tags.tags
    content {
      key = tag.key
      value = tag.value
      propagate_at_launch = true
    }
  }
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `tags` - Any default tags set on the provider.
    * `tags.#.key` - The key name of the tag.
    * `tags.#.value` - The value of the tag.
