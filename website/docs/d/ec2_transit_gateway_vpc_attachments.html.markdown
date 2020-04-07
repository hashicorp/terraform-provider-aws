---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpc_attachments"
description: |-
  Get information on EC2 Transit Gateway VPC Attachments
---

# Data Source: aws_ec2_transit_gateway_vpc_attachments

Get information on EC2 Transit Gateway VPC Attachments.

## Example Usage

### By Filter

```hcl
data "aws_ec2_transit_gateway_vpc_attachments" "example" {
  filter {
    name   = "state"
    values = ["pendingAcceptance"]
  }
}

# to get more information on the attachments
data "aws_ec2_transit_gateway_vpc_attachment" "sample" {
  count = length(data.aws_ec2_transit_gateway_vpc_attachments.example)
  id = data.aws_ec2_transit_gateway_vpc_attachments.example[count.index]
}

```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.

### filter Argument Reference

* `name` - (Required) Name of the filter @see https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-transit-gateway-attachments.html.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `ids` A list of all attachments ids matching filter


