---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_ingress_point"
description: |-
  Manages an AWS SES Mail Manager Ingress Point.
---

# Resource: aws_mailmanager_ingress_point

Manages an AWS SES Mail Manager Ingress Point.

## Example Usage

### Basic Usage

```terraform
resource "aws_mailmanager_ingress_point" "example" {
  name              = "example"
  type              = "OPEN"
  rule_set_id       = aws_mailmanager_rule_set.example.id
  traffic_policy_id = aws_mailmanager_traffic_policy.example.id
}
```

### Authenticated Ingress Point with SMTP Password

```terraform
resource "aws_mailmanager_ingress_point" "example" {
  name              = "example"
  type              = "AUTH"
  rule_set_id       = aws_mailmanager_rule_set.example.id
  traffic_policy_id = aws_mailmanager_traffic_policy.example.id

  ingress_point_configuration {
    smtp_password_wo         = var.smtp_password
    smtp_password_wo_version = 1
  }
}
```

### Private Network Configuration

```terraform
resource "aws_mailmanager_ingress_point" "example" {
  name              = "example"
  type              = "OPEN"
  rule_set_id       = aws_mailmanager_rule_set.example.id
  traffic_policy_id = aws_mailmanager_traffic_policy.example.id

  network_configuration {
    private_network_configuration {
      vpc_endpoint_id = aws_vpc_endpoint.example.id
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the ingress point.
* `rule_set_id` - (Required) Identifier of the rule set applied to the ingress point.
* `traffic_policy_id` - (Required) Identifier of the traffic policy applied to the ingress point.
* `type` - (Required) Type of the ingress point. Valid values are `OPEN`, `AUTH`, and `MTLS`. Changing this value forces a new resource.

The following arguments are optional:

* `ingress_point_configuration` - (Optional) Configuration used to authenticate with the ingress point. See [`ingress_point_configuration` Block](#ingress_point_configuration-block) for details.
* `network_configuration` - (Optional) Network configuration for the ingress point. See [`network_configuration` Block](#network_configuration-block) for details. Changing this value forces a new resource.
* `region` - (Optional) Region where this resource is managed.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tls_policy` - (Optional) TLS policy for the ingress point. Valid values are `REQUIRED`, `OPTIONAL`, and `FIPS`.

### `ingress_point_configuration` Block

The `ingress_point_configuration` block supports the following:

* `secret_arn` - (Optional) ARN of the secret in AWS Secrets Manager that holds the SMTP password, used for `AUTH` ingress points.
* `smtp_password_wo` - (Optional, Write-Only) SMTP password used for `AUTH` ingress points. This argument is not stored in state. Requires `smtp_password_wo_version` to be set. See [Write-Only Arguments](https://developer.hashicorp.com/terraform/language/resources/syntax#write-only-arguments) for more information.
* `smtp_password_wo_version` - (Optional) Version number for `smtp_password_wo`. Increment this value to trigger a password update. Required when using `smtp_password_wo`.
* `tls_auth_configuration` - (Optional) Configuration used to authenticate with `MTLS` ingress points. See [`tls_auth_configuration` Block](#tls_auth_configuration-block) for details.

### `tls_auth_configuration` Block

The `tls_auth_configuration` block supports the following:

* `trust_store` - (Required) Trust store used to validate client certificates. See [`trust_store` Block](#trust_store-block) for details.

### `trust_store` Block

The `trust_store` block supports the following:

* `ca_content` - (Required) PEM-encoded certificate authority (CA) content used to validate client certificates.
* `crl_content` - (Optional) PEM-encoded certificate revocation list (CRL) content used to check whether client certificates have been revoked.
* `kms_key_arn` - (Optional) ARN of the AWS KMS key used to decrypt the CRL content.

### `network_configuration` Block

The `network_configuration` block supports the following:

* `private_network_configuration` - (Optional) Configuration for a private ingress point that uses a VPC endpoint. See [`private_network_configuration` Block](#private_network_configuration-block) for details.
* `public_network_configuration` - (Optional) Configuration for a public ingress point. See [`public_network_configuration` Block](#public_network_configuration-block) for details.

### `private_network_configuration` Block

The `private_network_configuration` block supports the following:

* `vpc_endpoint_id` - (Required) Identifier of the VPC endpoint to associate with the ingress point.

### `public_network_configuration` Block

The `public_network_configuration` block supports the following:

* `ip_type` - (Required) IP address type for the public ingress point. Valid values are `IPV4` and `DUAL_STACK`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `a_record` - DNS A record that identifies your ingress endpoint for email clients.
* `arn` - ARN of the Ingress Point.
* `created_timestamp` - Timestamp of when the ingress point was created.
* `last_updated_timestamp` - Timestamp of when the ingress point was last updated.
* `status` - Status of the ingress point.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_mailmanager_ingress_point.example
  identity = {
    id = "ingress_point-id-12345678"
  }
}

resource "aws_mailmanager_ingress_point" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) Identifier of the Ingress Point.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SES Mail Manager Ingress Point using the `id`. For example:

```terraform
import {
  to = aws_mailmanager_ingress_point.example
  id = "ingress_point-id-12345678"
}
```

Using `terraform import`, import SES Mail Manager Ingress Point using the `id`. For example:

```console
% terraform import aws_mailmanager_ingress_point.example ingress_point-id-12345678
```
