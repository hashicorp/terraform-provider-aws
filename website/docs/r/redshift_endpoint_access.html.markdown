---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_endpoint_access"
description: |-
  Provides a Redshift Endpoint Access resource.
---

# Resource: aws_redshift_endpoint_access

Creates a new Amazon Redshift endpoint access.

## Example Usage

```terraform
resource "aws_redshift_endpoint_access" "example" {
  endpoint_name      = "example"
  subnet_group_name  = aws_redshift_subnet_group.example.id
  cluster_identifier = aws_redshift_cluster.example.cluster_identifier
}
```

## Argument Reference

The following arguments are supported:

* `cluster_identifier` - (Required) The cluster identifier of the cluster to access.
* `endpoint_name` - (Required) The Redshift-managed VPC endpoint name.
* `resource_owner` - (Optional) The Amazon Web Services account ID of the owner of the cluster. This is only required if the cluster is in another Amazon Web Services account.
* `subnet_group_name` - (Required) The subnet group from which Amazon Redshift chooses the subnet to deploy the endpoint.
* `vpc_security_group_ids` - (Optional) The security group that defines the ports, protocols, and sources for inbound traffic that you are authorizing into your endpoint.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `address` - The DNS address of the endpoint.
* `id` - The Redshift-managed VPC endpoint name.
* `port` - The port number on which the cluster accepts incoming connections.
* `vpc_endpoint` - The connection endpoint for connecting to an Amazon Redshift cluster through the proxy. See details below.

### VPC Endpoint

* `network_interface` - One or more network interfaces of the endpoint. Also known as an interface endpoint. See details below.
* `vpc_endpoint_id` - The connection endpoint ID for connecting an Amazon Redshift cluster through the proxy.
* `vpc_id` - The VPC identifier that the endpoint is associated.

### Network Interface

* `availability_zone` - The Availability Zone.
* `network_interface_id` - The network interface identifier.
* `private_ip_address` - The IPv4 address of the network interface within the subnet.
* `subnet_id` - The subnet identifier.

## Import

Redshift endpoint access can be imported using the `name`, e.g.,

```
$ terraform import aws_redshift_endpoint_access.example example
```
