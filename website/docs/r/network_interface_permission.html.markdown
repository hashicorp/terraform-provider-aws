---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_interface_permission"
description: |-
  Grant cross-account access to an Elastic network interface (ENI).
---

# Resource: aws_network_interface_permission

Grant cross-account access to an Elastic network interface (ENI).

## Example Usage

```terraform
resource "aws_network_interface" "example" {
  subnet_id       = aws_subnet.example.id
  private_ips     = ["10.0.0.50"]
  security_groups = [aws_security_group.example.id]

  attachment {
    instance     = aws_instance.example.id
    device_index = 1
  }
}

resource "aws_network_interface_permission" "example" {
  network_interface_id = aws_network_interface.example.id
  aws_account_id       = "123456789012"
  permission           = "INSTANCE-ATTACH"
}
```

## Argument Reference

The following arguments are required:

* `network_interface_id` - (Required) The ID of the network interface.
* `aws_account_id` - (Required) The Amazon Web Services account ID.
* `permission` - (Required) The type of permission to grant. Valid values are `INSTANCE-ATTACH` or `EIP-ASSOCIATE`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `network_interface_permission_id` - ENI permission ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Interface Permissions using the `network_interface_permission_id`. For example:

```terraform
import {
  to = aws_network_interface_permission.example
  id = "eni-perm-056ad97ce2ac377ed"
}
```

Using `terraform import`, import Network Interface Permissions using the `network_interface_permission_id`. For example:

```console
% terraform import aws_network_interface_permission.example eni-perm-056ad97ce2ac377ed
```
