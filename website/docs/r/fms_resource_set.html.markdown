---
subcategory: "FMS (Firewall Manager)"
layout: "aws"
page_title: "AWS: aws_fms_resource_set"
description: |-
  Terraform resource for managing an AWS FMS (Firewall Manager) Resource Set.
---

# Resource: aws_fms_resource_set

Terraform resource for managing an AWS FMS (Firewall Manager) Resource Set.

## Example Usage

### Basic Usage

```terraform
resource "aws_fms_resource_set" "example" {
  resource_set {
    name               = "testing"
    resource_type_list = ["AWS::NetworkFirewall::Firewall"]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_set` - (Required) Details about the resource set to be created or updated. See [`resource_set` Block](#resource_set-block) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `resource_set` Block

The following arguments are required:

* `name` - (Required) Descriptive name of the resource set. You can't change the name of a resource set after you create it.

The following arguments are optional:

* `description` - (Optional) Description of the resource set.
* `resource_set_status` - (Optional) Whether the resource set is in or out of the admin's Region scope. Valid values are `ACTIVE` (Admin can manage and delete the resource set) or `OUT_OF_ADMIN_SCOPE` (Admin can view the resource set, but they can't edit or delete the resource set.)
* `resource_type_list` - (Optional) Resources that can be associated to the resource set. Depending on your setting for max results and the number of resource sets, a single call might not return the full list.
* `update_token` - (Optional) Unique identifier for each update to the resource set.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Resource Set.
* `id` - Unique identifier for the resource set. It's returned in the responses to create and list commands. You provide it to operations like update and delete.
* `resource_set` - Details about the resource set. See [`resource_set` Block](#resource_set-block) below.

### `resource_set` Block

* `last_update_time` - Last time that the resource set was changed.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import FMS (Firewall Manager) Resource Set using the `id`. For example:

```terraform
import {
  to = aws_fms_resource_set.example
  id = "resource_set-id-12345678"
}
```

Using `terraform import`, import FMS (Firewall Manager) Resource Set using the `id`. For example:

```console
% terraform import aws_fms_resource_set.example resource_set-id-12345678
```
