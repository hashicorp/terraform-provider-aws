---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_prefix_list"
description: |-
    Provides details about a specific prefix list
---

# Data Source: aws_prefix_list

`aws_prefix_list` provides details about a specific AWS prefix list (PL)
in the current region.

This can be used both to validate a prefix list given in a variable
and to obtain the CIDR blocks (IP address ranges) for the associated
AWS service. The latter may be useful e.g., for adding network ACL
rules.

The [aws_ec2_managed_prefix_list](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ec2_managed_prefix_list) data source is normally more appropriate to use given it can return customer-managed prefix list info, as well as additional attributes.

## Example Usage

```terraform
resource "aws_vpc_endpoint" "private_s3" {
  vpc_id       = aws_vpc.foo.id
  service_name = "com.amazonaws.us-west-2.s3"
}

data "aws_prefix_list" "private_s3" {
  prefix_list_id = aws_vpc_endpoint.private_s3.prefix_list_id
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_network_acl_rule" "private_s3" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = data.aws_prefix_list.private_s3.cidr_blocks[0]
  from_port      = 443
  to_port        = 443
}
```

### Filter

```terraform
data "aws_prefix_list" "test" {
  filter {
    name   = "prefix-list-id"
    values = ["pl-68a54001"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `prefix_list_id` - (Optional) ID of the prefix list to select.
* `name` - (Optional) Name of the prefix list to select.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

The arguments of this data source act as filters for querying the available
prefix lists. The given filters must match exactly one prefix list
whose data will be exported as attributes.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [EC2 DescribePrefixLists API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribePrefixLists.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the selected prefix list.
* `name` - Name of the selected prefix list.
* `cidr_blocks` - List of CIDR blocks for the AWS service associated with the prefix list.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
