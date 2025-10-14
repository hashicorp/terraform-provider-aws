---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_principal_association"
description: |-
  Provides a Resource Access Manager (RAM) principal association.
---

# Resource: aws_ram_principal_association

Provides a Resource Access Manager (RAM) principal association. Depending if [RAM Sharing with AWS Organizations is enabled](https://docs.aws.amazon.com/ram/latest/userguide/getting-started-sharing.html#getting-started-sharing-orgs), the RAM behavior with different principal types changes.

When RAM Sharing with AWS Organizations is enabled:

- For AWS Account ID, Organization, and Organizational Unit principals within the same AWS Organization, no resource share invitation is sent and resources become available automatically after creating the association.
- For AWS Account ID principals outside the AWS Organization, a resource share invitation is sent and must be accepted before resources become available. See the [`aws_ram_resource_share_accepter` resource](/docs/providers/aws/r/ram_resource_share_accepter.html) to accept these invitations.

When RAM Sharing with AWS Organizations is not enabled:

- Organization and Organizational Unit principals cannot be used.
- For AWS Account ID principals, a resource share invitation is sent and must be accepted before resources become available. See the [`aws_ram_resource_share_accepter` resource](/docs/providers/aws/r/ram_resource_share_accepter.html) to accept these invitations.

## Example Usage

### AWS Account ID

```terraform
resource "aws_ram_resource_share" "example" {
  # ... other configuration ...
  allow_external_principals = true
}

resource "aws_ram_principal_association" "example" {
  principal          = "111111111111"
  resource_share_arn = aws_ram_resource_share.example.arn
}
```

### AWS Organization

```terraform
resource "aws_ram_principal_association" "example" {
  principal          = aws_organizations_organization.example.arn
  resource_share_arn = aws_ram_resource_share.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `principal` - (Required) The principal to associate with the resource share. Possible values are an AWS account ID, an AWS Organizations Organization ARN, or an AWS Organizations Organization Unit ARN.
* `resource_share_arn` - (Required) The Amazon Resource Name (ARN) of the resource share.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the Resource Share and the principal, separated by a comma.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RAM Principal Associations using their Resource Share ARN and the `principal` separated by a comma. For example:

```terraform
import {
  to = aws_ram_principal_association.example
  id = "arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12,123456789012"
}
```

Using `terraform import`, import RAM Principal Associations using their Resource Share ARN and the `principal` separated by a comma. For example:

```console
% terraform import aws_ram_principal_association.example arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12,123456789012
```
