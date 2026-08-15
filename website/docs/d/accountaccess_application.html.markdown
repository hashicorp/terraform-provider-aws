---
subcategory: "Account Access"
layout: "aws"
page_title: "AWS: aws_accountaccess_application"
description: |-
  Looks up an existing AWS Account Access Application.
---

# Data Source: aws_accountaccess_application

Looks up an existing AWS Account Access [Application](../r/accountaccess_application.html.markdown). Lookup is by Application ARN or by IAM Identity Center instance ARN (one Application per instance).

## Example Usage

### By Application ARN

```terraform
data "aws_accountaccess_application" "example" {
  arn = "arn:aws:account-access:us-east-1:123456789012:application/aam-0123456789abcdef"
}
```

### By Identity Center Instance

```terraform
data "aws_ssoadmin_instances" "example" {}

data "aws_accountaccess_application" "example" {
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}
```

## Argument Reference

The following arguments are optional:

* `arn` - (Optional) ARN of the Application to look up. Exactly one of `arn` or `identity_center_instance_arn` must be specified.
* `identity_center_instance_arn` - (Optional) ARN of the IAM Identity Center instance whose bound Application should be returned. Exactly one of `arn` or `identity_center_instance_arn` must be specified.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `created_at` - Date and time, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339), when the Application was created.
* `identity_center_application_arn` - ARN of the IAM Identity Center Application that Account Access provisioned for this resource.
* `status` - Current lifecycle status. One of `CREATE_IN_PROGRESS`, `ACTIVE`, `DELETE_IN_PROGRESS`, `CREATE_FAILED`, `DELETE_FAILED`.
* `tags` - Map of tags assigned to the Application.
* `tenant_id` - Internal tenant identifier returned by the service.
* `updated_at` - Date and time, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339), when the Application was last updated.
