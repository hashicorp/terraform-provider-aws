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
  policy_store_id = aws_verifiedpermissions_policy_store.test.id

  description = ""
  statement   = ""
}
```

## Argument Reference

The following arguments are supported:

* `policy_store_id` - (Required) The ID of the Policy Store.
* `statement` - (Required) Defines the content of the statement, written in Cedar policy language.
* `description` - (Optional) Provides a description for the policy template.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `policy_template_id` - The ID of the Policy Store.
* `created_date` - The date the Policy Store was created.
* `last_updated_date` - The date the Policy Store was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

Verified Permissions Policy Template can be imported using the `policy_store_id:policy_template_id`, e.g.,

```
$ terraform import aws_verifiedpermissions_policy_template.example policyStoreId:policyTemplateId
```
