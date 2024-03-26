---
subcategory: "Audit Manager"
layout: "aws"
page_title: "AWS: aws_auditmanager_assessment_report"
description: |-
  Terraform resource for managing an AWS Audit Manager Assessment Report.
---

# Resource: aws_auditmanager_assessment_report

Terraform resource for managing an AWS Audit Manager Assessment Report.

## Example Usage

### Basic Usage

```terraform
resource "aws_auditmanager_assessment_report" "test" {
  name          = "example"
  assessment_id = aws_auditmanager_assessment.test.id
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the assessment report.
* `assessment_id` - (Required) Unique identifier of the assessment to create the report from.

The following arguments are optional:

* `description` - (Optional) Description of the assessment report.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `author` - Name of the user who created the assessment report.
* `id` - Unique identifier for the assessment report.
* `status` - Current status of the specified assessment report. Valid values are `COMPLETE`, `IN_PROGRESS`, and `FAILED`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Audit Manager Assessment Reports using the assessment report `id`. For example:

```terraform
import {
  to = aws_auditmanager_assessment_report.example
  id = "abc123-de45"
}
```

Using `terraform import`, import Audit Manager Assessment Reports using the assessment report `id`. For example:

```console
% terraform import aws_auditmanager_assessment_report.example abc123-de45
```
