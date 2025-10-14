---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_resource_share"
description: |-
  Manages a Resource Access Manager (RAM) Resource Share.
---

# Resource: aws_ram_resource_share

Manages a Resource Access Manager (RAM) Resource Share. To associate principals with the share, see the [`aws_ram_principal_association` resource](/docs/providers/aws/r/ram_principal_association.html). To associate resources with the share, see the [`aws_ram_resource_association` resource](/docs/providers/aws/r/ram_resource_association.html).

## Example Usage

```terraform
resource "aws_ram_resource_share" "example" {
  name                      = "example"
  allow_external_principals = true

  tags = {
    Environment = "Production"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the resource share.
* `allow_external_principals` - (Optional) Indicates whether principals outside your organization can be associated with a resource share.
* `permission_arns` - (Optional) Specifies the Amazon Resource Names (ARNs) of the RAM permission to associate with the resource share. If you do not specify an ARN for the permission, RAM automatically attaches the default version of the permission for each resource type. You can associate only one permission with each resource type included in the resource share.
* `tags` - (Optional) A map of tags to assign to the resource share. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the resource share.
* `id` - The Amazon Resource Name (ARN) of the resource share.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import resource shares using the `arn` of the resource share. For example:

```terraform
import {
  to = aws_ram_resource_share.example
  id = "arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12"
}
```

Using `terraform import`, import resource shares using the `arn` of the resource share. For example:

```console
% terraform import aws_ram_resource_share.example arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12
```
