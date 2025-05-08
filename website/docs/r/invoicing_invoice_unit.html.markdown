---
subcategory: "Invoicing"
layout: "aws"
page_title: "AWS: aws_invoicing_invoice_unit"
description: |-
  Creates and manages Invoice Units
---

# Resource: aws_invoicing_invoice_unit

Creates and manages Invoice Units

## Example Usage

```terraform
resource "aws_invoicing_invoice_unit" "example" {
  name                     = "example"
  description              = "example description"
  invoice_receiver         = "12345678901"
  tax_inheritance_disabled = false
  linked_accounts = [
    "12345678903"
  ]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The unique name of the invoice unit that is shown on the generated invoice. This can't be changed once it is set. To change this name, you must delete the invoice unit recreate.
* `invoice_receiver` - (Required) The AWS account ID chosen to be the receiver of an invoice unit. All invoices generated for that invoice unit will be sent to this account ID.
* `linked_accounts` - (Required) List of AWS account IDs where charges are included within the invoice unit.

The following arguments are optional:

* `description` - (Optional) The invoice unit's description. This can be changed at a later time.
* `tax_inheritance_disabled` - (Optional) Whether the invoice unit based tax inheritance is/ should be enabled or disabled.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of the invoice unit.
* `arn` - The ARN of the invoice unit.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Invoice Unit using the ARN. For example:

```terraform
import {
  to = aws_invoicing_invoice_unit.example
  id = "arn:aws:invoicing::1234567890:invoice-unit/zyl7jvks"
}
```

Using `terraform import`, import Invoice Unit using the ARN. For example:

```console
% terraform import aws_invoicing_invoice_unit.example arn:aws:invoicing::1234567890:invoice-unit/zyl7jvks
```
