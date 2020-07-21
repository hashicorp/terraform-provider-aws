---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_prefix_list"
description: |-
    Provides details about a specific prefix list
---

# Data Source: aws_prefix_list

`aws_prefix_list` provides details about a specific AWS prefix list (PL)
or a customer-managed prefix list in the current region.

This can be used both to validate a prefix list given in a variable
and to obtain the CIDR blocks (IP address ranges) for the associated
AWS service. The latter may be useful e.g. for adding network ACL
rules.

## Example Usage

```hcl
resource "aws_vpc_endpoint" "private_s3" {
  vpc_id       = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
}

data "aws_prefix_list" "private_s3" {
  prefix_list_id = "${aws_vpc_endpoint.private_s3.prefix_list_id}"
}

resource "aws_network_acl" "bar" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_network_acl_rule" "private_s3" {
  network_acl_id = "${aws_network_acl.bar.id}"
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "${data.aws_prefix_list.private_s3.cidr_blocks[0]}"
  from_port      = 443
  to_port        = 443
}
```

### Find the regional DynamoDB prefix list

```hcl
data "aws_region" "current" {}
data "aws_prefix_list" "dynamo" {
  name = "com.amazonaws.${data.aws_region.current.name}.dynamodb"
}
```

### Filter

```hcl
data "aws_prefix_list" "test" {
  filter {
    name   = "prefix-list-id"
    values = ["pl-68a54001"]
  }
}
```

### Find a managed prefix list

```hcl
resource "aws_prefix_list" "example" {
  name           = "example"
  max_entries    = 5
  address_family = "IPv4"
  entry {
    cidr_block = "1.0.0.0/8"
  }
  entry {
    cidr_block = "2.0.0.0/8"
  }
  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}

data "aws_prefix_list" "example" {
  prefix_list_id = aws_prefix_list.example.id
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
prefix lists. The given filters must match exactly one prefix list
whose data will be exported as attributes.

* `prefix_list_id` - (Optional) The ID of the prefix list to select.
* `name` - (Optional) The name of the prefix list to select.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) The name of the filter field. Valid values can be found in the EC2 [DescribeManagedPrefixLists](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeManagedPrefixLists.html) API Reference.
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the selected prefix list.
* `arn` - The ARN of the selected prefix list.
* `name` - The name of the selected prefix list.
* `cidr_blocks` - The list of CIDR blocks for the AWS service associated with the prefix list.
* `owner_id` - The Account ID of the owner of a customer-managed prefix list, or `AWS` otherwise.
* `address_family` - The address family of the prefix list. Valid values are `IPv4` and `IPv6`.
* `max_entries` - When then prefix list is managed, the maximum number of entries it supports, or null otherwise.
* `tags` - A map of tags assigned to the resource.
