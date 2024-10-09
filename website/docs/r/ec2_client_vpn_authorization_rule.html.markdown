---
subcategory: "VPN (Client)"
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_authorization_rule"
description: |-
  Provides authorization rules for AWS Client VPN endpoints.
---

# Resource: aws_ec2_client_vpn_authorization_rule

Provides authorization rules for AWS Client VPN endpoints. For more information on usage, please see the
[AWS Client VPN Administrator's Guide](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/what-is.html).

## Example Usage

```terraform
resource "aws_ec2_client_vpn_authorization_rule" "example" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.example.id
  target_network_cidr    = aws_subnet.example.cidr_block
  authorize_all_groups   = true
}
```

## Argument Reference

This resource supports the following arguments:

* `client_vpn_endpoint_id` - (Required) The ID of the Client VPN endpoint.
* `target_network_cidr` - (Required) The IPv4 address range, in CIDR notation, of the network to which the authorization rule applies.
* `access_group_id` - (Optional) The ID of the group to which the authorization rule grants access. One of `access_group_id` or `authorize_all_groups` must be set.
* `authorize_all_groups` - (Optional) Indicates whether the authorization rule grants access to all clients. One of `access_group_id` or `authorize_all_groups` must be set.
* `description` - (Optional) A brief description of the authorization rule.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Client VPN authorization rules using the endpoint ID and target network CIDR. If there is a specific group name, include that also. All values are separated by a `,`. For example:

Using the endpoint ID and target network CIDR:

```terraform
import {
  to = aws_ec2_client_vpn_authorization_rule.example
  id = "cvpn-endpoint-0ac3a1abbccddd666,10.1.0.0/24"
}
```

Using the endpoint ID, target network CIDR, and group name:

```terraform
import {
  to = aws_ec2_client_vpn_authorization_rule.example
  id = "cvpn-endpoint-0ac3a1abbccddd666,10.1.0.0/24,team-a"
}
```

**Using `terraform import` to import** AWS Client VPN authorization rules using the endpoint ID and target network CIDR. If there is a specific group name, include that also. All values are separated by a `,`. For example:

Using the endpoint ID and target network CIDR:

```console
% terraform import aws_ec2_client_vpn_authorization_rule.example cvpn-endpoint-0ac3a1abbccddd666,10.1.0.0/24
```

Using the endpoint ID, target network CIDR, and group name:

```console
% terraform import aws_ec2_client_vpn_authorization_rule.example cvpn-endpoint-0ac3a1abbccddd666,10.1.0.0/24,team-a
```
