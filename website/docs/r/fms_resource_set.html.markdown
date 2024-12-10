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

The following arguments are required:

* `resource_set` - (Required) Details about the resource set to be created or updated. See [`resource_set` Attribute Reference](#resource_set-attribute-reference) below.

### `resource_set` Attribute Reference

* `name` - (Required) Descriptive name of the resource set. You can't change the name of a resource set after you create it.
* `resource_type_list` - (Required) Determines the resources that can be associated to the resource set. Depending on your setting for max results and the number of resource sets, a single call might not return the full list.
* `description` - (Optional) Description of the resource set.
* `last_update_time` - (Optional) Last time that the reosurce set was changed.
* `resource_set_status` - (Optional) Indicates whether the resource set is in or out of the admin's Region scope. Valid values are `ACTIVE` (Admin can manage and delete the resource set) or `OUT_OF_ADMIN_SCOPE` (Admin can view the resource set, but theyy can't edit or delete the resource set.)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Resource Set.
* `id` - Unique identifier for the resource set. It's returned in the responses to create and list commands. You provide it to operations like update and delete.

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
