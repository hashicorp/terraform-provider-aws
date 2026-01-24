---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_asset_type"
description: |-
  Terraform resource for managing an AWS DataZone Asset Type.
---

# Resource: aws_datazone_asset_type

Terraform resource for managing an AWS DataZone Asset Type.

## Example Usage

### Basic Usage

```terraform
resource "aws_datazone_asset_type" "test" {
  description               = "example"
  domain_identifier         = aws_datazone_domain.test.id
  name                      = "example"
  owning_project_identifier = aws_datazone_project.test.id
}
```

## Argument Reference

The following arguments are required:

* `domain_identifier` - (Required) The unique identifier of the Amazon DataZone domain where the custom asset type is being created.
* `name` - (Required) The name of the custom asset type.
* `owning_project_identifier` - (Required) The unique identifier of the Amazon DataZone project that owns the custom asset type.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) The description of the custom asset type.
* `forms_input` - (Optional) The metadata forms that are to be attached to the custom asset type.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - The timestamp when the custom asset type was created.
* `created_by` - The user who created the custom asset type.
* `revision` - The revision of the asset type.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30s`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Asset Type using the `domain_identifier,name`. For example:

```terraform
import {
  to = aws_datazone_asset_type.example
  id = "domain-id-12345678,example"
}
```

Using `terraform import`, import DataZone Asset Type using the `domain_identifier,name`. For example:

```console
% terraform import aws_datazone_asset_type.example domain-id-12345678,example
```
