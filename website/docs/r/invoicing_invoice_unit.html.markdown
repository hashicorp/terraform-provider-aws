---
subcategory: "Invoicing"
layout: "aws"
page_title: "AWS: aws_invoicing_invoice_unit"
description: |-
  Manages an AWS Invoice Unit
---

# Resource: aws_invoicing_invoice_unit

Manages an AWS Invoice Unit for organizational billing.

## Example Usage

```terraform
resource "aws_invoicing_invoice_unit" "example" {
  name             = "example-unit"
  description      = "Example invoice unit"
  invoice_receiver = "123456789012"

  rule {
    linked_accounts = ["098765432109"]
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `invoice_receiver` - (Required) AWS account ID that receives invoices for this unit. Cannot be changed after creation.
* `name` - (Required) Unique name of the invoice unit. Cannot be changed after creation.
* `rule` - (Required) Configuration block for invoice unit rules. See [`rule`](#rule) below.

The following arguments are optional:

* `description` - (Optional) Description of the invoice unit.
* `region` - (Optional, **Deprecated**) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tax_inheritance_disabled` - (Optional) Whether tax inheritance is disabled for this invoice unit.

### rule

* `linked_accounts` - (Required) Set of AWS account IDs included in this invoice unit.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the invoice unit.
* `last_modified` - Timestamp when the invoice unit was last modified.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Invoice Units using the ARN. For example:

```terraform
import {
  to = aws_invoicing_invoice_unit.example
  id = "arn:aws:invoicing::123456789012:invoice-unit/example-id"
}
```

Using `terraform import`, import Invoice Units using the ARN. For example:

```console
% terraform import aws_invoicing_invoice_unit.example arn:aws:invoicing::123456789012:invoice-unit/example-id
```
