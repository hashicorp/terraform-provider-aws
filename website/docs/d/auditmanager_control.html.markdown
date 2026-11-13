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

This data source supports the following arguments:

* `name` - (Required) Name of the control.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `type` - (Required) Type of control. Valid values are `Custom` and `Standard`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `action_plan_instructions` - Recommended actions to carry out if the control isn't fulfilled.
* `action_plan_title` - Title of the action plan for remediating the control.
* `arn` - ARN of the control.
* `control_mapping_sources` - Data mapping sources for the control.
* `description` - Description of the control.
* `id` - Unique identifier for the control.
* `tags` - Map of tags assigned to the control.
* `testing_information` - Steps to follow to determine if the control is satisfied.
