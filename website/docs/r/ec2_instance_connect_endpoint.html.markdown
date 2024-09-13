---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_connect_endpoint"
description: |-
  Provides an EC2 Instance Connect Endpoint resource.
---

# Resource: aws_ec2_instance_connect_endpoint

Manages an EC2 Instance Connect Endpoint.

## Example Usage

```terraform
resource "aws_ec2_instance_connect_endpoint" "example" {
  subnet_id = aws_subnet.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `preserve_client_ip` - (Optional) Indicates whether your client's IP address is preserved as the source. Default: `true`.
* `security_group_ids` - (Optional) One or more security groups to associate with the endpoint. If you don't specify a security group, the default security group for the VPC will be associated with the endpoint.
* `subnet_id` - (Required) The ID of the subnet in which to create the EC2 Instance Connect Endpoint.
* `tags` - (Optional) Map of tags to assign to this resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the EC2 Instance Connect Endpoint.
* `availability_zone` - The Availability Zone of the EC2 Instance Connect Endpoint.
* `dns_name` - The DNS name of the EC2 Instance Connect Endpoint.
* `fips_dns_name` - The DNS name of the EC2 Instance Connect FIPS Endpoint.
* `network_interface_ids` - The IDs of the ENIs that Amazon EC2 automatically created when creating the EC2 Instance Connect Endpoint.
* `owner_id` - The ID of the AWS account that created the EC2 Instance Connect Endpoint.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_id` - The ID of the VPC in which the EC2 Instance Connect Endpoint was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Instance Connect Endpoints using the `id`. For example:

```terraform
import {
  to = aws_ec2_instance_connect_endpoint.example
  id = "eice-012345678"
}
```

Using `terraform import`, import EC2 Instance Connect Endpoints using the `id`. For example:

```console
% terraform import aws_ec2_instance_connect_endpoint.example eice-012345678
```
