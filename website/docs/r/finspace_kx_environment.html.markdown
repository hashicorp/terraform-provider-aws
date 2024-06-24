---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_environment"
description: |-
  Terraform resource for managing an AWS FinSpace Kx Environment.
---

# Resource: aws_finspace_kx_environment

Terraform resource for managing an AWS FinSpace Kx Environment.

## Example Usage

### Basic Usage

```terraform
resource "aws_kms_key" "example" {
  description             = "Sample KMS Key"
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "example" {
  name       = "my-tf-kx-environment"
  kms_key_id = aws_kms_key.example.arn
}
```

### With Transit Gateway Configuration

```terraform
resource "aws_kms_key" "example" {
  description             = "Sample KMS Key"
  deletion_window_in_days = 7
}

resource "aws_ec2_transit_gateway" "example" {
  description = "example"
}

resource "aws_finspace_kx_environment" "example_env" {
  name        = "my-tf-kx-environment"
  description = "Environment description"
  kms_key_id  = aws_kms_key.example.arn

  transit_gateway_configuration {
    transit_gateway_id  = aws_ec2_transit_gateway.example.id
    routable_cidr_space = "100.64.0.0/26"
  }

  custom_dns_configuration {
    custom_dns_server_name = "example.finspace.amazonaws.com"
    custom_dns_server_ip   = "10.0.0.76"
  }
}
```

### With Transit Gateway Attachment Network ACL Configuration

```terraform
resource "aws_kms_key" "example" {
  description             = "Sample KMS Key"
  deletion_window_in_days = 7
}

resource "aws_ec2_transit_gateway" "example" {
  description = "example"
}

resource "aws_finspace_kx_environment" "example_env" {
  name        = "my-tf-kx-environment"
  description = "Environment description"
  kms_key_id  = aws_kms_key.example.arn

  transit_gateway_configuration {
    transit_gateway_id  = aws_ec2_transit_gateway.example.id
    routable_cidr_space = "100.64.0.0/26"
    attachment_network_acl_configuration {
      rule_number = 1
      protocol    = "6"
      rule_action = "allow"
      cidr_block  = "0.0.0.0/0"
      port_range {
        from = 53
        to   = 53
      }
      icmp_type_code {
        type = -1
        code = -1
      }
    }
  }

  custom_dns_configuration {
    custom_dns_server_name = "example.finspace.amazonaws.com"
    custom_dns_server_ip   = "10.0.0.76"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the KX environment that you want to create.
* `kms_key_id` - (Required) KMS key ID to encrypt your data in the FinSpace environment.

The following arguments are optional:

* `custom_dns_configuration` - (Optional) List of DNS server name and server IP. This is used to set up Route-53 outbound resolvers. Defined below.
* `description` - (Optional) Description for the KX environment.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_configuration` - (Optional) Transit gateway and network configuration that is used to connect the KX environment to an internal network. Defined below.

### custom_dns_configuration

The custom_dns_configuration block supports the following arguments:

* `custom_dns_server_ip` - (Required) IP address of the DNS server.
* `custom_dns_server_name` - (Required) Name of the DNS server.

### transit_gateway_configuration

The transit_gateway_configuration block supports the following arguments:

* `routable_cidr_space` - (Required) Routing CIDR on behalf of KX environment. It could be any “/26 range in the 100.64.0.0 CIDR space. After providing, it will be added to the customer’s transit gateway routing table so that the traffics could be routed to KX network.
* `transit_gateway_id` - (Required) Identifier of the transit gateway created by the customer to connect outbound traffics from KX network to your internal network.
* `attachment_network_acl_configuration` - (Optional) Rules that define how you manage outbound traffic from kdb network to your internal network. Defined below.

### attachment_network_acl_configuration

The network access control list (ACL) is an optional layer of security for VPCs that acts as a firewall for controlling traffic in and out of one or more subnets.
The entry is a set of numbered ingress and egress rules that determine whether a packet should be allowed in or out of a subnet associated with the ACL.
Entries in the ACL are processed according to the rule numbers, in ascending order. The `attachment_network_acl_configuration` block supports the following arguments:

* `cidr_block` - (Required) The IPv4 network range to allow or deny, in CIDR notation. The specified CIDR block is modified to its canonical form. For example, `100.68.0.18/18` will be converted to `100.68.0.0/18`.
* `protocol` - (Required) Protocol number. A value of `1` means all the protocols.
* `rule_action` - (Required) Indicates whether to `allow` or `deny` the traffic that matches the rule.
* `rule_number` - (Required) Rule number for the entry. All the network ACL entries are processed in ascending order by rule number.
* `icmp_type_code` - (Optional) Defines the ICMP protocol that consists of the ICMP type and code. Defined below.
* `port_range` - (Optional) Range of ports the rule applies to. Defined below.

### port_range

The range of ports the rule applies to (between `0` and `65535`). The `port_range` block supports the following arguments:

* `from` - (Required) First port in the range.
* `to` - (Required) Last port in the range.

### icmp_type_code

Defines the ICMP protocol that consists of the ICMP type and code. The `icmp_type_code` block supports the following arguments:

* `code` - (Required) ICMP code. A value of `-1` means all codes for the specified ICMP type.
* `type` - (Required) ICMP type. A value of `-1` means all types.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifier of the KX environment.
* `availability_zones` - AWS Availability Zone IDs that this environment is available in. Important when selecting VPC subnets to use in cluster creation.
* `created_timestamp` - Timestamp at which the environment is created in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `id` - Unique identifier for the KX environment.
* `infrastructure_account_id` - Unique identifier for the AWS environment infrastructure account.
* `last_modified_timestamp` - Last timestamp at which the environment was updated in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `status` - Status of environment creation
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `75m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS FinSpace Kx Environment using the `id`. For example:

```terraform
import {
  to = aws_finspace_kx_environment.example
  id = "n3ceo7wqxoxcti5tujqwzs"
}
```

Using `terraform import`, import an AWS FinSpace Kx Environment using the `id`. For example:

```console
% terraform import aws_finspace_kx_environment.example n3ceo7wqxoxcti5tujqwzs
```
