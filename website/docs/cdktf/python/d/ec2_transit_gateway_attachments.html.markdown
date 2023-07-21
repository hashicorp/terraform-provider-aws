---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_attachments"
description: |-
  Get information on EC2 Transit Gateway Attachments
---

# Data Source: aws_ec2_transit_gateway_attachments

Get information on EC2 Transit Gateway Attachments.

## Example Usage

### By Filter

```hcl
data "aws_ec2_transit_gateway_attachments" "filtered" {
  filter {
    name   = "state"
    values = ["pendingAcceptance"]
  }

  filter {
    name   = "resource-type"
    values = ["vpc"]
  }
}

data "aws_ec2_transit_gateway_attachment" "unit" {
  count = length(data.aws_ec2_transit_gateway_attachments.filtered.ids)
  id    = data.aws_ec2_transit_gateway_attachments.filtered.ids[count.index]
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.

### filter Argument Reference

* `name` - (Required) Name of the filter check available value on [official documentation][1]
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `ids` A list of all attachments ids matching the filter. You can retrieve more information about the attachment using the [aws_ec2_transit_gateway_attachment][2] data source, searching by identifier.

[1]: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayAttachments.html
[2]: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ec2_transit_gateway_attachment

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

<!-- cache-key: cdktf-0.17.0-pre.15 input-5aab4b9996d301d85a81444ffb46bbaa0a94e492c4d5d3b1c1df96c4c03b5905 -->