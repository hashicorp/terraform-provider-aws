---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall"
description: |-
  Provides an AWS Network Firewall Firewall resource.
---

# Resource: aws_networkfirewall_firewall

Provides an AWS Network Firewall Firewall Resource

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

### Proxy Firewall

Setting `nat_gateway_create_mapping` creates a proxy firewall. Network Firewall creates one NAT gateway per mapping, and clients reach the firewall through its proxy DNS name (`firewall_status[0].sync_states[*].attachment[0].dns_name`).

Network Firewall creates the NAT gateways with the caller's credentials, so they need EC2 NAT gateway and Elastic IP permissions in addition to Network Firewall permissions. Reading a proxy firewall also needs `ec2:DescribeNatGateways`, planning a proxy firewall or a change to its Availability Zones needs `ec2:DescribeSubnets`, and deleting released NAT gateways needs `ec2:DeleteNatGateway`.

```terraform
resource "aws_networkfirewall_firewall" "example" {
  name                = "example"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.example.arn
  vpc_id              = aws_vpc.example.id

  subnet_mapping {
    subnet_id = aws_subnet.endpoint_a.id
  }

  subnet_mapping {
    subnet_id = aws_subnet.endpoint_b.id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat_a.id
    allocation_ids = [aws_eip.nat_a.id]
    private_ips    = ["10.0.10.10"]
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat_b.id
    allocation_ids = [aws_eip.nat_b.id]
    private_ips    = ["10.0.20.10"]
  }

  proxy_settings {
    listener {
      port = 3128
      type = "HTTP"
    }

    listener {
      port = 8443
      type = "HTTPS"
    }
  }
}
```

### Proxy Firewall NAT Gateways Changed Outside Terraform

Terraform reads `nat_gateway_create_mapping` back from the NAT gateways that Network Firewall created, so it detects changes made outside of Terraform, for example adding or removing an Elastic IP or a private IP on one of these NAT gateways with the EC2 API or the console. The next plan then shows that Availability Zone's mapping as changed, and applying it removes the Availability Zone from the firewall and adds it again with a new NAT gateway. Traffic through that Availability Zone is interrupted while this happens, and its new NAT gateway gets the addresses from the configuration.

To prevent this, either make the change in the configuration instead, or tell Terraform to ignore differences in the NAT gateway mappings:

```terraform
resource "aws_networkfirewall_firewall" "example" {
  # ... other configuration ...

  lifecycle {
    ignore_changes = [nat_gateway_create_mapping]
  }
}
```

~> **Note:** `ignore_changes` also ignores changes you make to `nat_gateway_create_mapping` in the configuration, and Terraform can't ignore a single mapping. Adding or removing an Availability Zone needs a `subnet_mapping` and a `nat_gateway_create_mapping` change together, so remove `nat_gateway_create_mapping` from `ignore_changes` for that apply.

## Argument Reference

This resource supports the following arguments:

