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

This resource supports the following arguments:

* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_id` - (Required) The ID of the VPC to associate with the carrier gateway.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the carrier gateway.
* `id` - The ID of the carrier gateway.
* `owner_id` - The AWS account ID of the owner of the carrier gateway.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_carrier_gateway` using the carrier gateway's ID. For example:

```terraform
import {
  to = aws_ec2_carrier_gateway.example
  id = "cgw-12345"
}
```

Using `terraform import`, import `aws_ec2_carrier_gateway` using the carrier gateway's ID. For example:

```console
% terraform import aws_ec2_carrier_gateway.example cgw-12345
```
