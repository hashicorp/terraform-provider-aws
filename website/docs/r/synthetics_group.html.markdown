---
subcategory: "CloudWatch Synthetics"
layout: "aws"
page_title: "AWS: aws_synthetics_group"
description: |-
  Provides a Synthetics Group resource
---

# Resource: aws_synthetics_group

Provides a Synthetics Group resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_synthetics_group" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the group.

The following arguments are optional:

* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Group.
* `group_id` - ID of the Group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

CloudWatch Synthetics Group can be imported using the `name`, e.g.,

```
$ terraform import aws_synthetics_group.example example
```
