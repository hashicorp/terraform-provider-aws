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

By arn

```hcl
data "aws_networkfirewall_firewall" "example" {
  arn = aws_networkfirewall_firewall.arn
}
```

By name

```hcl
data "aws_networkfirewall_firewall" "example" {
  name = "Test"
}
```

By Both

```hcl
data "aws_networkfirewall_firewall" "example" {
  arn = aws_networkfirewall_firewall.arn
  name = "Test"
}
```

## Argument Reference

~> **NOTE:** either `arn` or `name` or both are required.

The following arguments are supported:

* `arn` - The Amazon Resource Name (ARN) that identifies the resource policy.
* `name` - A friendly name of the firewall.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) that identifies the resource policy.
* `delete_protection` - A boolean flag indicating whether it is possible to delete the firewall. Defaults to `false`.
* `description` - A friendly description of the firewall.
* `firewall_policy_arn` - The Amazon Resource Name (ARN) of the VPC Firewall policy.
* `firewall_policy_change_protection` - A boolean flag indicating whether it is possible to change the associated firewall policy. Defaults to `false`.
* `firewall_status` - Nested list of information about the current status of the firewall.
    * `sync_states` - Set of subnets configured for use by the firewall.
        * `attachment` - Nested list describing the attachment status of the firewall's association with a single VPC subnet.
            * `endpoint_id` - The identifier of the firewall endpoint that AWS Network Firewall has instantiated in the subnet. You use this to identify the firewall endpoint in the VPC route tables, when you redirect the VPC traffic through the endpoint.
            * `subnet_id` - The unique identifier of the subnet that you've specified to be used for a firewall endpoint.
        * `availability_zone` - The Availability Zone where the subnet is configured.
* `id` - The Amazon Resource Name (ARN) that identifies the firewall.
* `name` - A friendly name of the firewall.
* `subnet_change_protection` - A boolean flag indicating whether it is possible to change the associated subnet(s). Defaults to `false`.
* `subnet_mapping` - Set of configuration blocks describing the public subnets. Each subnet must belong to a different Availability Zone in the VPC. AWS Network Firewall creates a firewall endpoint in each subnet. See [Subnet Mapping](#subnet-mapping) below for details.
* `tags` - Map of resource tags to associate with the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `vpc_id` - The unique identifier of the VPC where AWS Network Firewall should create the firewall.

### Subnet Mapping

The `subnet_mapping` block supports the following arguments:

* `subnet_id` - The unique identifier for the subnet.
* `update_token` - A string token used when updating a firewall.
