---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_tag"
description: |-
  Manages an individual Organizations resource tag
---

# Resource: aws_organizations_tag

Manages an individual Organizations resource tag. This resource should only be used in cases where Organizations resources are created outside Terraform (e.g., Organizations Accounts implicitly created by AWS Control Tower).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_organizations_account` and `aws_organizations_tag` to manage tags of the same Organizations account will cause a perpetual difference where the `aws_organizations_account` resource will try to remove the tag being added by the `aws_organizations_tag` resource. However, if the parent resource is created in the same configuration (i.e., if you have no other choice), you should add `ignore_changes = [tags]` in the parent resource's lifecycle block. This ensures that Terraform ignores differences in tags managed via the separate tagging resource, avoiding the perpetual difference mentioned above.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```terraform
data "aws_organizations_organization" "example" {}

resource "aws_organizations_organizational_unit" "example" {
  name = "ExampleOU"
  parent_id = data.aws_organizations_organization.example.roots[0].id

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_organizations_tag" "example" {
  resource_id = aws_organizations_organizational_unit.example.id
  key          = "ExampleKey"
  value        = "ExampleValue"
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_id` - (Required) Id of the Organizations resource to tag.
* `key` - (Required) Tag name.
* `value` - (Required) Tag value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Organizations resource identifier and key, separated by a comma (`,`)

## Import

Import `aws_organizations_tag` using the Organizations resource identifier and key, separated by a comma (`,`). For example:

```console
$ terraform import aws_organizations_tag.example ou-1234567,ExampleKey
```
