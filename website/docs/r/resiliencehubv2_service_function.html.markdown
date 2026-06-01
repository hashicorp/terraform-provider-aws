---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_service_function"
description: |-
  Terraform resource for managing an AWS Resilience Hub V2 Service Function.
---

# Resource: aws_resiliencehubv2_service_function

Terraform resource for managing an AWS Resilience Hub V2 Service Function.

A service function represents a technical subset of the service topology that represents a specific workflow within a service. For example, an authentication service might have separate service functions for "SSO sign-in" and "Registration".

## Example Usage

### Basic Usage

```hcl
resource "aws_resiliencehubv2_service_function" "example" {
  service_arn = aws_resiliencehubv2_service.example.arn
  name        = "example-function"
  criticality = "PRIMARY"
}
```

### With Description

```hcl
resource "aws_resiliencehubv2_service_function" "example" {
  service_arn = aws_resiliencehubv2_service.example.arn
  name        = "payment-processing"
  description = "Handles payment transaction processing"
  criticality = "PRIMARY"
}
```

## Argument Reference

The following arguments are required:

* `criticality` - (Required) Criticality level of the service function. Valid values: `PRIMARY`, `SUPPLEMENTAL`.
* `name` - (Required) Name of the service function.
* `service_arn` - (Required) ARN of the service this function belongs to. Changing this value requires creating a new resource.

The following arguments are optional:

* `description` - (Optional) Description of the service function.
* `region` - (Optional, **Deprecated**) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Composite identifier in the format `service_arn,service_function_id`.
* `service_function_id` - Unique identifier of the service function.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub V2 Service Function using the `service_arn` and `service_function_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_resiliencehubv2_service_function.example
  id = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123,12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import Resilience Hub V2 Service Function using the `service_arn` and `service_function_id` separated by a comma (`,`). For example:

```console
% terraform import aws_resiliencehubv2_service_function.example arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123,12345678-1234-1234-1234-123456789012
```
