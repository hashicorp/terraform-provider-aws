---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_domain_name_configuration"
description: |-
    Creates and manages an AWS IoT domain name configuration.
---

# Resource: aws_iot_domain_name_configuration

Creates and manages an AWS IoT domain name configuration.

## Example Usage

```terraform
resource "aws_iot_domain_name_configuration" "iot" {
  name         = "iot-"
  domain_name  = "iot.example.com"
  service_type = "DATA"
  server_certificate_arns = [
    aws_acm_certificate.cert.arn
  ]
}
```


## Argument Reference

* `name` = (Required) The name of the domain configuration. This value must be unique to a region.
* `domain_name` = (Required) Fully-qualified domain name.
* `server_certificate_arns` = (Optional) The ARNs of the certificates that IoT passes to the device during the TLS handshake. Currently you can specify only one certificate ARN. This value is not required for Amazon Web Services-managed domains. When using a custom `domain_name`, the cert must include it.
* `service_type` = (Optional) The type of service delivered by the endpoint. Note: Amazon Web Services IoT Core currently supports only the DATA service type.
* `validation_certificate_arn` = (Optional) The certificate used to validate the server certificate and prove domain name ownership. This certificate must be signed by a public certificate authority. This value is not required for Amazon Web Services-managed domains.
* `authorizer_config` = (Optional) an object that specifies the authorization service for a domain. See Below.
* `tags` = (Optional) Key-value map of resource tags. If configured with a provider default_tags configuration block present, tags with matching keys will overwrite those defined at the provider-level.

### authorizer_config

* `allow_authorizer_override` = (Optional) A Boolean that specifies whether the domain configuration's authorization service can be overridden.
* `default_authorizer_name` = (Optional) The name of the authorization service for a domain configuration.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `name` - The name of the created domain name configuration.
