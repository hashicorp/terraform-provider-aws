---
subcategory: "SFN (Step Functions)"
layout: "aws"
page_title: "AWS: aws_sfn_activity"
description: |-
  Provides a Step Function Activity resource.
---

# Resource: aws_sfn_activity

Provides a Step Function Activity resource

## Example Usage

```terraform
resource "aws_sfn_activity" "sfn_activity" {
  name = "my-activity"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the activity to create.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) that identifies the created activity.
* `name` - The name of the activity.
* `creation_date` - The date the activity was created.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Import Activities using the `arn`. For example:

```
$ terraform import aws_sfn_activity.foo arn:aws:states:eu-west-1:123456789098:activity:bar
```
