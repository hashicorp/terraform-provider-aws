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

This resource supports the following arguments:

* `definition` - (Required) Definition of the policy. See [Definition](#definition-block) below.
* `policy_store_id` - (Required) Policy Store ID of the policy store.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `definition` Block

* `static` - (Optional) Static policy statement. See [Static](#static-block) below.
* `template_linked` - (Optional) Template linked policy. See [Template Linked](#template_linked-block) below.

#### `static` Block

* `description` - (Optional) Description of the static policy.
* `statement` - (Required) Statement of the static policy.

#### `template_linked` Block

* `policy_template_id` - (Required) ID of the template.
* `principal` - (Optional) Principal of the template linked policy. See [Principal](#principal-block) below.
* `resource` - (Optional) Resource of the template linked policy. See [Resource](#resource-block) below.

#### `principal` Block

* `entity_id` - (Required) Entity ID of the principal.
* `entity_type` - (Required) Entity type of the principal.

#### `resource` Block

* `entity_id` - (Required) Entity ID of the resource.
* `entity_type` - (Required) Entity type of the resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_date` - Date the policy was created.
* `policy_id` - Policy ID of the policy.

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
