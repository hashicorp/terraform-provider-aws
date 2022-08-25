---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_attachment"
description: |-
  Get information on an EC2 Transit Gateway's attachment to a resource
---

# Data Source: aws_ec2_transit_gateway_attachment

Get information on an EC2 Transit Gateway's attachment to a resource.

## Example Usage

```terraform
data "aws_ec2_transit_gateway_attachment" "example" {
  filter {
    name   = "transit-gateway-id"
    values = [aws_ec2_transit_gateway.example.id]
  }

  filter {
    name   = "resource-type"
    values = ["peering"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `transit_gateway_attachment_id` - (Optional) The ID of the attachment.

### filter Argument Reference

* `name` - (Required) The name of the field to filter by, as defined by the [underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayAttachments.html).
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the attachment.
* `resource_id` - The ID of the resource.
* `resource_owner_id` - The ID of the AWS account that owns the resource.
* `resource_type` - The resource type.
* `state` - The attachment state.
* `tags` - Key-value tags for the attachment.
* `transit_gateway_id` - The ID of the transit gateway.
* `transit_gateway_owner_id` - The ID of the AWS account that owns the transit gateway.
