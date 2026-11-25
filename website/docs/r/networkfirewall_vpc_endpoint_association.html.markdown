---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_vpc_endpoint_association"
description: |-
  Manages a firewall endpoint for an AWS Network Firewall firewall.
---

# Resource: aws_networkfirewall_vpc_endpoint_association

Manages a firewall endpoint for an AWS Network Firewall firewall.

Use `aws_networkfirewall_vpc_endpoint_association` to establish new firewall endpoints in any Availability Zone where the firewall is already being used. The first use of a firewall in an Availability Zone must be defined by `aws_networkfirewall_firewall` resource and `subnet_mapping` argument.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkfirewall_vpc_endpoint_association" "example" {
  firewall_arn = aws_networkfirewall_firewall.example.arn
  vpc_id       = aws_vpc.example.id

  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }

  tags = {
    Name = "example endpoint"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) Description of the VPC endpoint association.
* `firewall_arn` - (Required) ARN that identifies the firewall.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subnet_mapping` - (Required) ID for a subnet that's used in an association with a firewall. See [`subnet_mapping` Block](#subnet_mapping-block) below for details.
* `tags` - (Optional) Map of resource tags to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_id` - (Required) Unique identifier of the VPC for the endpoint association.

### `subnet_mapping` Block

The `subnet_mapping` block supports the following arguments:

* `ip_address_type` - (Optional) Subnet's IP address type. Valid values: `"DUALSTACK"`, `"IPV4"`.
* `subnet_id` - (Required) Unique identifier for the subnet.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_endpoint_association_arn` - ARN of the VPC Endpoint Association.
* `vpc_endpoint_association_id` - Unique identifier of the VPC endpoint association.
* `vpc_endpoint_association_status` - Nested list of information about the current status of the VPC Endpoint Association.
    * `association_sync_state` - Set of subnets configured for use by the VPC Endpoint Association.
        * `attachment` - Nested list describing the attachment status of the firewall's VPC Endpoint Association with a single VPC subnet.
            * `endpoint_id` - Identifier of the VPC endpoint that AWS Network Firewall has instantiated in the subnet. You use this to identify the firewall endpoint in the VPC route tables, when you redirect the VPC traffic through the endpoint.
            * `status` - Current status of the VPC endpoint association attachment.
            * `status_message` - Message providing additional information about the current status of the attachment.
            * `subnet_id` - Unique identifier of the subnet that you've specified to be used for a VPC Endpoint Association endpoint.
        * `availability_zone` - Availability Zone where the subnet is configured.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall VPC Endpoint Association using the `vpc_endpoint_association_arn`. For example:

```terraform
import {
  to = aws_networkfirewall_vpc_endpoint_association.example
  id = "arn:aws:network-firewall:us-west-1:123456789012:vpc-endpoint-association/example"
}
```

Using `terraform import`, import Network Firewall VPC Endpoint Association using the `vpc_endpoint_association_arn`. For example:

```console
% terraform import aws_networkfirewall_vpc_endpoint_association.example arn:aws:network-firewall:us-west-1:123456789012:vpc-endpoint-association/example
```
