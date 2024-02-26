---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_domain_configuration"
description: |-
    Creates and manages an AWS IoT domain configuration.
---

# Resource: aws_iot_domain_configuration

Creates and manages an AWS IoT domain configuration.

## Example Usage

```terraform
resource "aws_iot_domain_configuration" "iot" {
  name         = "iot-"
  domain_name  = "iot.example.com"
  service_type = "DATA"

  server_certificate_arns = [
    aws_acm_certificate.cert.arn
  ]
}
```

## Argument Reference

* `authorizer_config` - (Optional) An object that specifies the authorization service for a domain. See the [`authorizer_config` Block](#authorizer_config-block) below for details.
* `domain_name` - (Optional) Fully-qualified domain name.
* `name` - (Required) The name of the domain configuration. This value must be unique to a region.
* `server_certificate_arns` - (Optional) The ARNs of the certificates that IoT passes to the device during the TLS handshake. Currently you can specify only one certificate ARN. This value is not required for Amazon Web Services-managed domains. When using a custom `domain_name`, the cert must include it.
* `service_type` - (Optional) The type of service delivered by the endpoint. Note: Amazon Web Services IoT Core currently supports only the `DATA` service type.
* `status` - (Optional) The status to which the domain configuration should be set. Valid values are `ENABLED` and `DISABLED`.
* `tags` - (Optional) Map of tags to assign to this resource. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tls_config` - (Optional) An object that specifies the TLS configuration for a domain. See the [`tls_config` Block](#tls_config-block) below for details.
* `validation_certificate_arn` - (Optional) The certificate used to validate the server certificate and prove domain name ownership. This certificate must be signed by a public certificate authority. This value is not required for Amazon Web Services-managed domains.

### `authorizer_config` Block

The `authorizer_config` configuration block supports the following arguments:

* `allow_authorizer_override` - (Optional) A Boolean that specifies whether the domain configuration's authorization service can be overridden.
* `default_authorizer_name` - (Optional) The name of the authorization service for a domain configuration.

### `tls_config` Block

The `tls_config` configuration block supports the following arguments:

* `security_policy` - (Optional) The security policy for a domain configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the domain configuration.
* `domain_type` - The type of the domain.
* `id` - The name of the created domain configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IOT domain configurations using the name. For example:

```terraform
import {
  to = aws_iot_domain_configuration.example
  id = "example"
}
```

Using `terraform import`, import domain configurations using the name. For example:

```console
% terraform import aws_iot_domain_configuration.example example
```
