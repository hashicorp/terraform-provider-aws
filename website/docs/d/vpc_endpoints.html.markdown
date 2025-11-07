---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoints"
description: |-
  Provides information about VPC endpoints.
---

# Data Source: aws_vpc_endpoints

Provides information about VPC endpoints.

## Example Usage

The following shows getting all VPC endpoints.

```terraform
data "aws_vpc_endpoints" "example" {}
```

The following example retrieves a list of all VPC endpoints with a custom tag of `Service` set to a value of `web-service`.

```terraform
data "aws_vpc_endpoints" "example" {
  tags = {
    Service = "web-service"
  }
}
```

The following example retrieves a list of all VPC endpoints in a specific VPC.

```terraform
data "aws_vpc_endpoints" "example" {
  vpc_id = var.vpc_id
}
```

The following example shows filtering by service name and state.

```terraform
data "aws_vpc_endpoints" "s3_endpoints" {
  service_name = "com.amazonaws.us-west-2.s3"
  state        = "available"
}
```

The following example shows filtering by VPC endpoint type.

```terraform
data "aws_vpc_endpoints" "gateway_endpoints" {
  vpc_endpoint_type = "Gateway"
  vpc_id            = var.vpc_id
}
```

## Argument Reference

* `filter` - (Optional) Custom filter block as described below.
* `ip_address_type` - (Optional) IP address type (`ipv4` or `ipv6`).
* `service_name` - (Optional) Service name. For AWS services the service name is usually in the form `com.amazonaws.<region>.<service>` (e.g., `com.amazonaws.us-west-2.s3`).
* `service_region` - (Optional) Region where the service is offered.
* `state` - (Optional) State of the VPC endpoint. Valid values are `pendingAcceptance`, `pending`, `available`, `deleting`, `deleted`, `rejected`, `failed`.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired VPC endpoints.
* `vpc_endpoint_ids` - (Optional) Set of VPC endpoint IDs to retrieve.
* `vpc_endpoint_type` - (Optional) Type of VPC endpoint. Valid values are `Gateway`, `GatewayLoadBalancer`, `Interface`, `Resource`, `ServiceNetwork`.
* `vpc_id` - (Optional) ID of the VPC in which the endpoint resides.

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcEndpoints.html).
* `values` - (Required) Set of values that are accepted for the given field. A VPC endpoint will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of VPC endpoint IDs.
* `vpc_endpoints` - List of VPC endpoints. Each VPC endpoint object contains the following attributes:
    * `arn` - ARN of the VPC endpoint.
    * `cidr_blocks` - List of CIDR blocks for the exposed AWS service. Applicable for endpoints of type Gateway.
    * `dns_entry` - DNS entries for the VPC endpoint. [DNS entry blocks are documented below](#dns_entry-block).
    * `dns_options` - DNS options for the VPC endpoint. [DNS options blocks are documented below](#dns_options-block).
    * `id` - ID of the VPC endpoint.
    * `ip_address_type` - IP address type for the VPC endpoint.
    * `network_interface_ids` - List of network interface IDs associated with the VPC endpoint.
    * `owner_id` - ID of the AWS account that owns the VPC endpoint.
    * `policy` - Policy document associated with the VPC endpoint. Applicable for endpoints of type Gateway.
    * `prefix_list_id` - Prefix list ID of the exposed AWS service. Applicable for endpoints of type Gateway.
    * `private_dns_enabled` - Whether private DNS is enabled for the VPC endpoint.
    * `requester_managed` - Whether the VPC endpoint is being managed by its service.
    * `route_table_ids` - List of route table IDs associated with the VPC endpoint.
    * `security_group_ids` - List of security group IDs associated with the network interfaces.
    * `service_name` - Service name that is specified when creating the VPC endpoint.
    * `state` - State of the VPC endpoint.
    * `subnet_ids` - List of subnet IDs associated with the VPC endpoint.
    * `tags` - Map of tags assigned to the VPC endpoint.
    * `vpc_endpoint_type` - VPC endpoint type, `Gateway`, `GatewayLoadBalancer`, or `Interface`.
    * `vpc_id` - ID of the VPC in which the endpoint is used.

### `dns_entry` Block

DNS blocks (for `dns_entry`) support the following attributes:

* `dns_name` - DNS name.
* `hosted_zone_id` - ID of the private hosted zone.

### `dns_options` Block

DNS options (for `dns_options`) support the following attributes:

* `dns_record_ip_type` - The DNS records created for the endpoint.
* `private_dns_only_for_inbound_resolver_endpoint` - Indicates whether to enable private DNS only for inbound endpoints.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
