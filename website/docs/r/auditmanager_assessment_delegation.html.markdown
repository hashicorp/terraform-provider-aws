---
subcategory: "Audit Manager"
layout: "aws"
page_title: "AWS: aws_auditmanager_assessment_delegation"
description: |-
  Terraform resource for managing an AWS Audit Manager Assessment Delegation.
---

# Resource: aws_auditmanager_assessment_delegation

Terraform resource for managing an AWS Audit Manager Assessment Delegation.

## Example Usage

### Basic Usage

```terraform
resource "aws_auditmanager_assessment_delegation" "example" {
  assessment_id  = aws_auditmanager_assessment.example.id
  role_arn       = aws_iam_role.example.arn
  role_type      = "RESOURCE_OWNER"
  control_set_id = "example"
}
```

## Argument Reference

The following arguments are required:

* `assessment_id` - (Required) Identifier for the assessment.
* `control_set_id` - (Required) Assessment control set name. This value is the control set name used during assessment creation (not the AWS-generated ID). The `_id` suffix on this attribute has been preserved to be consistent with the underlying AWS API.
* `role_arn` - (Required) Amazon Resource Name (ARN) of the IAM role.
* `role_type` - (Required) Type of customer persona. For assessment delegation, type must always be `RESOURCE_OWNER`.

The following arguments are optional:

* `comment` - (Optional) Comment describing the delegation request.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `delegation_id` - Unique identifier for the delegation.
* `id` - Unique identifier for the resource. This is a comma-separated string containing `assessment_id`, `role_arn`, and `control_set_id`.
* `status` - Status of the delegation.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Audit Manager Assessment Delegation using the `id`. For example:

```terraform
import {
  to = aws_auditmanager_assessment_delegation.example
  id = "abcdef-123456,arn:aws:iam::012345678901:role/example,example"
}
```

Using `terraform import`, import Audit Manager Assessment Delegation using the `id`. For example:

```console
% terraform import aws_auditmanager_assessment_delegation.example abcdef-123456,arn:aws:iam::012345678901:role/example,example
```
