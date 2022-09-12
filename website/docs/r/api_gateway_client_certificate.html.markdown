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

The following arguments are supported:

* `description` - (Optional) Description of the client certificate.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the client certificate.
* `created_date` - Date when the client certificate was created.
* `expiration_date` - Date when the client certificate will expire.
* `pem_encoded_certificate` - The PEM-encoded public key of the client certificate.
* `arn` - ARN
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

API Gateway Client Certificates can be imported using the id, e.g.,

```
$ terraform import aws_api_gateway_client_certificate.demo ab1cqe
```
