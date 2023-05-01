---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_endpoint_access"
description: |-
  Provides a Redshift Serverless Endpoint Access resource.
---

# Resource: aws_redshiftserverless_endpoint_access

Creates a new Amazon Redshift Serverless Endpoint Access.

## Example Usage

```terraform
resource "aws_redshiftserverless_endpoint_access" "example" {
  endpoint_name  = "example"
  workgroup_name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `endpoint_name` - (Required) The name of the endpoint.
* `subnet_ids` - (Required) An array of VPC subnet IDs to associate with the endpoint.
* `vpc_security_group_ids` - (Optional) An array of security group IDs to associate with the workgroup.
* `workgroup_name` - (Required) The name of the workgroup.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the Redshift Serverless Endpoint Access.
* `id` - The Redshift Endpoint Access Name.
* `address` - The DNS address of the VPC endpoint.
* `port` - The port that Amazon Redshift Serverless listens on.
* `vpc_endpoint` - The VPC endpoint or the Redshift Serverless workgroup. See `VPC Endpoint` below.

#### VPC Endpoint

* `vpc_endpoint_id` - The DNS address of the VPC endpoint.
* `vpc_id` - The port that Amazon Redshift Serverless listens on.
* `network_interface` - The network interfaces of the endpoint.. See `Network Interface` below.

##### Network Interface

* `availability_zone` - The availability Zone.
* `network_interface_id` - The unique identifier of the network interface.
* `private_ip_address` - The IPv4 address of the network interface within the subnet.
* `subnet_id` - The unique identifier of the subnet.

## Import

Redshift Serverless Endpoint Access can be imported using the `endpoint_name`, e.g.,

```
$ terraform import aws_redshiftserverless_endpoint_access.example example
```
