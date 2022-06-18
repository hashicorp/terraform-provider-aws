---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_host"
description: |-
  Get information on an EC2 Host.
---

# Data Source: aws_ec2_host

Use this data source to get information about an EC2 Dedicated Host.

## Example Usage

```terraform
resource "aws_ec2_host" "test" {
  instance_type     = "c5.18xlarge"
  availability_zone = "us-west-2a"
}

data "aws_ec2_host" "test" {
  host_id = aws_ec2_host.test.id
}
```

### Filter Example

```terraform
data "aws_ec2_host" "test" {
  filter {
    name   = "instance-type"
    values = ["c5.18xlarge"]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available EC2 Hosts in the current region.
The given filters must match exactly one host whose data will be exported as attributes.

* `filter` - (Optional) Configuration block. Detailed below.
* `host_id` - (Optional) The ID of the Dedicated Host.

### filter

This block allows for complex filters. You can use one or more `filter` blocks.

The following arguments are required:

* `name` - (Required) The name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeHosts.html).
* `values` - (Required) Set of values that are accepted for the given field. A host will be selected if any one of the given values matches.

## Attributes Reference

In addition to the attributes above, the following attributes are exported:

* `id` - The ID of the Dedicated Host.
* `arn` - The ARN of the Dedicated Host.
* `auto_placement` - Whether auto-placement is on or off.
* `availability_zone` - The Availability Zone of the Dedicated Host.
* `cores` - The number of cores on the Dedicated Host.
* `host_recovery` - Indicates whether host recovery is enabled or disabled for the Dedicated Host.
* `instance_family` - The instance family supported by the Dedicated Host. For example, "m5".
* `instance_type` - The instance type supported by the Dedicated Host. For example, "m5.large". If the host supports multiple instance types, no instanceType is returned.
* `outpost_arn` - The Amazon Resource Name (ARN) of the AWS Outpost on which the Dedicated Host is allocated.
* `owner_id` - The ID of the AWS account that owns the Dedicated Host.
* `sockets` - The number of sockets on the Dedicated Host.
* `total_vcpus` - The total number of vCPUs on the Dedicated Host.
