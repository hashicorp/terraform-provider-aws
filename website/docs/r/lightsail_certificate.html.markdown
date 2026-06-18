---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_certificate"
description: |-
  Manages a Lightsail SSL/TLS certificate for custom domains.
---

# Resource: aws_lightsail_certificate

Manages a Lightsail certificate. Use this resource to create and manage SSL/TLS certificates for securing custom domains with your Lightsail resources.

## Example Usage

```terraform
resource "aws_lightsail_certificate" "example" {
  name                      = "example-certificate"
  domain_name               = "example.com"
  subject_alternative_names = ["www.example.com"]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the certificate.

The following arguments are optional:

* `domain_name` - (Optional) Domain name for which the certificate should be issued.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subject_alternative_names` - (Optional) Set of domains that should be SANs in the issued certificate. `domain_name` attribute is automatically added as a Subject Alternative Name.
* `tags` - (Optional) Map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the certificate.
* `created_at` - Date and time when the certificate was created.
* `domain_validation_options` - Set of domain validation objects which can be used to complete certificate validation. Can have more than one element, e.g., if SANs are defined. Each element contains the following attributes:
    * `domain_name` - Domain name for which the certificate should be issued.
    * `resource_record_name` - Name of the DNS record to create to validate the certificate.
    * `resource_record_type` - Type of DNS record to create to validate the certificate.
    * `resource_record_value` - Value of the DNS record to create to validate the certificate.
* `id` - Name of the certificate (matches `name`).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_certificate` using the certificate name. For example:

```terraform
import {
  to = aws_lightsail_certificate.example
  id = "example-certificate"
}
```

Using `terraform import`, import `aws_lightsail_certificate` using the certificate name. For example:

```console
% terraform import aws_lightsail_certificate.example example-certificate
```
