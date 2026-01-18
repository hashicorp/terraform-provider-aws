---
subcategory: "VPN (Site-to-Site)"
layout: "aws"
page_title: "AWS: aws_vpn_concentrator"
description: |-
  Provides a resource to create a VPN Concentrator.
---

# Resource: aws_vpn_concentrator

Provides a resource to create a VPN Concentrator that aggregates multiple VPN connections to a transit gateway.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway" "example" {
  description = "example"

  tags = {
    Name = "example"
  }
}

resource "aws_vpn_concentrator" "example" {
  type               = "ipsec.1"
  transit_gateway_id = aws_ec2_transit_gateway.example.id

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `type` - (Required) Type of VPN concentrator. Valid value: `ipsec.1`.
* `transit_gateway_id` - (Required) ID of the transit gateway to attach the VPN concentrator to.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `vpn_concentrator_id` - ID of the VPN Concentrator.
* `transit_gateway_attachment_id` - ID of the transit gateway attachment created for the VPN concentrator.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPN Concentrators using the VPN concentrator ID. For example:

```terraform
import {
  to = aws_vpn_concentrator.example
  id = "vcn-12345678"
}
```

Using `terraform import`, import VPN Concentrators using the VPN concentrator ID. For example:

```console
% terraform import aws_vpn_concentrator.example vcn-12345678
```
