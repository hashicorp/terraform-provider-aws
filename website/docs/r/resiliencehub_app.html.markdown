---
subcategory: "Resilience Hub"
layout: "aws"
page_title: "AWS: aws_resiliencehub_app"
description: |-
  Terraform resource for managing an AWS Resilience Hub App.
---

# Resource: aws_resiliencehub_app

Terraform resource for managing an AWS Resilience Hub App.

## Example Usage

### Basic Usage

```terraform
resource "aws_resiliencehub_app" "example" {
  name = "example-app"
}
```

### With a Resiliency Policy

```terraform
resource "aws_resiliencehub_resiliency_policy" "example" {
  name = "example-policy"
  tier = "NotApplicable"

  policy {
    az {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    hardware {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    software {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
  }
}

resource "aws_resiliencehub_app" "example" {
  name                  = "example-app"
  description           = "Example application"
  assessment_schedule   = "Daily"
  resiliency_policy_arn = aws_resiliencehub_resiliency_policy.example.arn

  tags = {
    Environment = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the application. Must be 2-60 characters, start with alphanumeric, and contain only alphanumeric, underscore, and hyphen characters.

The following arguments are optional:

* `assessment_schedule` - (Optional) Assessment schedule for the application. Valid values are `Disabled` and `Daily`.
* `description` - (Optional) Description of the application. Maximum 500 characters.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resiliency_policy_arn` - (Optional) ARN of the resiliency policy to associate with the application.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the application.
* `drift_status` - Status of compliance drift (deviation) detected while running an assessment for the application.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_resiliencehub_app.example
  identity = {
    arn = "arn:aws:resiliencehub:us-east-1:123456789012:app/example-app-id"
  }
}

resource "aws_resiliencehub_app" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` - (String) ARN of the application.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub App using the `arn`. For example:

```terraform
import {
  to = aws_resiliencehub_app.example
  id = "arn:aws:resiliencehub:us-east-1:123456789012:app/example-app-id"
}
```

Using `terraform import`, import Resilience Hub App using the `arn`. For example:

```console
% terraform import aws_resiliencehub_app.example arn:aws:resiliencehub:us-east-1:123456789012:app/example-app-id
```
