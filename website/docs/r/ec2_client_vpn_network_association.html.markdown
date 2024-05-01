---
subcategory: "VPN (Client)"
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_network_association"
description: |-
  Provides network associations for AWS Client VPN endpoints.
---

# Resource: aws_ec2_client_vpn_network_association

Provides network associations for AWS Client VPN endpoints. For more information on usage, please see the
[AWS Client VPN Administrator's Guide](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/what-is.html).

## Example Usage

```terraform
resource "aws_ec2_client_vpn_network_association" "example" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.example.id
  subnet_id              = aws_subnet.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `client_vpn_endpoint_id` - (Required) The ID of the Client VPN endpoint.
* `subnet_id` - (Required) The ID of the subnet to associate with the Client VPN endpoint.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique ID of the target network association.
* `association_id` - The unique ID of the target network association.
* `vpc_id` - The ID of the VPC in which the target subnet is located.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Client VPN network associations using the endpoint ID and the association ID. Values are separated by a `,`. For example:

```terraform
import {
  to = aws_ec2_client_vpn_network_association.example
  id = "cvpn-endpoint-0ac3a1abbccddd666,cvpn-assoc-0b8db902465d069ad"
}
```

Using `terraform import`, import AWS Client VPN network associations using the endpoint ID and the association ID. Values are separated by a `,`. For example:

```console
% terraform import aws_ec2_client_vpn_network_association.example cvpn-endpoint-0ac3a1abbccddd666,cvpn-assoc-0b8db902465d069ad
```
