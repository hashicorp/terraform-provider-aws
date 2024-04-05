---
subcategory: "Verified Permissions"
layout: "aws"
page_title: "AWS: aws_verifiedpermissions_policy_template"
description: |-
  Terraform resource for managing an AWS Verified Permissions Policy Template.
---
# Resource: aws_verifiedpermissions_policy_template

Terraform resource for managing an AWS Verified Permissions Policy Template.

## Example Usage

### Basic Usage

```terraform
resource "aws_verifiedpermissions_policy_template" "example" {
  policy_store_id = aws_verifiedpermissions_policy_store.example.id
  statement       = "permit (principal in ?principal, action in PhotoFlash::Action::\"FullPhotoAccess\", resource == ?resource) unless { resource.IsPrivate };"
}
```

## Argument Reference

The following arguments are required:

* `policy_store_id` - (Required) The ID of the Policy Store.
* `statement` - (Required) Defines the content of the statement, written in Cedar policy language.

The following arguments are optional:

* `description` - (Optional) Provides a description for the policy template.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `policy_template_id` - The ID of the Policy Store.
* `created_date` - The date the Policy Store was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Verified Permissions Policy Store using the `policy_store_id:policy_template_id`. For example:

```terraform
import {
  to = aws_verifiedpermissions_policy_template.example
  id = "DxQg2j8xvXJQ1tQCYNWj9T:X19yzj8xvXJQ1tQCYNWj9T"
}
```

Using `terraform import`, import Verified Permissions Policy Store using the `policy_store_id:policy_template_id`. For example:

```console
% terraform import aws_verifiedpermissions_policy_template.example policyStoreId:policyTemplateId
```
