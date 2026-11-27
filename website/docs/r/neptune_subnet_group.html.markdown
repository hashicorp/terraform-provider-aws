---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_subnet_group"
description: |-
  Provides an Neptune subnet group resource.
---

# Resource: aws_neptune_subnet_group

Provides an Neptune subnet group resource.

## Example Usage

```terraform
resource "aws_neptune_subnet_group" "default" {
  name       = "main"
  subnet_ids = [aws_subnet.frontend.id, aws_subnet.backend.id]

  tags = {
    Name = "My neptune subnet group"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `description` - (Optional) Description of the neptune subnet group. Defaults to "Managed by Terraform".
* `name` - (Optional, Forces new resource) Name of the neptune subnet group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subnet_ids` - (Required) List of VPC subnet IDs.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the neptune subnet group.
* `id` - Neptune subnet group name.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Neptune Subnet groups using the `name`. For example:

```terraform
import {
  to = aws_neptune_subnet_group.default
  id = "production-subnet-group"
}
```

Using `terraform import`, import Neptune Subnet groups using the `name`. For example:

```console
% terraform import aws_neptune_subnet_group.default production-subnet-group
```
