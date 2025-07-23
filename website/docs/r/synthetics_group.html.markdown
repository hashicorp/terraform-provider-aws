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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Group.
* `group_id` - ID of the Group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Synthetics Group using the `name`. For example:

```terraform
import {
  to = aws_synthetics_group.example
  id = "example"
}
```

Using `terraform import`, import CloudWatch Synthetics Group using the `name`. For example:

```console
% terraform import aws_synthetics_group.example example
```
