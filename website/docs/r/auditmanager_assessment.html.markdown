---
subcategory: "Audit Manager"
layout: "aws"
page_title: "AWS: aws_auditmanager_assessment"
description: |-
  Terraform resource for managing an AWS Audit Manager Assessment.
---

# Resource: aws_auditmanager_assessment

Terraform resource for managing an AWS Audit Manager Assessment.

## Example Usage

### Basic Usage

```terraform
resource "aws_auditmanager_assessment" "test" {
  name = "example"

  assessment_reports_destination {
    destination      = "s3://${aws_s3_bucket.test.id}"
    destination_type = "S3"
  }

  framework_id = aws_auditmanager_framework.test.id

  roles {
    role_arn  = aws_iam_role.test.arn
    role_type = "PROCESS_OWNER"
  }

  scope {
    aws_accounts {
      id = data.aws_caller_identity.current.account_id
    }
    aws_services {
      service_name = "S3"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the assessment.
* `assessment_reports_destination` - (Required) Assessment report storage destination configuration. See [`assessment_reports_destination`](#assessment_reports_destination) below.
* `framework_id` - (Required) Unique identifier of the framework the assessment will be created from.
* `roles` - (Required) List of roles for the assessment. See [`roles`](#roles) below.
* `scope` - (Required) Amazon Web Services accounts and services that are in scope for the assessment. See [`scope`](#scope) below.

The following arguments are optional:

* `description` - (Optional) Description of the assessment.
* `tags` - (Optional) A map of tags to assign to the assessment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### assessment_reports_destination

* `destination` - (Required) Destination of the assessment report. This value be in the form `s3://{bucket_name}`.
* `destination_type` - (Required) Destination type. Currently, `S3` is the only valid value.

### roles

* `role_arn` - (Required) Amazon Resource Name (ARN) of the IAM role.
* `role_type` - (Required) Type of customer persona. For assessment creation, type must always be `PROCESS_OWNER`.

### scope

* `aws_accounts` - Amazon Web Services accounts that are in scope for the assessment. See [`aws_accounts`](#aws_accounts) below.
* `aws_services` - Amazon Web Services services that are included in the scope of the assessment. See [`aws_services`](#aws_services) below.

### aws_accounts

* `id` - (Required) Identifier for the Amazon Web Services account.

### aws_services

* `service_name` - (Required) Name of the Amazon Web Service.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the assessment.
* `id` - Unique identifier for the assessment.
* `roles_all` - Complete list of all roles with access to the assessment. This includes both roles explicitly configured via the `roles` block, and any roles which have access to all Audit Manager assessments by default.
* `status` - Status of the assessment. Valid values are `ACTIVE` and `INACTIVE`.

## Import

Audit Manager Assessments can be imported using the assessment `id`, e.g.,

```
$ terraform import aws_auditmanager_assessment.example abc123-de45
```
