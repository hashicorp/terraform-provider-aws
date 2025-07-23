---
subcategory: "Resource Explorer"
layout: "aws"
page_title: "AWS: aws_resourceexplorer2_index"
description: |-
  Provides a resource to manage a Resource Explorer index in the current AWS Region.
---

# Resource: aws_resourceexplorer2_index

Provides a resource to manage a Resource Explorer index in the current AWS Region.

## Example Usage

```terraform
resource "aws_resourceexplorer2_index" "example" {
  type = "LOCAL"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `type` - (Required) The type of the index. Valid values: `AGGREGATOR`, `LOCAL`. To understand the difference between `LOCAL` and `AGGREGATOR`, see the [_AWS Resource Explorer User Guide_](https://docs.aws.amazon.com/resource-explorer/latest/userguide/manage-aggregator-region.html).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `2h`)
- `update` - (Default `2h`)
- `delete` - (Default `10m`)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Resource Explorer index.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resource Explorer indexes using the `arn`. For example:

```terraform
import {
  to = aws_resourceexplorer2_index.example
  id = "arn:aws:resource-explorer-2:us-east-1:123456789012:index/6047ac4e-207e-4487-9bcf-cb53bb0ff5cc"
}
```

Using `terraform import`, import Resource Explorer indexes using the `arn`. For example:

```console
% terraform import aws_resourceexplorer2_index.example arn:aws:resource-explorer-2:us-east-1:123456789012:index/6047ac4e-207e-4487-9bcf-cb53bb0ff5cc
```
