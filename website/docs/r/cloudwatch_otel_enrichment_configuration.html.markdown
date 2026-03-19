---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_otel_enrichment_configuration"
description: |-
  Manages AWS CloudWatch OTel enrichment configuration.
---

# Resource: aws_cloudwatch_otel_enrichment_configuration

Manages AWS CloudWatch OTel enrichment configuration. This is a singleton resource that configures OTel enrichment at the account level.

## Example Usage

### Enable OTel Enrichment

```terraform
resource "aws_cloudwatch_otel_enrichment_configuration" "example" {
  enabled = true
}
```

### Disable OTel Enrichment

```terraform
resource "aws_cloudwatch_otel_enrichment_configuration" "example" {
  enabled = false
}
```

## Argument Reference

The following arguments are required:

* `enabled` - (Required) Whether to enable OTel enrichment for CloudWatch.

The following arguments are optional:

* `region` - (Optional) AWS region where this resource is managed.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS region where the configuration is managed.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_otel_enrichment_configuration.example
  identity = {
  }
}

resource "aws_cloudwatch_otel_enrichment_configuration" "example" {
  enabled = true
}
```

### Identity Schema

#### Required

No required attributes for singleton identity.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch OTel Enrichment Configuration using the region. For example:

```terraform
import {
  to = aws_cloudwatch_otel_enrichment_configuration.example
  id = "us-west-2"
}
```

Using `terraform import`, import CloudWatch OTel Enrichment Configuration using the region. For example:

```console
% terraform import aws_cloudwatch_otel_enrichment_configuration.example us-west-2
```
