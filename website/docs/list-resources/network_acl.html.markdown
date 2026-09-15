---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_acl"
description: |-
  Lists Network ACL resources.
---

# List Resource: aws_network_acl

Lists Network ACL resources.

## Example Usage

### Basic Usage

```terraform
list "aws_network_acl" "example" {
  provider = aws
}
```

### Filter Usage

This example returns Network ACLs associated with a specific VPC.

```terraform
list "aws_network_acl" "example" {
  provider = aws

  config {
    filter {
      name   = "vpc-id"
      values = [aws_vpc.example.id]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) One or more filters to apply to the search.
  If multiple `filter` blocks are provided, they all must be true.
  For a full reference of filter names, see [describe-network-acls in the AWS CLI reference](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-network-acls.html).
  See [`filter` Block](#filter-block) below.
* `network_acl_ids` - (Optional) List of Network ACL IDs to query.
* `region` - (Optional) Region to query. Defaults to provider region.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter.
  For a full reference of filter names, see [describe-network-acls in the AWS CLI reference](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-network-acls.html).
* `values` - (Required) One or more values to match.
