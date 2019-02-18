---
layout: "aws"
page_title: "AWS: aws_ram_principal_association"
sidebar_current: "docs-aws-resource-ram-principal-association"
description: |-
  Provides a Resource Access Manager (RAM) principal association.
---

# aws_ram_principal_association

Provides a Resource Access Manager (RAM) principal association.

~> *NOTE:* For an AWS Account ID principal, the target account must accept the RAM association invitation after resource creation.

## Example Usage

### AWS Account ID

```hcl
resource "aws_ram_resource_share" "example" {
  # ... other configuration ...
  allow_external_principals = true
}

resource "aws_ram_principal_association" "example" {
  principal          = "111111111111"
  resource_share_arn = "${aws_ram_resource_share.example.id}"
}
```

### AWS Organization

```hcl
resource "aws_ram_principal_association" "example" {
  principal          = "${aws_organizations_organization.example.id}"
  resource_share_arn = "${aws_ram_resource_share.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `principal` - (Required) The principal to associate with the resource share. Possible values are an AWS account ID, an AWS Organizations Organization ID, or an AWS Organizations Organization Unit ID.
* `resource_share_arn` - (Required) The Amazon Resource Name (ARN) of the resource share.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the Resource Share and the principal, separated by a comma.

## Import

RAM Resource Associations can be imported using their Resource Share ARN and the `principal` separated by a comma, e.g.

```
$ terraform import aws_ram_resource_share.example arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12,123456789012
```
