---
subcategory: "MemoryDB for Redis"
layout: "aws"
page_title: "AWS: aws_memorydb_subnet_group"
description: |-
  Provides a MemoryDB Subnet Group.
---

# Resource: aws_memorydb_subnet_group

Provides a MemoryDB Subnet Group.

More information about subnet groups can be found in the [MemoryDB User Guide](https://docs.aws.amazon.com/memorydb/latest/devguide/subnetgroups.html).

## Example Usage

```terraform
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  vpc_id            = aws_vpc.example.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_memorydb_subnet_group" "example" {
  name       = "my-subnet-group"
  subnet_ids = [aws_subnet.example.id]
}
```

## Argument Reference

The following arguments are required:

* `subnet_ids` - (Required) Set of VPC Subnet ID-s for the subnet group. At least one subnet must be provided.

The following arguments are optional:

* `name` - (Optional, Forces new resource) Name of the subnet group. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional) Description for the subnet group. Defaults to `"Managed by Terraform"`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the subnet group.
* `arn` - The ARN of the subnet group.
* `vpc_id` - The VPC in which the subnet group exists.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a subnet group using its `name`. For example:

```terraform
import {
  to = aws_memorydb_subnet_group.example
  id = "my-subnet-group"
}
```

Using `terraform import`, import a subnet group using its `name`. For example:

```console
% terraform import aws_memorydb_subnet_group.example my-subnet-group
```
