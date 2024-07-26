---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb_subnet_group"
description: |-
  Provides an DocumentDB subnet group resource.
---

# Resource: aws_docdb_subnet_group

Provides an DocumentDB subnet group resource.

## Example Usage

```terraform
resource "aws_docdb_subnet_group" "default" {
  name       = "main"
  subnet_ids = [aws_subnet.frontend.id, aws_subnet.backend.id]

  tags = {
    Name = "My docdb subnet group"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional, Forces new resource) The name of the docDB subnet group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional) The description of the docDB subnet group. Defaults to "Managed by Terraform".
* `subnet_ids` - (Required) A list of VPC subnet IDs.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The docDB subnet group name.
* `arn` - The ARN of the docDB subnet group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DocumentDB Subnet groups using the `name`. For example:

```terraform
import {
  to = aws_docdb_subnet_group.default
  id = "production-subnet-group"
}
```

Using `terraform import`, import DocumentDB Subnet groups using the `name`. For example:

```console
% terraform import aws_docdb_subnet_group.default production-subnet-group
```
