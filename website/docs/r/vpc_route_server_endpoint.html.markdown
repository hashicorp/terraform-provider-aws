---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_route_server_endpoint"
description: |-
  Terraform resource for managing a VPC (Virtual Private Cloud) Route Server.
---
# Resource: aws_vpc_route_server_endpoint

  Provides a resource for managing a VPC (Virtual Private Cloud) Route Server Endpoint.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc_route_server_endpoint" "test" {
  route_server_id = aws_vpc_route_server.example.route_server_id
  subnet_id       = aws_subnet.main.id

  tags = {
    Name = "Endpoint A"
  }
}
```

## Argument Reference

The following arguments are required:

* `route_server_id` - (Required) The ID of the route server for which to create an endpoint.
* `subnet_id` - (Required) The ID of the subnet in which to create the route server endpoint.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the route server endpoint.
* `route_server_endpoint_id` - The unique identifier of the route server endpoint.
* `eni_id` - The ID of the Elastic network interface for the endpoint.
* `eni_address` - The IP address of the Elastic network interface for the endpoint.
* `vpc_id` - The ID of the VPC containing the endpoint.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC (Virtual Private Cloud) Route Server Endpoint using the `route_server_endpoint_id`. For example:

```terraform
import {
  to = aws_vpc_route_server_endpoint.example
  id = "rse-12345678"
}
```

Using `terraform import`, import VPC (Virtual Private Cloud) Route Server Endpoint using the `route_server_endpoint_id`. For example:

```console
% terraform import aws_vpc_route_server_endpoint.example rse-12345678
```
