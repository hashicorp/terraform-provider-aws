---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_user_journey"
description: |-
  Terraform resource for managing an AWS Resilience Hub V2 User Journey.
---

# Resource: aws_resiliencehubv2_user_journey

Terraform resource for managing an AWS Resilience Hub V2 User Journey.

A user journey describes a critical end-user path or business capability within a system (e.g., "Path to purchase", "Order fulfillment"). User journeys reference services and can have resilience policies applied at the journey level.

## Example Usage

### Basic Usage

```hcl
resource "aws_resiliencehubv2_system" "example" {
  name = "example-system"
}

resource "aws_resiliencehubv2_user_journey" "example" {
  system_arn = aws_resiliencehubv2_system.example.arn
  name       = "example-user-journey"
}
```

### With Policy

```hcl
resource "aws_resiliencehubv2_system" "example" {
  name = "example-system"
}

resource "aws_resiliencehubv2_policy" "example" {
  name = "example-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_user_journey" "example" {
  system_arn  = aws_resiliencehubv2_system.example.arn
  name        = "checkout-flow"
  description = "End-to-end checkout user journey"
  policy_arn  = aws_resiliencehubv2_policy.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the user journey.
* `system_arn` - (Required) ARN of the system this user journey belongs to. Changing this value requires creating a new resource.

The following arguments are optional:

* `description` - (Optional) Description of the user journey.
* `policy_arn` - (Optional) ARN of the resilience policy to associate with this user journey.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Composite identifier in the format `system_arn,user_journey_id`.
* `user_journey_id` - Unique identifier of the user journey.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_resiliencehubv2_user_journey.example
  identity = {
    system_arn      = "arn:aws:resiliencehub:us-west-2:123456789012:system/example-system:abc123"
    user_journey_id = "12345678-1234-1234-1234-123456789012"
  }
}

resource "aws_resiliencehubv2_user_journey" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `system_arn` (String) ARN of the system this user journey belongs to.
* `user_journey_id` (String) Unique identifier of the user journey.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub V2 User Journey using the `system_arn` and `user_journey_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_resiliencehubv2_user_journey.example
  id = "arn:aws:resiliencehub:us-west-2:123456789012:system/example-system:abc123,12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import Resilience Hub V2 User Journey using the `system_arn` and `user_journey_id` separated by a comma (`,`). For example:

```console
% terraform import aws_resiliencehubv2_user_journey.example arn:aws:resiliencehub:us-west-2:123456789012:system/example-system:abc123,12345678-1234-1234-1234-123456789012
```
