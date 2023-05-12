---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_types"
description: |-
  Information about EC2 Instance Types.
---

# Data Source: aws_ec2_instance_types

Information about EC2 Instance Types.

## Example Usage

```terraform
data "aws_ec2_instance_types" "test" {
  filter {
    name   = "auto-recovery-supported"
    values = ["true"]
  }

  filter {
    name   = "network-info.encryption-in-transit-supported"
    values = ["true"]
  }

  filter {
    name   = "instance-storage-supported"
    values = ["true"]
  }

  filter {
    name   = "instance-type"
    values = ["g5.2xlarge", "g5.4xlarge"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstanceTypes.html) for supported filters. Detailed below.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Region.
* `instance_types` - List of EC2 Instance Types.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
