---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall"
description: |-
  Provides an AWS Network Firewall Firewall resource.
---

# Resource: aws_networkfirewall_firewall

Provides an AWS Network Firewall Firewall Resource.

A firewall has one of two capabilities, chosen when it is created:

* **Routed firewall** - The firewall filters traffic that your VPC route tables send to the firewall endpoints and preserves the client source IP address.
  The firewall is either VPC attached (`vpc_id` and `subnet_mapping`) or transit gateway attached (`transit_gateway_id` and `availability_zone_mapping`).
* **NAT/Proxy firewall** - Network Firewall creates a NAT gateway and a firewall endpoint in each Availability Zone.
  The firewall acts as an explicit forward proxy for HTTP and HTTPS and sends filtered traffic out through the NAT gateways.
  Configure `vpc_id`, `subnet_mapping`, and one `nat_gateway_create_mapping` per Availability Zone.
  See [NAT/Proxy firewall](https://docs.aws.amazon.com/network-firewall/latest/developerguide/nfw-proxy.html) for details.

Changing between the two capabilities forces a new resource to be created.

## Example Usage

```terraform
resource "aws_networkfirewall_firewall" "example" {
  name                   = "example"
  firewall_policy_arn    = aws_networkfirewall_firewall_policy.example.arn
  vpc_id                 = aws_vpc.example.id
  enabled_analysis_types = ["TLS_SNI", "HTTP_HOST"]
  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  timeouts {
    create = "40m"
    update = "50m"
    delete = "1h"
  }
}
```

### NAT/Proxy Firewall

A public NAT gateway needs a route from its subnet to an internet gateway, so the firewall depends on that subnet's route table association.

```terraform
resource "aws_networkfirewall_firewall" "example" {
  name                           = "example"
  firewall_policy_arn            = aws_networkfirewall_firewall_policy.example.arn
  vpc_id                         = aws_vpc.example.id
  delete_associated_nat_gateways = true

  subnet_mapping {
    subnet_id = aws_subnet.firewall.id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat.id
    allocation_ids = [aws_eip.example.allocation_id]
    private_ips    = ["10.0.10.10"]
  }

  proxy_settings {
    listener_property {
      port = 3128
      type = "HTTP"
    }

    listener_property {
      port = 8443
      type = "HTTPS"
    }
  }

  depends_on = [aws_route_table_association.nat]
}
```

### NAT/Proxy Firewall with a Custom Proxy Hostname

```terraform
resource "aws_networkfirewall_firewall" "example" {
  name                            = "example"
  firewall_policy_arn             = aws_networkfirewall_firewall_policy.example.arn
  vpc_id                          = aws_vpc.example.id
  delete_associated_nat_gateways  = true
  transparent_filtering_nat_scope = "NAT_NONE"

  subnet_mapping {
    subnet_id = aws_subnet.firewall.id
  }

  # Private NAT gateway.
  nat_gateway_create_mapping {
    subnet_id   = aws_subnet.nat.id
    private_ips = ["10.0.10.10"]
  }

  proxy_settings {
    certificate_arn = aws_acm_certificate.example.arn
    domain_name     = "proxy.example.com"

    listener_property {
      port = 8443
      type = "HTTPS"
    }
  }
}
```

### Transit Gateway Attached Firewall

```terraform
data "aws_availability_zones" "example" {
  state = "available"
}

resource "aws_networkfirewall_firewall" "example" {
  name                = "example"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.example.arn
  transit_gateway_id  = aws_ec2_transit_gateway.example.id

  availability_zone_mapping {
    availability_zone_id = data.aws_availability_zones.example.zone_ids[0]
  }

  availability_zone_mapping {
    availability_zone_id = data.aws_availability_zones.example.zone_ids[1]
  }
}
```

### Transit Gateway Attached Firewall (Cross Account)

This example creates a Transit Gateway in one AWS account, shares it with a second AWS account using Resource Access Manager (RAM), creates a Network Firewall attached to the Transit Gateway in the second account, and accepts the attachment in the first account with the [`aws_networkfirewall_firewall_transit_gateway_attachment_accepter` resource](/docs/providers/aws/r/networkfirewall_firewall_transit_gateway_attachment_accepter.html).

~> **Note:** This example requires both AWS accounts to be in the same AWS Organization with [RAM sharing with AWS Organizations enabled](https://docs.aws.amazon.com/ram/latest/userguide/getting-started-sharing.html#getting-started-sharing-orgs). Availability Zone IDs are used rather than names because Availability Zone names can map to different physical locations in each account.

```terraform
provider "aws" {
  # Accepter's (Transit Gateway owner's) credentials.
}

provider "aws" {
  alias = "firewall_owner"

  # Firewall owner's credentials.
}

data "aws_caller_identity" "firewall_owner" {
  provider = aws.firewall_owner
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

# Accepter's side: create and share the Transit Gateway.
resource "aws_ec2_transit_gateway" "example" {}

resource "aws_ram_resource_share" "example" {
  name = "example"
}

resource "aws_ram_resource_association" "example" {
  resource_arn       = aws_ec2_transit_gateway.example.arn
  resource_share_arn = aws_ram_resource_share.example.arn
}

resource "aws_ram_principal_association" "example" {
  principal          = data.aws_caller_identity.firewall_owner.account_id
  resource_share_arn = aws_ram_resource_share.example.arn
}

# Firewall owner's side: create the firewall attached to the shared Transit Gateway.
resource "aws_networkfirewall_firewall_policy" "example" {
  provider = aws.firewall_owner

  name = "example"

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_firewall" "example" {
  provider = aws.firewall_owner

  name                = "example"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.example.arn
  transit_gateway_id  = aws_ec2_transit_gateway.example.id

  availability_zone_mapping {
    availability_zone_id = data.aws_availability_zones.available.zone_ids[0]
  }

  # The Transit Gateway must be shared before the firewall can be attached.
  depends_on = [
    aws_ram_principal_association.example,
    aws_ram_resource_association.example,
  ]
}

# Accepter's side: accept the attachment.
resource "aws_networkfirewall_firewall_transit_gateway_attachment_accepter" "example" {
  transit_gateway_attachment_id = aws_networkfirewall_firewall.example.firewall_status[0].transit_gateway_attachment_sync_states[0].attachment_id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `availability_zone_change_protection` - (Optional) A setting indicating whether the firewall is protected against changes to its Availability Zone configuration. When set to `true`, you must first disable this protection before adding or removing Availability Zones.
* `availability_zone_mapping` - (Optional) Required when creating a transit gateway-attached firewall. Set of configuration blocks describing the availability availability where you want to create firewall endpoints for a transit gateway-attached firewall.
* `delete_associated_nat_gateways` - (Optional) Whether to delete the NAT gateways that Network Firewall created for a NAT/Proxy firewall when the firewall is destroyed. If `false`, the NAT gateways are detached from the firewall and keep running as regular NAT gateways, which also prevents their subnets from being destroyed. Only applies to NAT/Proxy firewalls. Defaults to `false`.
* `delete_protection` - (Optional) A flag indicating whether the firewall is protected against deletion. Use this setting to protect against accidentally deleting a firewall that is in use. Defaults to `false`.
* `description` - (Optional) A friendly description of the firewall.
* `enabled_analysis_types` - (Optional) Set of types for which to collect analysis metrics. See [Reporting on network traffic in Network Firewall](https://docs.aws.amazon.com/network-firewall/latest/developerguide/reporting.html) for details on how to use the data. Valid values: `TLS_SNI`, `HTTP_HOST`. Defaults to `[]`.
* `encryption_configuration` - (Optional) KMS encryption configuration settings. See [Encryption Configuration](#encryption-configuration) below for details.
* `firewall_policy_arn` - (Required) ARN of the VPC Firewall policy.
* `firewall_policy_change_protection` - (Optional) A flag indicating whether the firewall is protected against a change to the firewall policy association. Use this setting to protect against accidentally modifying the firewall policy for a firewall that is in use. Defaults to `false`.
* `name` - (Required, Forces new resource) A friendly name of the firewall.
* `nat_gateway_create_mapping` - (Optional) Required when creating a NAT/Proxy firewall. List of configuration blocks describing the NAT gateways that Network Firewall creates, one for each Availability Zone in `subnet_mapping`. Specifying this block creates a NAT/Proxy firewall, which can't be attached to a transit gateway; adding it to or removing it from a routed firewall forces a new resource to be created. A mapping can only be added or removed together with the `subnet_mapping` for its Availability Zone. Removing an Availability Zone detaches its NAT gateway without deleting it, so the NAT gateway keeps running as a regular NAT gateway until you delete it. Changing an existing mapping forces a new resource to be created. Network Firewall doesn't return the mappings, so Terraform reads them from the NAT gateways attached to the firewall, which requires the `ec2:DescribeNatGateways` permission. See [NAT Gateway Create Mapping](#nat-gateway-create-mapping) below for details.
* `proxy_settings` - (Optional) Proxy settings of a NAT/Proxy firewall. If not specified, the firewall listens for HTTP on port `3128` and HTTPS on port `8443`. See [Proxy Settings](#proxy-settings) below for details.
* `subnet_change_protection` - (Optional) A flag indicating whether the firewall is protected against changes to the subnet associations. Use this setting to protect against accidentally modifying the subnet associations for a firewall that is in use. Defaults to `false`.
* `subnet_mapping` - (Optional) Required when creating a VPC attached or NAT/Proxy firewall. Set of configuration blocks describing the public subnets. Each subnet must belong to a different Availability Zone in the VPC. AWS Network Firewall creates a firewall endpoint in each subnet. See [Subnet Mapping](#subnet-mapping) below for details.
* `tags` - (Optional) Map of resource tags to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_id` - (Optional, Forces new resource). Required when creating a transit gateway-attached firewall. The unique identifier of the transit gateway to attach to this firewall. You can provide either a transit gateway from your account or one that has been shared with you through AWS Resource Access Manager
* `transparent_filtering_nat_scope` - (Optional, Forces new resource) How a NAT/Proxy firewall handles traffic that isn't sent to its proxy listener. Valid values: `NAT_ALL` (the firewall sends the traffic out through its NAT gateway after inspection), `NAT_NONE` (the firewall returns the traffic to the firewall endpoint without address translation). Defaults to `NAT_ALL`.
* `vpc_id` - (Optional, Forces new resource) Required when creating a VPC attached or NAT/Proxy firewall. The unique identifier of the VPC where AWS Network Firewall should create the firewall.

### Availability Zone Mapping

The `availability_zone_mapping` block supports the following arguments:

* `availability_zone_id` - (Required)The ID of the Availability Zone where the firewall endpoint is located..

### Encryption Configuration

`encryption_configuration` settings for customer managed KMS keys. Remove this block to use the default AWS-managed KMS encryption (rather than setting `type` to `AWS_OWNED_KMS_KEY`).

* `key_id` - (Optional) The ID of the customer managed key. You can use any of the [key identifiers](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#key-id) that KMS supports, unless you're using a key that's managed by another account. If you're using a key managed by another account, then specify the key ARN.
* `type` - (Required) The type of AWS KMS key to use for encryption of your Network Firewall resources. Valid values are `CUSTOMER_KMS` and `AWS_OWNED_KMS_KEY`.

### NAT Gateway Create Mapping

The `nat_gateway_create_mapping` block supports the following arguments:

* `allocation_ids` - (Optional) List of 1 to 8 Elastic IP address allocation IDs for a public NAT gateway. If not specified, Network Firewall creates a private NAT gateway.
* `private_ips` - (Optional) List of 1 to 8 private IPv4 addresses of the NAT gateway in `subnet_id`. Required for a private NAT gateway, and if `allocation_ids` is specified, with exactly one address for each allocation ID. The first address is the primary address.
* `subnet_id` - (Required) ID of the subnet where Network Firewall creates the NAT gateway. The subnet must be in the firewall's VPC, can't be one of the firewall endpoint subnets, and each mapping must be in a different Availability Zone. For a public NAT gateway, the subnet must have a route to an internet gateway.

### Proxy Settings

The `proxy_settings` block supports the following arguments:

* `certificate_arn` - (Optional, Forces new resource) ARN of an ACM certificate in the same account and Region that the HTTPS proxy listener presents to clients. Its subject alternative names must cover `domain_name`. Required if `domain_name` is specified.
* `domain_name` - (Optional, Forces new resource) Domain name that clients use to reach the proxy. Network Firewall creates the Route 53 private hosted zones that resolve it. Required if `certificate_arn` is specified.
* `listener_property` - (Required) Set of 1 or 2 configuration blocks describing the proxy listeners. Specify at most one HTTP and one HTTPS listener, on different ports. See [Listener Property](#listener-property) below for details.

### Listener Property

The `listener_property` block supports the following arguments:

* `port` - (Required) Port, from `1` to `65535`, that the listener uses.
* `type` - (Required) Type of traffic that the listener handles. Valid values: `HTTP`, `HTTPS`.

### Subnet Mapping

The `subnet_mapping` block supports the following arguments:

* `ip_address_type` - (Optional) The subnet's IP address type. Valid values: `"DUALSTACK"`, `"IPV4"`.
* `subnet_id` - (Required) The unique identifier for the subnet.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ARN that identifies the firewall.
* `arn` - ARN that identifies the firewall.
* `firewall_status` - Nested list of information about the current status of the firewall.
    * `sync_states` - Set of subnets configured for use by the firewall.
        * `attachment` - Nested list describing the attachment status of the firewall's association with a single VPC subnet.
            * `dns_name` - The proxy hostname of a NAT/Proxy firewall in the Availability Zone. This is `proxy_settings.domain_name` if specified, otherwise a name that Network Firewall generates.
            * `endpoint_id` - The identifier of the firewall endpoint that AWS Network Firewall has instantiated in the subnet. You use this to identify the firewall endpoint in the VPC route tables, when you redirect the VPC traffic through the endpoint.
            * `subnet_id` - The unique identifier of the subnet that you've specified to be used for a firewall endpoint.
        * `availability_zone` - The Availability Zone where the subnet is configured.
        * `nat_gateway_attachment` - Nested list of the NAT gateways attached to a NAT/Proxy firewall in the Availability Zone.
            * `nat_gateway_id` - The ID of the NAT gateway.
    * `transit_gateway_attachment_sync_states` - Set of transit gateway configured for use by the firewall.
        * `attachment_id` - The unique identifier of the transit gateway attachment.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `transit_gateway_owner_account_id` - The AWS account ID that owns the transit gateway.
* `update_token` - A string token used when updating a firewall.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `60m`)
- `update` - (Default `60m`)
- `delete` - (Default `60m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networkfirewall_firewall.example
  identity = {
    "arn" = "arn:aws:network-firewall:us-west-1:123456789012:firewall/example"
  }
}

resource "aws_networkfirewall_firewall" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the Network Firewall firewall.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Firewalls using their `arn`. For example:

```terraform
import {
  to = aws_networkfirewall_firewall.example
  id = "arn:aws:network-firewall:us-west-1:123456789012:firewall/example"
}
```

Using `terraform import`, import Network Firewall Firewalls using their `arn`. For example:

```console
% terraform import aws_networkfirewall_firewall.example arn:aws:network-firewall:us-west-1:123456789012:firewall/example
```
