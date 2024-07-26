---
subcategory: "Audit Manager"
layout: "aws"
page_title: "AWS: aws_auditmanager_control"
description: |-
  Terraform data source for managing an AWS Audit Manager Control.
---

# Data Source: aws_auditmanager_control

Terraform data source for managing an AWS Audit Manager Control.

## Example Usage

### Basic Usage

```terraform
data "aws_auditmanager_control" "example" {
  name = "1. Risk Management"
  type = "Standard"
}
```

### With Framework Resource

```terraform
data "aws_auditmanager_control" "example" {
  name = "1. Risk Management"
  type = "Standard"
}

data "aws_auditmanager_control" "example2" {
  name = "2. Personnel"
  type = "Standard"
}

resource "aws_auditmanager_framework" "example" {
  name = "example"

  control_sets {
    name = "example"
    controls {
      id = data.aws_auditmanager_control.example.id
    }
  }
  control_sets {
    name = "example2"
    controls {
      id = data.aws_auditmanager_control.example2.id
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the control.
* `type` - (Required) Type of control. Valid values are `Custom` and `Standard`.

## Attribute Reference

See the [`aws_auditmanager_control` resource](/docs/providers/aws/r/auditmanager_control.html) for details on the returned attributes - they are identical.