* `availability_zone_change_protection` - (Optional) Whether the firewall is protected against changes to its Availability Zone configuration. When set to `true`, you must first disable this protection before adding or removing Availability Zones.
* `availability_zone_mapping` - (Optional) Required when creating a transit gateway-attached firewall. Set of configuration blocks describing the Availability Zones where you want to create firewall endpoints for a transit gateway-attached firewall.
* `delete_associated_nat_gateways` - (Optional) Whether to delete the NAT gateways that Network Firewall created for a proxy firewall when they're released: when the firewall is deleted, or when an update removes or changes an Availability Zone. When `false`, released NAT gateways are kept and keep incurring charges. Not read from AWS; set it and apply before destroying. Ignored for other firewalls. Defaults to `true`.
* `delete_protection` - (Optional) Whether the firewall is protected against deletion. Use this setting to protect against accidentally deleting a firewall that is in use. Defaults to `false`.
* `description` - (Optional) Friendly description of the firewall.
* `enabled_analysis_types` - (Optional) Set of types for which to collect analysis metrics. See [Reporting on network traffic in Network Firewall](https://docs.aws.amazon.com/network-firewall/latest/developerguide/reporting.html) for details on how to use the data. Valid values: `TLS_SNI`, `HTTP_HOST`. Defaults to `[]`.
* `encryption_configuration` - (Optional) KMS encryption configuration settings. See [`encryption_configuration` Block](#encryption_configuration-block) below for details.
* `firewall_policy_arn` - (Required) ARN of the VPC Firewall policy.
* `firewall_policy_change_protection` - (Optional) Whether the firewall is protected against a change to the firewall policy association. Use this setting to protect against accidentally modifying the firewall policy for a firewall that is in use. Defaults to `false`.
* `name` - (Required, Forces new resource) Friendly name of the firewall.
* `nat_gateway_create_mapping` - (Optional) Set of configuration blocks describing the NAT gateways that Network Firewall creates for a proxy firewall, one per Availability Zone. Setting this makes the firewall a proxy firewall. Adding or removing this argument forces a new resource. Each Availability Zone in `subnet_mapping` needs a mapping in the same Availability Zone, in a different subnet. Adding, removing, or changing an Availability Zone (its `subnet_mapping` or `nat_gateway_create_mapping`) is done in place; a changed Availability Zone is removed and then added again, with a new NAT gateway. Terraform reads the mappings from the NAT gateways, so changes made to them outside of Terraform show up as changes to this argument; see [Proxy Firewall NAT Gateways Changed Outside Terraform](#proxy-firewall-nat-gateways-changed-outside-terraform). A proxy firewall must keep at least one Availability Zone, so changing its only Availability Zone forces a new resource, with a new proxy DNS name. The replacement reuses Elastic IPs only if `delete_associated_nat_gateways` is `true`, which frees them when the old firewall is deleted. Conflicts with `transit_gateway_id`. See [`nat_gateway_create_mapping` Block](#nat_gateway_create_mapping-block) below for details.
* `proxy_settings` - (Optional) Required for a proxy firewall. Listener configuration of the firewall's proxy. See [`proxy_settings` Block](#proxy_settings-block) below for details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subnet_change_protection` - (Optional) Whether the firewall is protected against changes to the subnet associations. Use this setting to protect against accidentally modifying the subnet associations for a firewall that is in use. Defaults to `false`.
* `subnet_mapping` - (Optional) Required when creating a VPC attached firewall. Set of configuration blocks describing the public subnets. Each subnet must belong to a different Availability Zone in the VPC. AWS Network Firewall creates a firewall endpoint in each subnet. See [`subnet_mapping` Block](#subnet_mapping-block) below for details.
* `tags` - (Optional) Map of resource tags to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_id` - (Optional, Forces new resource). Required when creating a transit gateway-attached firewall. The unique identifier of the transit gateway to attach to this firewall. You can provide either a transit gateway from your account or one that has been shared with you through AWS Resource Access Manager
* `transparent_filtering_nat_scope` - (Optional, Forces new resource) For a proxy firewall, how traffic that reaches the firewall but isn't sent to the proxy listener leaves it. `NAT_ALL` sends it out through the firewall's NAT gateway; `NAT_NONE` returns it to the endpoint without address translation. Valid values: `NAT_ALL`, `NAT_NONE`. Defaults to `NAT_ALL`. Can only be set with `nat_gateway_create_mapping`.
* `vpc_id` - (Optional, Forces new resource)  Required when creating a VPC attached firewall. The unique identifier of the VPC where AWS Network Firewall should create the firewall.

### `availability_zone_mapping` Block

The `availability_zone_mapping` block supports the following arguments:

* `availability_zone_id` - (Required) ID of the Availability Zone where the firewall endpoint is located.

### `encryption_configuration` Block

`encryption_configuration` settings for customer managed KMS keys. Remove this block to use the default AWS-managed KMS encryption (rather than setting `type` to `AWS_OWNED_KMS_KEY`).

* `key_id` - (Optional) ID of the customer managed key. You can use any of the [key identifiers](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#key-id) that KMS supports, unless you're using a key that's managed by another account. If you're using a key managed by another account, then specify the key ARN.
* `type` - (Required) Type of AWS KMS key to use for encryption of your Network Firewall resources. Valid values are `CUSTOMER_KMS` and `AWS_OWNED_KMS_KEY`.

### `nat_gateway_create_mapping` Block

The `nat_gateway_create_mapping` block supports the following arguments:

* `allocation_ids` - (Optional) List of Elastic IP allocation IDs for a public NAT gateway, up to 8. Omit for a private NAT gateway.
* `private_ips` - (Optional) List of private IPv4 addresses in `subnet_id`, up to 8. For a public NAT gateway, provide exactly one per entry in `allocation_ids`, in the same order. For a private NAT gateway, the first address is the primary address.
* `subnet_id` - (Required) ID of the subnet where Network Firewall creates the NAT gateway. Must be in the firewall's VPC and must not be one of the subnets in `subnet_mapping`.

### `proxy_settings` Block

The `proxy_settings` block supports the following arguments:

* `certificate_arn` - (Optional, Forces new resource) ARN of an ACM certificate, in the same account and Region as the firewall, that the HTTPS listener serves for `domain_name`. Requires `domain_name`.
* `domain_name` - (Optional, Forces new resource) Custom hostname for the proxy. Requires `certificate_arn`. Defaults to a hostname that Network Firewall manages.
* `listener` - (Required) Set of one or two listener configuration blocks: at most one `HTTP` and one `HTTPS` listener, on different ports. See [`listener` Block](#listener-block) below for details.

#### `listener` Block

The `listener` block supports the following arguments:

* `port` - (Required) Port that the listener accepts proxy requests on.
* `type` - (Required) Protocol of the listener. Valid values: `HTTP`, `HTTPS`.

### `subnet_mapping` Block

The `subnet_mapping` block supports the following arguments:

* `ip_address_type` - (Optional) Subnet's IP address type. Valid values: `"DUALSTACK"`, `"IPV4"`.
* `subnet_id` - (Required) Unique identifier for the subnet.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN that identifies the firewall.
* `firewall_status` - Nested list of information about the current status of the firewall.
    * `sync_states` - Set of subnets configured for use by the firewall.
        * `attachment` - Nested list describing the attachment status of the firewall's association with a single VPC subnet.
            * `dns_name` - For a proxy firewall, proxy DNS name that clients use.
            * `endpoint_id` - Identifier of the firewall endpoint that AWS Network Firewall has instantiated in the subnet. You use this to identify the firewall endpoint in the VPC route tables, when you redirect the VPC traffic through the endpoint.
            * `status` - Status of the firewall endpoint.
            * `status_message` - Details about the status, such as the reason for a failure.
            * `subnet_id` - Unique identifier of the subnet that you've specified to be used for a firewall endpoint.
        * `availability_zone` - Availability Zone where the subnet is configured.
        * `nat_gateway_attachment` - For a proxy firewall, list of NAT gateways attached in the Availability Zone.
            * `nat_gateway_id` - ID of the NAT gateway.
            * `status` - Status of the NAT gateway attachment.
            * `status_message` - Details about the status, such as the reason for a failure.
    * `transit_gateway_attachment_sync_states` - Set of transit gateway configured for use by the firewall.
        * `attachment_id` - Unique identifier of the transit gateway attachment.
* `id` - ARN that identifies the firewall.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `transit_gateway_owner_account_id` - AWS account ID that owns the transit gateway.
* `update_token` - String token used when updating a firewall.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `60m`)
- `update` - (Default `60m`)
- `delete` - (Default `60m`)

## Import

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
