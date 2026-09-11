---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_connections"
description: |-
  Provides details of VPC Endpoint connections to a VPC Endpoint Service.
---

# Data Source: aws_vpc_endpoint_connections

Terraform data source for listing connections to an AWS VPC Endpoint Service.

## Example Usage

### Basic Usage

```terraform
data "aws_vpc_endpoint_connections" "example" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.example.id
}
```

### Listing Pending Connections Using Filter

```terraform
data "aws_vpc_endpoint_connections" "pending" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.example.id

  filter {
    name   = "vpc-endpoint-state"
    values = ["pendingAcceptance"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) Configuration block(s) for filtering. [Filter blocks are documented below](#filter-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_endpoint_service_id` - (Required) ID of the VPC Endpoint Service.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values for `aws_vpc_endpoint_connections` include `service-id`, `vpc-endpoint-owner`, and `vpc-endpoint-state`.
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `connections` - List of VPC Endpoint connections to the service. [Connection blocks are documented below](#connections-block).

### `connections` Block

Connections blocks (for `connections`) support the following attributes:

* `creation_timestamp` - Date and time the VPC endpoint was created.
* `dns_entries` - DNS entries for the VPC endpoint. [DNS entry blocks are documented below](#dns_entries-block).
* `gateway_load_balancer_arns` - ARNs of the Gateway Load Balancers for the service.
* `ip_address_type` - IP address type for the endpoint (`ipv4`, `dualstack`, or `ipv6`).
* `network_load_balancer_arns` - ARNs of the Network Load Balancers for the service.
* `tags` - Map of tags assigned to the VPC endpoint connection.
* `vpc_endpoint_id` - ID of the VPC endpoint.
* `vpc_endpoint_owner` - AWS account ID of the VPC endpoint owner.
* `vpc_endpoint_state` - State of the VPC endpoint (`pendingAcceptance`, `pending`, `available`, `deleting`, `deleted`, `rejected`, `failed`, or `expired`).

### `dns_entries` Block

DNS blocks (for `dns_entries`) support the following attributes:

* `dns_name` - DNS name for the endpoint.
* `hosted_zone_id` - ID of the private hosted zone.
