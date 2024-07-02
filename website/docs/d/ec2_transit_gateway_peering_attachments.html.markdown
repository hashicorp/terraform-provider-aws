---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_peering_attachments"
description: |-
  Get information on EC2 Transit Gateway Peering Attachments
---

# Data Source: aws_ec2_transit_gateway_peering_attachments

Get information on EC2 Transit Gateway Peering Attachments.

## Example Usage

### All Resources

```hcl
data "aws_ec2_transit_gateway_peering_attachments" "test" {}
```

### By Filter

```hcl
data "aws_ec2_transit_gateway_peering_attachments" "filtered" {
  filter {
    name   = "state"
    values = ["pendingAcceptance"]
  }
}

data "aws_ec2_transit_gateway_peering_attachment" "unit" {
  count = length(data.aws_ec2_transit_gateway_peering_attachments.filtered.ids)
  id    = data.aws_ec2_transit_gateway_peering_attachments.filtered.ids[count.index]
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.

### filter Argument Reference

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API][1]
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` A list of all attachments ids matching the filter. You can retrieve more information about the attachment using the [aws_ec2_transit_gateway_peering_attachment][2] data source, searching by identifier.

[1]: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayPeeringAttachments.html
[2]: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ec2_transit_gateway_peering_attachment

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
