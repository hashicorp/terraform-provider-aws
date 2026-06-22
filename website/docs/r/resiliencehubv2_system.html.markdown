---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_system"
description: |-
  Terraform resource for managing an AWS Resilience Hub V2 System.
---

# Resource: aws_resiliencehubv2_system

Terraform resource for managing an AWS Resilience Hub V2 System.

A system represents a business application or platform that delivers value to your organization. Systems contain user journeys and services, and serve as the top-level container for organizing your resilience posture.

## Example Usage

### Basic Usage

```hcl
resource "aws_resiliencehubv2_system" "example" {
  name = "example-system"
}
```

### With Sharing Enabled

```hcl
resource "aws_resiliencehubv2_system" "example" {
  name            = "example-system"
  description     = "Production system grouping"
  sharing_enabled = true

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the system. Changing this value requires creating a new resource.

The following arguments are optional:

* `description` - (Optional) Description of the system.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sharing_enabled` - (Optional) Whether cross-account sharing is enabled for this system.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the system.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_resiliencehubv2_system.example
  identity = {
    "arn" = "arn:aws:resiliencehub:us-west-2:123456789012:system/example-system:abc123"
  }
}

resource "aws_resiliencehubv2_system" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the Resilience Hub V2 System.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub V2 System using the `arn`. For example:

```terraform
import {
  to = aws_resiliencehubv2_system.example
  id = "arn:aws:resiliencehub:us-west-2:123456789012:system/example-system:abc123"
}
```

Using `terraform import`, import Resilience Hub V2 System using the `arn`. For example:

```console
% terraform import aws_resiliencehubv2_system.example arn:aws:resiliencehub:us-west-2:123456789012:system/example-system:abc123
```
