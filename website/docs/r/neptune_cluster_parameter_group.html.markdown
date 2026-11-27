---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_cluster_parameter_group"
description: |-
  Manages a Neptune Cluster Parameter Group
---

# Resource: aws_neptune_cluster_parameter_group

Manages a Neptune Cluster Parameter Group

## Example Usage

```terraform
resource "aws_neptune_cluster_parameter_group" "example" {
  family      = "neptune1"
  name        = "example"
  description = "neptune cluster parameter group"

  parameter {
    name  = "neptune_enable_audit_log"
    value = 1
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `description` - (Optional) Description of the neptune cluster parameter group. Defaults to "Managed by Terraform".
* `family` - (Required) Family of the neptune cluster parameter group.
* `name` - (Optional, Forces new resource) Name of the neptune cluster parameter group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `parameter` - (Optional) List of neptune parameters to apply. See [`parameter` Block](#parameter-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### parameter Block

* `apply_method` - (Optional) Valid values are `immediate` and `pending-reboot`. Defaults to `pending-reboot`.
* `name` - (Required) Name of the neptune parameter.
* `value` - (Required) Value of the neptune parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the neptune cluster parameter group.
* `id` - Neptune cluster parameter group name.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Neptune Cluster Parameter Groups using the `name`. For example:

```terraform
import {
  to = aws_neptune_cluster_parameter_group.cluster_pg
  id = "production-pg-1"
}
```

Using `terraform import`, import Neptune Cluster Parameter Groups using the `name`. For example:

```console
% terraform import aws_neptune_cluster_parameter_group.cluster_pg production-pg-1
```
