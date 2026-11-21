---
subcategory: "ACM (Certificate Manager)"
layout: "aws"
page_title: "AWS: aws_acm_acme_endpoint"
description: |-
  Manages an AWS ACM (Certificate Manager) ACME Endpoint.
---

# Resource: aws_acm_acme_endpoint

Manages an AWS ACM (Certificate Manager) ACME Endpoint.

An ACME endpoint is a managed [ACME](https://datatracker.ietf.org/doc/html/rfc8555) server with a unique endpoint URL. ACME clients such as `certbot` use the endpoint URL to automate public certificate issuance. Clients authenticate to the endpoint using an external account binding, and each domain that the endpoint may issue certificates for must be pre-validated.

## Example Usage

### Basic Usage

```terraform
resource "aws_acm_acme_endpoint" "example" {
  authorization_behavior = "PRE_APPROVED"

  certificate_authority {
    public_certificate_authority {
      allowed_key_algorithms = ["RSA_2048"]
    }
  }
}
```

### Requiring Contact Information

```terraform
resource "aws_acm_acme_endpoint" "example" {
  authorization_behavior = "PRE_APPROVED"
  contact                = "REQUIRED"

  certificate_authority {
    public_certificate_authority {
      allowed_key_algorithms = ["RSA_2048", "EC_prime256v1"]
    }
  }
}
```

### Tagging Issued Certificates

```terraform
resource "aws_acm_acme_endpoint" "example" {
  authorization_behavior = "PRE_APPROVED"

  certificate_tags = {
    ManagedBy = "acme"
  }

  certificate_authority {
    public_certificate_authority {
      allowed_key_algorithms = ["RSA_2048"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `authorization_behavior` - (Required) Authorization behavior for the ACME endpoint. Valid value: `PRE_APPROVED`.
* `certificate_authority` - (Required) Certificate authority that issues certificates through this endpoint. See [`certificate_authority` Block](#certificate_authority-block) for details.

The following arguments are optional:

* `certificate_tags` - (Optional) Map of tags to apply to certificates issued through this ACME endpoint. Changing this forces a new resource to be created.
* `contact` - (Optional) Contact information requirement for ACME account registration. Valid values: `REQUIRED`, `NOT_REQUIRED`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `certificate_authority` Block

The `certificate_authority` configuration block supports the following arguments:

* `public_certificate_authority` - (Required) Configuration for issuing certificates from the Amazon public certificate authority. See [`public_certificate_authority` Block](#public_certificate_authority-block) for details.

### `public_certificate_authority` Block

The `public_certificate_authority` configuration block supports the following arguments:

* `allowed_key_algorithms` - (Optional) Set of key algorithms allowed for certificates issued by this certificate authority. Valid values: `RSA_2048`, `EC_prime256v1`, `EC_secp384r1`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the ACME endpoint.
* `created_at` - Date and time when the ACME endpoint was created, in [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339) format.
* `endpoint_url` - Directory URL that ACME clients use to reach the endpoint.
* `failure_reason` - Reason why the ACME endpoint failed, when `status` is `FAILED`.
* `status` - Status of the ACME endpoint. One of `CREATING`, `ACTIVE`, `DELETING`, or `FAILED`.
* `updated_at` - Date and time when the ACME endpoint was last updated, in [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339) format.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_acm_acme_endpoint.example
  identity = {
    "arn" = "arn:aws:acm:us-east-1:123456789012:acme-endpoint/12345678-1234-1234-1234-123456789012"
  }
}

resource "aws_acm_acme_endpoint" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) Amazon Resource Name (ARN) of the ACME endpoint.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ACM (Certificate Manager) ACME Endpoint using the `arn`. For example:

```terraform
import {
  to = aws_acm_acme_endpoint.example
  id = "arn:aws:acm:us-east-1:123456789012:acme-endpoint/12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import ACM (Certificate Manager) ACME Endpoint using the `arn`. For example:

```console
% terraform import aws_acm_acme_endpoint.example arn:aws:acm:us-east-1:123456789012:acme-endpoint/12345678-1234-1234-1234-123456789012
```
