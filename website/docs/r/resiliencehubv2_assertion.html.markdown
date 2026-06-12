---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_assertion"
description: |-
  Terraform resource for managing an AWS Resilience Hub V2 Assertion.
---

# Resource: aws_resiliencehubv2_assertion

Terraform resource for managing an AWS Resilience Hub V2 Assertion.

An assertion is a statement about your application that provides context for failure mode assessments. Assertions help the GenAI assessment engine understand aspects of your architecture that aren't visible from resource configuration alone (e.g., "Data loss is unacceptable", "Typical traffic is 1000 TPS spiking to 10000 TPS").

## Example Usage

### Basic Usage

```hcl
resource "aws_resiliencehubv2_assertion" "example" {
  service_arn = aws_resiliencehubv2_service.example.arn
  text        = "The service must recover within 5 minutes of an AZ failure"
}
```

## Argument Reference

The following arguments are required:

* `service_arn` - (Required) ARN of the service this assertion belongs to. Changing this value requires creating a new resource.
* `text` - (Required) Text of the resilience assertion.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `assertion_id` - Unique identifier of the assertion.
* `id` - Composite identifier in the format `service_arn,assertion_id`.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_resiliencehubv2_assertion.example
  identity = {
    service_arn  = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123"
    assertion_id = "12345678-1234-1234-1234-123456789012"
  }
}

resource "aws_resiliencehubv2_assertion" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `service_arn` (String) ARN of the service this assertion belongs to.
* `assertion_id` (String) Unique identifier of the assertion.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub V2 Assertion using the `service_arn` and `assertion_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_resiliencehubv2_assertion.example
  id = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123,12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import Resilience Hub V2 Assertion using the `service_arn` and `assertion_id` separated by a comma (`,`). For example:

```console
% terraform import aws_resiliencehubv2_assertion.example arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123,12345678-1234-1234-1234-123456789012
```
