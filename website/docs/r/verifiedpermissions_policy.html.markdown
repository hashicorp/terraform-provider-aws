---
subcategory: "Verified Permissions"
layout: "aws"
page_title: "AWS: aws_verifiedpermissions_policy"
description: |-
  Terraform resource for managing an AWS Verified Permissions Policy.
---

# Resource: aws_verifiedpermissions_policy

Terraform resource for managing an AWS Verified Permissions Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_verifiedpermissions_policy" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id

  definition {
    static {
      statement = "permit (principal, action == Action::\"view\", resource in Album:: \"test_album\");"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `policy_store_id` - (Required) The Policy Store ID of the policy store.
* `definition`- (Required) The definition of the policy. See [Definition](#definition) below.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.


### Definition

* `static` - (Optional) The static policy statement. See [Static](#static) below.
* `template_linked` - (Optional) The template linked policy. See [Template Linked](#template_linked) below.

#### Static

* `description` - (Optional) The description of the static policy.
* `statement` - (Required) The statement of the static policy.

#### Template Linked

* `policy_template_id` - (Required) The ID of the template.
* `principal` - (Optional) The principal of the template linked policy.
  * `entity_id` - (Required) The entity ID of the principal.
  * `entity_type` - (Required) The entity type of the principal.
* `resource` - (Optional) The resource of the template linked policy.
  * `entity_id` - (Required) The entity ID of the resource.
  * `entity_type` - (Required) The entity type of the resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Policy. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Verified Permissions Policy using the `policy_id,policy_store_id`. For example:

```terraform
import {
  to = aws_verifiedpermissions_policy.example
  id = "policy-id-12345678,policy-store-id-12345678"
}
```

Using `terraform import`, import Verified Permissions Policy using the `policy_id,policy_store_id`. For example:

```console
% terraform import aws_verifiedpermissions_policy.example policy-id-12345678,policy-store-id-12345678
```
