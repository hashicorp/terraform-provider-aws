---
subcategory: "Amazon Q Business"
layout: "aws"
page_title: "AWS: aws_qbusiness_application"
description: |-
  Provides a Q Business Application resource.
---

# Resource: aws_qbusiness_application

Provides a Q Business Application resource.

## Example Usage

```terraform
resource "aws_qbusiness_application" "example" {
  display_name                 = "example-app"
  iam_service_role_arn         = aws_iam_role.example.arn
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]

  attachments_configuration {
    attachments_control_mode = "ENABLED"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `display_name` - (Required) A name for the Amazon Q application.
* `iam_service_role_arn` - (Required) The Amazon Resource Name (ARN) of an IAM role with permissions to access your Amazon CloudWatch logs and metrics.
* `identity_center_instance_arn` - (Required) ARN of the IAM Identity Center instance you are either creating for — or connecting to — your Amazon Q Business application
* `description` - (Optional) A description for the Amazon Q application.
* `attachments_configuration` - (Required) Status information about whether file upload functionality is activated or deactivated for your end user.
* `encryption_configuration` - (Optional) The identifier of the AWS KMS key that is used to encrypt your data. Amazon Q doesn't support asymmetric keys.

`attachments_configuration` supports the following:

* `attachments_control_mode` - (Required) Status information about whether file upload functionality is activated or deactivated for your end user. Valid values are `ENABLED` and `DISABLED`.

`encryption_configuration` supports the following:

* `kms_key_id` - (Required) The identifier of the AWS KMS key that is used to encrypt your data. Amazon Q doesn't support asymmetric keys.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `application_id` - The Application ID of the Q Business App.
* `arn` - Amazon Resource Name (ARN) of the Q Business App.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `identity_center_application_arn` - ARN of the AWS IAM Identity Center application attached to your Amazon Q Business application.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Q Business Application using the `id`. For example:

```terraform
import {
  to = aws_qbusiness_application.example
  id = "id-12345678"
}
```

Using `terraform import`, import a Q Business Application using the `id`. For example:

```console
% terraform import aws_qbusiness_application.example id-12345678
```
