---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_peering_attachment_accepter"
description: |-
  Manages the accepter's side of an EC2 Transit Gateway peering Attachment
---

# Resource: aws_ec2_transit_gateway_peering_attachment_accepter

Manages the accepter's side of an EC2 Transit Gateway Peering Attachment.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "example" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.example.id

  tags = {
    Name = "Example cross-account attachment"
  }
}
```

A full example of how to create a Transit Gateway in one AWS account, share it with a second AWS account, and attach a to a Transit Gateway in the second account via the `aws_ec2_transit_gateway_peering_attachment` resource can be found in [the `./examples/transit-gateway-cross-account-peering-attachment` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/transit-gateway-cross-account-peering-attachment).

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `transit_gateway_attachment_id` - (Required) The ID of the EC2 Transit Gateway Peering Attachment to manage.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Peering Attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Attachment identifier
* `transit_gateway_id` - Identifier of EC2 Transit Gateway.
* `peer_transit_gateway_id` - Identifier of EC2 Transit Gateway to peer with.
* `peer_account_id` - Identifier of the AWS account that owns the EC2 TGW peering.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_peering_attachment_accepter` using the EC2 Transit Gateway Attachment identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_peering_attachment_accepter.example
  id = "tgw-attach-12345678"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_peering_attachment_accepter` using the EC2 Transit Gateway Attachment identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_peering_attachment_accepter.example tgw-attach-12345678
```
