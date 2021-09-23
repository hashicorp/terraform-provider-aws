---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_host"
description: |-
  Get information on an EC2 host.
---

# Data Source: aws_ec2_host

Use this data source to get information about the host when allocating an EC2 Dedicated Host.

## Example Usage

```terraform
# Create a new host with instance type of c5.18xlarge with Auto Placement 
# and Host Recovery enabled. 
resource "aws_ec2_host" "test" {
  instance_type     = "c5.18xlarge"
  availability_zone = "us-west-2a"
  host_recovery     = "on"
  auto_placement    = "on"
}

data "aws_ec2_host" "test_data" {
  host_id = aws_ec2_host.test.id
}
```

## Argument Reference

The following arguments are supported:

* `auto_placement` - (Optional) Indicates whether the host accepts any untargeted instance launches that match its instance type configuration, or if it only accepts Host tenancy instance launches that specify its unique host ID.
* `availability_zone` - (Optional) The AZ to start the host in.
* `host_recovery` - (Optional) Indicates whether to enable or disable host recovery for the Dedicated Host. Host recovery is disabled by default.
* `instance_type` - (Optional) Specifies the instance type for which to configure your Dedicated Hosts. When you specify the instance type, that is the only instance type that you can launch onto that host. 

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `host_id` - The host ID. 
* `cores` - The number of cores on the Dedicated Host.
* `instance_family` - The instance family supported by the Dedicated Host. For example, m5.
* `instance_type` -The instance type supported by the Dedicated Host. For example, m5.large. If the host supports multiple instance types, no instanceType is returned.
* `sockets` - The instance family supported by the Dedicated Host. For example, m5.
* `total_vcpus` - The total number of vCPUs on the Dedicated Host.
