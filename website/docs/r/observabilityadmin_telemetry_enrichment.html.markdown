---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_enrichment"
description: |-
  Manages an AWS CloudWatch Observability Admin Telemetry Enrichment.
---

# Resource: aws_observabilityadmin_telemetry_enrichment

Manages an AWS CloudWatch Observability Admin Telemetry Enrichment.

Telemetry enrichment enables resource tags for telemetry data in your account, enhancing telemetry with additional resource metadata from AWS Resource Explorer to provide richer context for monitoring and observability.

For more information, see the [AWS CloudWatch Observability Admin documentation](https://docs.aws.amazon.com/cloudwatch/latest/observabilityadmin/what-is-observabilityadmin.html).

~> **NOTE:** Only one telemetry enrichment can exist per account per region. Creating this resource enables the feature; destroying it disables the feature.

## Example Usage

### Basic Usage

```terraform
resource "aws_observabilityadmin_telemetry_enrichment" "example" {}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `aws_resource_explorer_managed_view_arn` - ARN of the AWS Resource Explorer managed view created for the telemetry enrichment feature.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_enrichment.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_observabilityadmin_telemetry_enrichment" "example" {}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin Telemetry Enrichment using the region name. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_enrichment.example
  id = "us-west-2"
}
```

Using `terraform import`, import CloudWatch Observability Admin Telemetry Enrichment using the region name. For example:

```console
% terraform import aws_observabilityadmin_telemetry_enrichment.example us-west-2
```
