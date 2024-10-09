---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_subnet_group"
description: |-
  Provides an ElastiCache Subnet Group resource.
---

# Resource: aws_elasticache_subnet_group

Provides an ElastiCache Subnet Group resource.

## Example Usage

```terraform
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-test"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-test"
  }
}

resource "aws_elasticache_subnet_group" "bar" {
  name       = "tf-test-cache-subnet"
  subnet_ids = [aws_subnet.foo.id]
}
```

## Argument Reference

This resource supports the following arguments:

* `name` – (Required) Name for the cache subnet group. ElastiCache converts this name to lowercase.
* `description` – (Optional) Description for the cache subnet group. Defaults to "Managed by Terraform".
* `subnet_ids` – (Required) List of VPC Subnet IDs for the cache subnet group
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_id` - The Amazon Virtual Private Cloud identifier (VPC ID) of the cache subnet group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ElastiCache Subnet Groups using the `name`. For example:

```terraform
import {
  to = aws_elasticache_subnet_group.bar
  id = "tf-test-cache-subnet"
}
```

Using `terraform import`, import ElastiCache Subnet Groups using the `name`. For example:

```console
% terraform import aws_elasticache_subnet_group.bar tf-test-cache-subnet
```
