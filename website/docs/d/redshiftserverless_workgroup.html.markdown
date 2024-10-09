---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_workgroup"
description: |-
  Terraform data source for managing an AWS Redshift Serverless Workgroup.
---

# Data Source: aws_redshiftserverless_workgroup

Terraform data source for managing an AWS Redshift Serverless Workgroup.

## Example Usage

### Basic Usage

```terraform
data "aws_redshiftserverless_workgroup" "example" {
  workgroup_name = aws_redshiftserverless_workgroup.example.workgroup_name
}
```

## Argument Reference

The following arguments are required:

* `workgroup_name` - (Required) The name of the workgroup associated with the database.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Redshift Serverless Workgroup.
* `id` - The Redshift Workgroup Name.
* `endpoint` - The endpoint that is created from the workgroup. See `Endpoint` below.
* `enhanced_vpc_routing` - The value that specifies whether to turn on enhanced virtual private cloud (VPC) routing, which forces Amazon Redshift Serverless to route traffic through your VPC instead of over the internet.
* `publicly_accessible` - A value that specifies whether the workgroup can be accessed from a public network.
* `security_group_ids` - An array of security group IDs to associate with the workgroup.
* `subnet_ids` - An array of VPC subnet IDs to associate with the workgroup. When set, must contain at least three subnets spanning three Availability Zones. A minimum number of IP addresses is required and scales with the Base Capacity. For more information, see the following [AWS document](https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-known-issues.html).
* `workgroup_id` - The Redshift Workgroup ID.

### Endpoint

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
