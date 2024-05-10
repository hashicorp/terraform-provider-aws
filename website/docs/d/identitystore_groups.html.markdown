---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_groups"
description: |-
  Terraform data source for managing an AWS SSO Identity Store Groups.
---

# Data Source: aws_identitystore_groups

Terraform data source for managing an AWS SSO Identity Store Groups.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

data "aws_identitystore_groups" "example" {
  identity_store_id = data.aws_ssoadmin_instances.example.identity_store_ids[0]
}
```

## Argument Reference

The following arguments are required:

* `identity_store_id` - (Required) Identity Store ID associated with the Single Sign-On (SSO) Instance.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `groups` - List of Identity Store Groups
    * `group_id` - Identifier of the group in the Identity Store.
    * `description` - Description of the specified group.
    * `display_name` - Group's display name.
    * `external_ids` - List of identifiers issued to this resource by an external identity provider.
        * `id` - Identifier issued to this resource by an external identity provider.
        * `issuer` - Issuer for an external identifier.
