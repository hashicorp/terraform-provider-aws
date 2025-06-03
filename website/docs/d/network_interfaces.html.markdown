---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_interfaces"
description: |-
    Provides a list of network interface ids
---

# Data Source: aws_network_interfaces

## Example Usage

The following shows outputting all network interface ids in a region.

```terraform
data "aws_network_interfaces" "example" {}

output "example" {
  value = data.aws_network_interfaces.example.ids
}
```

The following example retrieves a list of all network interface ids with a custom tag of `Name` set to a value of `test`.

```terraform
data "aws_network_interfaces" "example" {
  tags = {
    Name = "test"
  }
}

output "example1" {
  value = data.aws_network_interfaces.example.ids
}
```

The following example retrieves a network interface ids which associated
with specific subnet.

```terraform
data "aws_network_interfaces" "example" {
  filter {
    name   = "subnet-id"
    values = [aws_subnet.test.id]
  }
}

output "example" {
  value = data.aws_network_interfaces.example.ids
}
```

## Argument Reference

This data source supports the following arguments:

* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired network interfaces.
* `filter` - (Optional) Custom filter block as described below.

### `filter`

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeNetworkInterfaces.html).
* `values` - (Required) Set of values that are accepted for the given field.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `ids` - List of all the network interface ids found.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
