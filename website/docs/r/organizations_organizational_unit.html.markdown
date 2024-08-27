---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_organizational_unit"
description: |-
  Provides a resource to create an organizational unit.
---

# Resource: aws_organizations_organizational_unit

Provides a resource to create an organizational unit.

## Example Usage

```terraform
resource "aws_organizations_organizational_unit" "example" {
  name      = "example"
  parent_id = aws_organizations_organization.example.roots[0].id
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - The name for the organizational unit
* `parent_id` - ID of the parent organizational unit, which may be the root
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `accounts` - List of child accounts for this Organizational Unit. Does not return account information for child Organizational Units. All elements have these attributes:
    * `arn` - ARN of the account
    * `email` - Email of the account
    * `id` - Identifier of the account
    * `name` - Name of the account
* `arn` - ARN of the organizational unit
* `id` - Identifier of the organization unit
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Organizations Organizational Units using the `id`. For example:

```terraform
import {
  to = aws_organizations_organizational_unit.example
  id = "ou-1234567"
}
```

Using `terraform import`, import AWS Organizations Organizational Units using the `id`. For example:

```console
% terraform import aws_organizations_organizational_unit.example ou-1234567
```
