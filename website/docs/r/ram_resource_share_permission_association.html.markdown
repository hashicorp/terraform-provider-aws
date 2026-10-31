---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_resource_share_permission_association"
description: |-
  Manages an AWS RAM (Resource Access Manager) Resource Share Permission Association.
---

# Resource: aws_ram_resource_share_permission_association

Manages an AWS RAM (Resource Access Manager) Resource Share Permission Association.

When a custom RAM permission policy is updated, AWS creates a new version of that permission. This resource ensures the resource share is always associated with the latest version of the permission. Without this resource, updating a permission policy would create a new version in AWS but the resource share would remain on the old version with no way to update it via Terraform.

To create a RAM resource share, see the [`aws_ram_resource_share` resource](/docs/providers/aws/r/ram_resource_share.html). To create a custom RAM permission, see the [`aws_ram_permission` resource](/docs/providers/aws/r/ram_permission.html). To associate principals with the share, see the [`aws_ram_principal_association` resource](/docs/providers/aws/r/ram_principal_association.html). To associate resources with the share, see the [`aws_ram_resource_association` resource](/docs/providers/aws/r/ram_resource_association.html).

~> **NOTE:** This resource always uses the current default version of the permission when associating with a resource share. The `permission_version` attribute is read-only and reflects the version AWS has attached. If you need to pin to a specific version, you must first set that version as the default in AWS before applying.

## Example Usage

### Basic Usage

```terraform
resource "aws_ram_permission" "example" {
  name          = "example"
  resource_type = "route53profiles:Profile"
  policy_template = jsonencode({
    Effect = "Allow"
    Action = [
      "route53profiles:GetProfile",
      "route53profiles:GetProfileResourceAssociation",
      "route53profiles:ListProfileResourceAssociations",
    ]
  })
}

resource "aws_ram_resource_share" "example" {
  name                      = "example"
  allow_external_principals = false
}

resource "aws_ram_resource_share_permission_association" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn
  permission_arn     = aws_ram_permission.example.arn
}

output "permission_version" {
  value = aws_ram_resource_share_permission_association.example.permission_version
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `permission_arn` - (Required) Amazon Resource Name (ARN) of the AWS RAM permission to associate with the resource share.
* `resource_share_arn` - (Required) Amazon Resource Name (ARN) of the resource share.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `permission_version` - Version of the AWS RAM permission currently associated with the resource share.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RAM Resource Share Permission Association using the `resource_share_arn` and `permission_arn` separated by a comma. For example:

```terraform
import {
  to = aws_ram_resource_share_permission_association.example
  id = "arn:aws:ram:us-west-2:123456789012:resource-share/example,arn:aws:ram::aws:permission/example"
}
```

Using `terraform import`, import RAM Resource Share Permission Association using the `resource_share_arn` and `permission_arn` separated by a comma. For example:

```console
% terraform import aws_ram_resource_share_permission_association.example arn:aws:ram:us-west-2:123456789012:resource-share/example,arn:aws:ram::aws:permission/example
```