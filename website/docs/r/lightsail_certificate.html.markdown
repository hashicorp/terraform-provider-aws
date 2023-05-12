---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_certificate"
description: |-
  Provides a lightsail certificate
---

# Resource: aws_lightsail_certificate

Provides a lightsail certificate.

## Example Usage

```terraform
resource "aws_lightsail_certificate" "test" {
  name                      = "test"
  domain_name               = "testdomain.com"
  subject_alternative_names = ["www.testdomain.com"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Lightsail load balancer.
* `domain_name` - (Required) A domain name for which the certificate should be issued.
* `subject_alternative_names` - (Optional) Set of domains that should be SANs in the issued certificate. `domain_name` attribute is automatically added as a Subject Alternative Name.
* `tags` - (Optional) A map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the lightsail certificate (matches `name`).
* `arn` - The ARN of the lightsail certificate.
* `created_at` - The timestamp when the instance was created.
* `domain_validation_options` - Set of domain validation objects which can be used to complete certificate validation. Can have more than one element, e.g., if SANs are defined.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.

## Import

`aws_lightsail_certificate` can be imported using the certificate name, e.g.

```shell
$ terraform import aws_lightsail_certificate.test CertificateName
```
