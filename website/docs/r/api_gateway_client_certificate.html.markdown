---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_client_certificate"
description: |-
  Provides an API Gateway Client Certificate.
---

# Resource: aws_api_gateway_client_certificate

Provides an API Gateway Client Certificate.

## Example Usage

```terraform
resource "aws_api_gateway_client_certificate" "demo" {
  description = "My client certificate"
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) Description of the client certificate.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the client certificate.
* `created_date` - Date when the client certificate was created.
* `expiration_date` - Date when the client certificate will expire.
* `pem_encoded_certificate` - The PEM-encoded public key of the client certificate.
* `arn` - ARN
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway Client Certificates using the id. For example:

```terraform
import {
  to = aws_api_gateway_client_certificate.demo
  id = "ab1cqe"
}
```

Using `terraform import`, import API Gateway Client Certificates using the id. For example:

```console
% terraform import aws_api_gateway_client_certificate.demo ab1cqe
```
