---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_local_gateway"
description: |-
    Provides details about a specific Local Gateway
---

# Data Source: aws_vpc

`aws_local_gateway` provides details about a specific Local Gateway.

## Example Usage

The following example shows how one might accept a Local Gateway Id as a variable.

```hcl
variable "local_gateway_id" {}

data "aws_local_gateway" "selected" {
  id = "${var.local_gateway_id}"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
VPCs in the current region. The given filters must match exactly one
VPC whose data will be exported as attributes.

* `filter` - (Optional) Custom filter block as described below.

* `id` - (Optional) The id of the specific VPC to retrieve.

* `state` - (Optional) The current state of the desired VPC.
  Can be either `"pending"` or `"available"`.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired VPC.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLocalGateways.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A VPC will be selected if any one of the given values matches.

## Attributes Reference

All of the argument attributes except `filter` blocks are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected VPC.

The following attribute is additionally exported:

* `outpost_arn` - Amazon Resource Name (ARN) of VPC
* `owner_id` - The ID of the AWS account that owns the VPC.
* `state` - The State of the local gateway.
