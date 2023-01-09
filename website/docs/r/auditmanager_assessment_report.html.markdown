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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `author` - Name of the user who created the assessment report.
* `id` - Unique identifier for the assessment report.
* `status` - Current status of the specified assessment report. Valid values are `COMPLETE`, `IN_PROGRESS`, and `FAILED`.

## Import

Audit Manager Assessment Reports can be imported using the assessment report `id`, e.g.,

```
$ terraform import aws_auditmanager_assessment_report.example abc123-de45
```
