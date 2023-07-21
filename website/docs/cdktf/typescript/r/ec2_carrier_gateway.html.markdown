---
subcategory: "Wavelength"
layout: "aws"
page_title: "AWS: aws_ec2_carrier_gateway"
description: |-
  Manages an EC2 Carrier Gateway.
---

# Resource: aws_ec2_carrier_gateway

Manages an EC2 Carrier Gateway. See the AWS [documentation](https://docs.aws.amazon.com/vpc/latest/userguide/Carrier_Gateway.html) for more information.

## Example Usage

```terraform
resource "aws_ec2_carrier_gateway" "example" {
  vpc_id = aws_vpc.example.id

  tags = {
    Name = "example-carrier-gateway"
  }
}
```

## Argument Reference

The following arguments are supported:

* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`defaultTags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpcId` - (Required) The ID of the VPC to associate with the carrier gateway.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the carrier gateway.
* `id` - The ID of the carrier gateway.
* `ownerId` - The AWS account ID of the owner of the carrier gateway.
* `tagsAll` - A map of tags assigned to the resource, including those inherited from the provider [`defaultTags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

`awsEc2CarrierGateway` can be imported using the carrier gateway's ID,
e.g.,

```
$ terraform import aws_ec2_carrier_gateway.example cgw-12345
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-69fb2719765f13af292ecd34e9026074198e61b364e7f0845c2b65af0b9641db -->