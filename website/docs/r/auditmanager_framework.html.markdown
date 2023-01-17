---
subcategory: "Audit Manager"
layout: "aws"
page_title: "AWS: aws_auditmanager_framework"
description: |-
  Terraform resource for managing an AWS Audit Manager Framework.
---

# Resource: aws_auditmanager_framework

Terraform resource for managing an AWS Audit Manager Framework.

## Example Usage

### Basic Usage

```terraform
resource "aws_auditmanager_framework" "test" {
  name = "example"

  control_sets {
    name = "example"
    controls {
      id = aws_auditmanager_control.test.id
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the framework.
* `control_sets` - (Required) Control sets that are associated with the framework. See [`control_sets`](#control_sets) below.

The following arguments are optional:

* `compliance_type` - (Optional) Compliance type that the new custom framework supports, such as `CIS` or `HIPAA`.
* `description` - (Optional) Description of the framework.
* `tags` - (Optional) A map of tags to assign to the framework. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### control_sets

* `name` - (Required) Name of the control set.
* `controls` - (Required) List of controls within the control set. See [`controls`](#controls) below.

### controls

* `id` - (Required) Unique identifier of the control.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the framework.
* `control_sets[*].id` - Unique identifier for the framework control set.
* `id` - Unique identifier for the framework.
* `framework_type` - Framework type, such as a custom framework or a standard framework.

## Import

Audit Manager Framework can be imported using the framework `id`, e.g.,

```
$ terraform import aws_auditmanager_framework.example abc123-de45
```
