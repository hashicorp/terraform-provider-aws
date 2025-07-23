---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_acls"
description: |-
    Provides a list of network ACL ids for a VPC
---

# Data Source: aws_network_acls

## Example Usage

The following shows outputting all network ACL ids in a vpc.

```terraform
data "aws_network_acls" "example" {
  vpc_id = var.vpc_id
}

output "example" {
  value = data.aws_network_acls.example.ids
}
```

The following example retrieves a list of all network ACL ids in a VPC with a custom
tag of `Tier` set to a value of "Private".

```terraform
data "aws_network_acls" "example" {
  vpc_id = var.vpc_id

  tags = {
    Tier = "Private"
  }
}
```

The following example retrieves a network ACL id in a VPC which associated
with specific subnet.

```terraform
data "aws_network_acls" "example" {
  vpc_id = var.vpc_id

  filter {
    name   = "association.subnet-id"
    values = [aws_subnet.test.id]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_id` - (Optional) VPC ID that you want to filter from.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired network ACLs.
* `filter` - (Optional) Custom filter block as described below.

### `filter`

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeNetworkAcls.html).
* `values` - (Required) Set of values that are accepted for the given field.
  A VPC will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `ids` - List of all the network ACL ids found.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
