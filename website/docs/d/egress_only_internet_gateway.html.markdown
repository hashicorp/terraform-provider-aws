---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_egress_only_internet_gateway"
description: |-
  Provides details about an Amazon VPC (Virtual Private Cloud) Egress-Only Internet Gateway
---

# Data Source: aws_egress_only_internet_gateway

Provides details about an Amazon VPC (Virtual Private Cloud) Egress-Only Internet Gateway.

## Example Usage

### By ID

```terraform
data "aws_egress_only_internet_gateway" "example" {
  id = "eigw-12345678"
}
```

### By Filter

```terraform
data "aws_egress_only_internet_gateway" "example" {
  filter {
    name   = "tag:Name"
    values = ["example-eigw"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) ID of the specific egress-only internet gateway to retrieve.
* `filter` - (Optional) Custom filter block as described below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `filter` Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeEgressOnlyInternetGateways.html).
* `values` - (Required) Set of values that are accepted for the given field.
  An egress-only Internet gateway will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the egress-only internet gateway.
* `state` - State of the attachment between the gateway and the VPC.
* `tags` - Map of tags assigned to the resource.
* `vpc_id` - ID of the VPC to which the egress-only internet gateway is attached.
