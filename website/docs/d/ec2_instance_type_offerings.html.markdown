---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_type_offerings"
description: |-
  Information about EC2 Instance Type Offerings.
---

# Data Source: aws_ec2_instance_type_offerings

Information about EC2 Instance Type Offerings.

## Example Usage

```terraform
data "aws_ec2_instance_type_offerings" "example" {
  filter {
    name   = "instance-type"
    values = ["t2.micro", "t3.micro"]
  }

  filter {
    name   = "location"
    values = ["usw2-az4"]
  }

  location_type = "availability-zone-id"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstanceTypeOfferings.html) for supported filters. Detailed below.
* `location_type` - (Optional) Location type. Defaults to `region`. Valid values: `availability-zone`, `availability-zone-id`, and `region`.

### filter Argument Reference

* `name` - (Required) Name of the filter. The `location` filter depends on the top-level `location_type` argument and if not specified, defaults to the current region.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `instance_types` - List of EC2 Instance Types.
* `locations` - List of locations.
* `location_types` - List of location types.

Note that the indexes of Instance Type Offering instance types, locations and location types correspond.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
