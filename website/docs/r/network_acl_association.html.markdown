---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_acl_association"
description: |-
  Provides an network ACL association resource.
---

# Resource: aws_network_acl_association

Provides an network ACL association resource which allows you to associate your network ACL with any subnet(s).

~> **NOTE on Network ACLs and Network ACL Associations:** Terraform provides both a standalone network ACL association resource
and a [network ACL](network_acl.html) resource with a `subnet_ids` attribute. Do not use the same subnet ID in both a network ACL
resource and a network ACL association resource. Doing so will cause a conflict of associations and will overwrite the association.

## Example Usage

```terraform
resource "aws_network_acl_association" "main" {
  network_acl_id = aws_network_acl.main.id
  subnet_id      = aws_subnet.main.id
}
```

## Argument Reference

This resource supports the following arguments:

* `network_acl_id` - (Required) The ID of the network ACL.
* `subnet_id` - (Required) The ID of the associated Subnet.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the network ACL association

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network ACL associations using the `id`. For example:

```terraform
import {
  to = aws_network_acl_association.main
  id = "aclassoc-02baf37f20966b3e6"
}
```

Using `terraform import`, import Network ACL associations using the `id`. For example:

```console
% terraform import aws_network_acl_association.main aclassoc-02baf37f20966b3e6
```
