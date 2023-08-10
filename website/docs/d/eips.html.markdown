---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_eips"
description: |-
    Provides a list of Elastic IPs in a region
---

# Data Source: aws_eips

Provides a list of Elastic IPs in a region.

## Example Usage

The following shows outputting all Elastic IPs with the a specific tag value.

```terraform
data "aws_eips" "example" {
  tags = {
    Env = "dev"
  }
}

# VPC EIPs.
output "allocation_ids" {
  value = data.aws_eips.example.allocation_ids
}

# EC2-Classic EIPs.
output "public_ips" {
  value = data.aws_eips.example.public_ips
}
```

## Argument Reference

* `filter` - (Optional) Custom filter block as described below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired Elastic IPs.

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeAddresses.html).
* `values` - (Required) Set of values that are accepted for the given field. An Elastic IP will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `allocation_ids` - List of all the allocation IDs for address for use with EC2-VPC.
* `public_ips` - List of all the Elastic IP addresses.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
