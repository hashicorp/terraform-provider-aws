---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_internet_gateway"
description: |-
    Provides details about a specific Internet Gateway
---

# Data Source: aws_internet_gateway

`aws_internet_gateway` provides details about a specific Internet Gateway.

## Example Usage

```terraform
variable "vpc_id" {}

data "aws_internet_gateway" "default" {
  filter {
    name   = "attachment.vpc-id"
    values = [var.vpc_id]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `internet_gateway_id` - (Optional) ID of the specific Internet Gateway to retrieve.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired Internet Gateway.
* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInternetGateways.html).
* `values` - (Required) Set of values that are accepted for the given field.
  An Internet Gateway will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Internet Gateway.

All of the argument attributes except `filter` block are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected Internet Gateway.

`attachments` are also exported with the following attributes, when there are relevants:
Each attachment supports the following:

* `owner_id` - ID of the AWS account that owns the internet gateway.
* `state` - Current state of the attachment between the gateway and the VPC. Present only if a VPC is attached
* `vpc_id` - ID of an attached VPC.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
