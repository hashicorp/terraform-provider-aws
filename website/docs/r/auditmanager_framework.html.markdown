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
      id = aws_auditmanager_control.test_1.id
    }
    controls {
      id = aws_auditmanager_control.test_2.id
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the framework.
* `control_sets` - (Required) Configuration block(s) for the control sets that are associated with the framework. See [`control_sets` Block](#control_sets-block) below for details.

The following arguments are optional:

* `compliance_type` - (Optional) Compliance type that the new custom framework supports, such as `CIS` or `HIPAA`.
* `description` - (Optional) Description of the framework.
* `tags` - (Optional) A map of tags to assign to the framework. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `control_sets` Block

The `control_sets` configuration block supports the following arguments:

* `name` - (Required) Name of the control set.
* `controls` - (Required) Configuration block(s) for the controls within the control set. See [`controls` Block](#controls-block) below for details.

### `controls` Block

The `controls` configuration block supports the following arguments:

* `id` - (Required) Unique identifier of the control.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the framework.
* `control_sets[*].id` - Unique identifier for the framework control set.
* `id` - Unique identifier for the framework.
* `framework_type` - Framework type, such as a custom framework or a standard framework.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Audit Manager Framework using the framework `id`. For example:

```terraform
import {
  to = aws_auditmanager_framework.example
  id = "abc123-de45"
}
```

Using `terraform import`, import Audit Manager Framework using the framework `id`. For example:

```console
% terraform import aws_auditmanager_framework.example abc123-de45
```
