---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application"
description: |-
  Terraform data source for managing an AWS SSO Admin Application.
---

# Data Source: aws_ssoadmin_application

Terraform data source for managing an AWS SSO Admin Application.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_application" "example" {
  application_arn = "arn:aws:sso::123456789012:application/ssoins-1234/apl-5678"
}
```

## Argument Reference

The following arguments are required:

* `application_arn` - (Required) ARN of the application.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `application_account` - AWS account ID.
* `application_provider_arn` - ARN of the application provider.
* `description` - Description of the application.
* `id` - ARN of the application.
* `instance_arn` - ARN of the instance of IAM Identity Center.
* `name` - Name of the application.
* `portal_options` - Options for the portal associated with an application. See the `aws_ssoadmin_application` [resource documentation](../r/ssoadmin_application.html.markdown#portal_options-argument-reference). The attributes are the same.
* `status` - Status of the application.
