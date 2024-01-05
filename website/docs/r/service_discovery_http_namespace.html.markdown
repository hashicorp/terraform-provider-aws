---
subcategory: "Cloud Map"
layout: "aws"
page_title: "AWS: aws_service_discovery_http_namespace"
description: |-
  Provides a Service Discovery HTTP Namespace resource.
---

# Resource: aws_service_discovery_http_namespace

## Example Usage

```terraform
resource "aws_service_discovery_http_namespace" "example" {
  name        = "development"
  description = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the http namespace.
* `description` - (Optional) The description that you specify for the namespace when you create it.
* `tags` - (Optional) A map of tags to assign to the namespace. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of a namespace.
* `arn` - The ARN that Amazon Route 53 assigns to the namespace when you create it.
* `http_name` - The name of an HTTP namespace.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Discovery HTTP Namespace using the namespace ID. For example:

```terraform
import {
  to = aws_service_discovery_http_namespace.example
  id = "ns-1234567890"
}
```

Using `terraform import`, import Service Discovery HTTP Namespace using the namespace ID. For example:

```console
% terraform import aws_service_discovery_http_namespace.example ns-1234567890
```
