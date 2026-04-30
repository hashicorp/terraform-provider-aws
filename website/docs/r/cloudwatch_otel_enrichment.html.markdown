---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_otel_enrichment"
description: |-
  Manages AWS CloudWatch OTel enrichment.
---

# Resource: aws_cloudwatch_otel_enrichment

Manages AWS CloudWatch OTel enrichment. This is a singleton resource that enables OTel enrichment at the account level.

~> **NOTE:** This resource requires the `aws_observabilityadmin_telemetry_enrichment` resource to be configured first. Without telemetry enrichment enabled, OTel enrichment will not function properly even if the API accepts the configuration.

## Example Usage

### Enable OTel Enrichment

```terraform
resource "aws_observabilityadmin_telemetry_enrichment" "example" {
}

resource "aws_cloudwatch_otel_enrichment" "example" {
  depends_on = [aws_observabilityadmin_telemetry_enrichment.example]
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) AWS region where this resource is managed.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS region where the enrichment is managed.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_otel_enrichment.example
  identity = {
  }
}

resource "aws_cloudwatch_otel_enrichment" "example" {
}
```

### Identity Schema

#### Required

No required attributes for singleton identity.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch OTel Enrichment using the region. For example:

```terraform
import {
  to = aws_cloudwatch_otel_enrichment.example
  id = "us-west-2"
}
```

Using `terraform import`, import CloudWatch OTel Enrichment using the region. For example:

```console
% terraform import aws_cloudwatch_otel_enrichment.example us-west-2
```
