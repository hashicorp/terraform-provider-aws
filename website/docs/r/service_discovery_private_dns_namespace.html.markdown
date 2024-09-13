---
subcategory: "Cloud Map"
layout: "aws"
page_title: "AWS: aws_service_discovery_private_dns_namespace"
description: |-
  Provides a Service Discovery Private DNS Namespace resource.
---

# Resource: aws_service_discovery_private_dns_namespace

Provides a Service Discovery Private DNS Namespace resource.

## Example Usage

```terraform
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_service_discovery_private_dns_namespace" "example" {
  name        = "hoge.example.local"
  description = "example"
  vpc         = aws_vpc.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the namespace.
* `vpc` - (Required) The ID of VPC that you want to associate the namespace with.
* `description` - (Optional) The description that you specify for the namespace when you create it.
* `tags` - (Optional) A map of tags to assign to the namespace. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of a namespace.
* `arn` - The ARN that Amazon Route 53 assigns to the namespace when you create it.
* `hosted_zone` - The ID for the hosted zone that Amazon Route 53 creates when you create a namespace.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Discovery Private DNS Namespace using the namespace ID and VPC ID. For example:

```terraform
import {
  to = aws_service_discovery_private_dns_namespace.example
  id = "0123456789:vpc-123345"
}
```

Using `terraform import`, import Service Discovery Private DNS Namespace using the namespace ID and VPC ID. For example:

```console
% terraform import aws_service_discovery_private_dns_namespace.example 0123456789:vpc-123345
```
