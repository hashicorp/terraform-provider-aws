---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application_assignments"
description: |-
  Terraform data source for managing AWS SSO Admin Application Assignments.
---

# Data Source: aws_ssoadmin_application_assignments

Terraform data source for managing AWS SSO Admin Application Assignments.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_application_assignments" "example" {
  application_arn = aws_ssoadmin_application.example.application_arn
}
```

## Argument Reference

The following arguments are required:

* `application_arn` - (Required) ARN of the application.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `application_assignments` - List of principals assigned to the application. See the [`application_assignments` attribute reference](#application_assignments-attribute-reference) below.

### `application_assignments` Attribute Reference

* `application_arn` - ARN of the application.
* `principal_id` - An identifier for an object in IAM Identity Center, such as a user or group.
* `principal_type` - Entity type for which the assignment will be created. Valid values are `USER` or `GROUP`.
