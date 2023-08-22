---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_observability_configuration"
description: |-
  Manages an App Runner Observability Configuration.
---

# Resource: aws_apprunner_observability_configuration

Manages an App Runner Observability Configuration.

## Example Usage

```terraform
resource "aws_apprunner_observability_configuration" "example" {
  observability_configuration_name = "example"

  trace_configuration {
    vendor = "AWSXRAY"
  }

  tags = {
    Name = "example-apprunner-observability-configuration"
  }
}
```

## Argument Reference

The following arguments supported:

* `observability_configuration_name` - (Required, Forces new resource) Name of the observability configuration.
* `trace_configuration` - (Optional) Configuration of the tracing feature within this observability configuration. If you don't specify it, App Runner doesn't enable tracing. See [Trace Configuration](#trace-configuration) below for more details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Trace Configuration

The `trace_configuration` block supports the following argument:

* `vendor` - (Required) Implementation provider chosen for tracing App Runner services. Valid values: `AWSXRAY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of this observability configuration.
* `observability_configuration_revision` - The revision of this observability configuration.
* `latest` - Whether the observability configuration has the highest `observability_configuration_revision` among all configurations that share the same `observability_configuration_name`.
* `status` - Current state of the observability configuration. An INACTIVE configuration revision has been deleted and can't be used. It is permanently removed some time after deletion.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Runner Observability Configuration using the `arn`. For example:

```terraform
import {
  to = aws_apprunner_observability_configuration.example
  id = "arn:aws:apprunner:us-east-1:1234567890:observabilityconfiguration/example/1/d75bc7ea55b71e724fe5c23452fe22a1"
}
```

Using `terraform import`, import App Runner Observability Configuration using the `arn`. For example:

```console
% terraform import aws_apprunner_observability_configuration.example arn:aws:apprunner:us-east-1:1234567890:observabilityconfiguration/example/1/d75bc7ea55b71e724fe5c23452fe22a1
```
