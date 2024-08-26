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
  count                         = length(data.aws_ec2_transit_gateway_attachments.filtered.ids)
  transit_gateway_attachment_id = data.aws_ec2_transit_gateway_attachments.filtered.ids[count.index]
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.

### filter Argument Reference

* `name` - (Required) Name of the filter check available value on [official documentation][1]
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` A list of all attachments ids matching the filter. You can retrieve more information about the attachment using the [aws_ec2_transit_gateway_attachment][2] data source, searching by identifier.

[1]: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayAttachments.html
[2]: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ec2_transit_gateway_attachment

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
