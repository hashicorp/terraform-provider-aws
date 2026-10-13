---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_domain"
description: |-
  Manages an Amazon CloudWatch Omni domain.
---

# Resource: aws_cloudwatchomni_domain

Manages an Amazon CloudWatch Omni domain.

A domain is the top-level container for CloudWatch Omni. It configures how users sign in and contains the spaces that hold telemetry. An account can have one domain.

~> **NOTE:** A domain cannot be deleted while it contains spaces.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatchomni_domain" "example" {
  name               = "example"
  identity_providers = ["IAM"]
}
```

### With IAM Identity Center

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_cloudwatchomni_domain" "example" {
  name               = "example"
  identity_providers = ["IAM", "IDC"]

  identity_provider_configuration {
    identity_center_configuration {
      identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `identity_providers` - (Required) Set of identity providers that users can sign in with. Valid values: `IAM`, `IDC`.
* `name` - (Required) Name of the domain. Must be 3-63 characters, contain only lowercase letters, numbers, and hyphens, begin and end with a letter or number, and not contain consecutive hyphens. Changing the name changes the domain's custom endpoint URLs.

The following arguments are optional:

* `identity_provider_configuration` - (Optional) Identity provider configuration. Required when `identity_providers` includes `IDC`. [See below](#identity_provider_configuration-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `identity_provider_configuration` Block

The `identity_provider_configuration` configuration block supports the following arguments:

* `identity_center_configuration` - (Required) IAM Identity Center configuration. [See below](#identity_center_configuration-block).

### `identity_center_configuration` Block

The `identity_center_configuration` configuration block supports the following arguments:

* `identity_center_instance_arn` - (Required) ARN of the IAM Identity Center instance.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the domain.
* `custom_endpoint_urls` - Additional endpoint URLs derived from the domain name.
* `domain_endpoint_url` - HTTPS endpoint URL for accessing the domain.
* `domain_id` - Unique identifier of the domain.
* `identity_center_application_arn` - ARN of the IAM Identity Center application. Only set when `identity_providers` includes `IDC`.

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatchomni_domain.example
  identity = {
    domain_id = "d-0123456789abcdefghijklmno"
  }
}

resource "aws_cloudwatchomni_domain" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `domain_id` (String) Domain ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a CloudWatch Omni Domain by domain ID. For example:

```terraform
import {
  to = aws_cloudwatchomni_domain.example
  id = "d-0123456789abcdefghijklmno"
}
```

Using `terraform import`, import a CloudWatch Omni Domain by domain ID. For example:

```console
% terraform import aws_cloudwatchomni_domain.example d-0123456789abcdefghijklmno
```
