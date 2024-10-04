---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall"
description: |-
  Retrieve information about an AWS Network Firewall Firewall resource.
---

# Data Source: aws_networkfirewall_firewall

Retrieve information about a firewall.

## Example Usage

### Find firewall policy by ARN

```hcl
data "aws_networkfirewall_firewall" "example" {
  arn = aws_networkfirewall_firewall.arn
}
```

### Find firewall policy by Name

```hcl
data "aws_networkfirewall_firewall" "example" {
  name = "Test"
}
```

### Find firewall policy by ARN and Name

```hcl
data "aws_networkfirewall_firewall" "example" {
  arn  = aws_networkfirewall_firewall.arn
  name = "Test"
}
```

## Argument Reference

One or more of the following arguments are required:

* `arn` - ARN of the firewall.
* `name` - Descriptive name of the firewall.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the firewall.
* `delete_protection` - A flag indicating whether the firewall is protected against deletion.
* `description` - Description of the firewall.
* `encryption_configuration` - AWS Key Management Service (AWS KMS) encryption settings for the firewall.
    * `key_id` - The ID of the AWS Key Management Service (AWS KMS) customer managed key.
    * `type` - The type of the AWS Key Management Service (AWS KMS) key use by the firewall.
* `firewall_policy_arn` - ARN of the VPC Firewall policy.
* `firewall_policy_change_protection` - A flag indicating whether the firewall is protected against a change to the firewall policy association.
* `firewall_status` - Nested list of information about the current status of the firewall.
    * `sync_states` - Set of subnets configured for use by the firewall.
        * `attachment` - Nested list describing the attachment status of the firewall's association with a single VPC subnet.
            * `endpoint_id` - The identifier of the firewall endpoint that AWS Network Firewall has instantiated in the subnet. You use this to identify the firewall endpoint in the VPC route tables, when you redirect the VPC traffic through the endpoint.
            * `subnet_id` - The unique identifier of the subnet that you've specified to be used for a firewall endpoint.
        * `availability_zone` - The Availability Zone where the subnet is configured.
    * `capacity_usage_summary` - Aggregated count of all resources used by reference sets in a firewall.
        * `cidrs` - Capacity usage of CIDR blocks used by IP set references in a firewall.
            * `available_cidr_count` - Available number of CIDR blocks available for use by the IP set references in a firewall.
            * `ip_set_references` - The list of IP set references used by a firewall.
                * `resolved_cidr_count` - Total number of CIDR blocks used by the IP set references in a firewall.
            * `utilized_cidr_count` - Number of CIDR blocks used by the IP set references in a firewall.
    * `configuration_sync_state_summary` - Summary of sync states for all availability zones in which the firewall is configured.
* `id` - ARN that identifies the firewall.
* `name` - Descriptive name of the firewall.
* `subnet_change_protection` - A flag indicating whether the firewall is protected against changes to the subnet associations.
* `subnet_mapping` - Set of configuration blocks describing the public subnets. Each subnet must belong to a different Availability Zone in the VPC. AWS Network Firewall creates a firewall endpoint in each subnet.
    * `subnet_id` - The unique identifier for the subnet.
* `tags` - Map of resource tags to associate with the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `update_token` - String token used when updating a firewall.
* `vpc_id` - Unique identifier of the VPC where AWS Network Firewall should create the firewall.
