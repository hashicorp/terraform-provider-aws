---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_proxy"
description: |-
  Manages an AWS Network Firewall Proxy.
---

# Resource: aws_networkfirewall_proxy

Manages an AWS Network Firewall Proxy.

~> **NOTE:** This resource is in preview and may change before general availability.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkfirewall_proxy" "example" {
  name                    = "example"
  nat_gateway_id          = aws_nat_gateway.example.id
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.example.arn

  tls_intercept_properties {
    tls_intercept_mode = "DISABLED"
  }

  listener_properties {
    port = 8080
    type = "HTTP"
  }

  listener_properties {
    port = 443
    type = "HTTPS"
  }
}
```

### With TLS Interception Enabled

```terraform
resource "aws_networkfirewall_proxy" "example" {
  name                    = "example"
  nat_gateway_id          = aws_nat_gateway.example.id
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.example.arn

  tls_intercept_properties {
    tls_intercept_mode = "ENABLED"
    pca_arn            = aws_acmpca_certificate_authority.example.arn
  }

  listener_properties {
    port = 8080
    type = "HTTP"
  }

  listener_properties {
    port = 443
    type = "HTTPS"
  }

  tags = {
    Name = "example"
  }
}
```

### With Logging

```terraform
resource "aws_networkfirewall_proxy" "example" {
  name                    = "example"
  nat_gateway_id          = aws_nat_gateway.example.id
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.example.arn

  tls_intercept_properties {
    tls_intercept_mode = "DISABLED"
  }

  listener_properties {
    port = 8080
    type = "HTTP"
  }

  listener_properties {
    port = 443
    type = "HTTPS"
  }
}

# CloudWatch Logs delivery

resource "aws_cloudwatch_log_group" "example" {
  name              = "example"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_delivery_source" "cwl" {
  name         = "example-cwl"
  log_type     = "ALERT_LOGS"
  resource_arn = aws_networkfirewall_proxy.example.arn
}

resource "aws_cloudwatch_log_delivery_destination" "cwl" {
  name = "example-cwl"

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.example.arn
  }
}

resource "aws_cloudwatch_log_delivery" "cwl" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.cwl.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.cwl.arn
}

# S3 delivery

resource "aws_s3_bucket" "example" {
  bucket        = "example"
  force_destroy = true
}

resource "aws_cloudwatch_log_delivery_source" "s3" {
  name         = "example-s3"
  log_type     = "ALLOW_LOGS"
  resource_arn = aws_networkfirewall_proxy.example.arn
}

resource "aws_cloudwatch_log_delivery_destination" "s3" {
  name = "example-s3"

  delivery_destination_configuration {
    destination_resource_arn = aws_s3_bucket.example.arn
  }
}

resource "aws_cloudwatch_log_delivery" "s3" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.s3.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.s3.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) Descriptive name of the proxy.
* `nat_gateway_id` - (Required, Forces new resource) ID of the NAT gateway to associate with the proxy.
* `tls_intercept_properties` - (Required) TLS interception properties block. See [TLS Intercept Properties](#tls-intercept-properties) below.

The following arguments are optional:

* `listener_properties` - (Optional) One or more listener properties blocks defining the ports and protocols the proxy listens on. See [Listener Properties](#listener-properties) below.
* `proxy_configuration_arn` - (Optional, Forces new resource) ARN of the proxy configuration to use. Exactly one of `proxy_configuration_arn` or `proxy_configuration_name` must be provided.
* `proxy_configuration_name` - (Optional, Forces new resource) Name of the proxy configuration to use. Exactly one of `proxy_configuration_arn` or `proxy_configuration_name` must be provided.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### TLS Intercept Properties

The `tls_intercept_properties` block supports the following:

* `tls_intercept_mode` - (Optional) TLS interception mode. Valid values: `ENABLED`, `DISABLED`.
* `pca_arn` - (Optional) ARN of the AWS Private Certificate Authority (PCA) used for TLS interception. Required when `tls_intercept_mode` is `ENABLED`.

### Listener Properties

Each `listener_properties` block supports the following:

* `port` - (Required) Port number the proxy listens on.
* `type` - (Required) Protocol type. Valid values: `HTTP`, `HTTPS`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Proxy.
* `create_time` - Creation timestamp of the proxy.
* `id` - ARN of the Proxy (deprecated, use `arn`).
* `private_dns_name` - Private DNS name assigned to the proxy.
* `proxy_configuration_arn` - ARN of the proxy configuration (populated when `proxy_configuration_name` is used).
* `proxy_configuration_name` - Name of the proxy configuration (populated when `proxy_configuration_arn` is used).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `update_time` - Last update timestamp of the proxy.
* `update_token` - Token for optimistic locking, required for update operations.
* `vpc_endpoint_service_name` - VPC endpoint service name for the proxy.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `60m`)
* `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Proxy using the `arn`. For example:

```terraform
import {
  to = aws_networkfirewall_proxy.example
  id = "arn:aws:network-firewall:us-west-2:123456789012:proxy/example"
}
```

Using `terraform import`, import Network Firewall Proxy using the `arn`. For example:

```console
% terraform import aws_networkfirewall_proxy.example arn:aws:network-firewall:us-west-2:123456789012:proxy/example
```
