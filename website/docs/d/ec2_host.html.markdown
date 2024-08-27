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
* `host_id` - (Optional) ID of the Dedicated Host.

### filter

This block allows for complex filters. You can use one or more `filter` blocks.

The following arguments are required:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeHosts.html).
* `values` - (Required) Set of values that are accepted for the given field. A host will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the Dedicated Host.
* `arn` - ARN of the Dedicated Host.
* `asset_id` - The ID of the Outpost hardware asset on which the Dedicated Host is allocated.
* `auto_placement` - Whether auto-placement is on or off.
* `availability_zone` - Availability Zone of the Dedicated Host.
* `cores` - Number of cores on the Dedicated Host.
* `host_recovery` - Whether host recovery is enabled or disabled for the Dedicated Host.
* `instance_family` - Instance family supported by the Dedicated Host. For example, "m5".
* `instance_type` - Instance type supported by the Dedicated Host. For example, "m5.large". If the host supports multiple instance types, no instanceType is returned.
* `outpost_arn` - ARN of the AWS Outpost on which the Dedicated Host is allocated.
* `owner_id` - ID of the AWS account that owns the Dedicated Host.
* `sockets` - Number of sockets on the Dedicated Host.
* `total_vcpus` - Total number of vCPUs on the Dedicated Host.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
