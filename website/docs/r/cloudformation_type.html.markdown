---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_type"
description: |-
    Manages a version of a CloudFormation Type.
---

# Resource: aws_cloudformation_type

Manages a version of a CloudFormation Type.

~> **NOTE:** The destroy operation of this resource marks the version as deprecated. If this was the only `LIVE` version, the type is marked as deprecated. Enable the [resource `lifecycle` configuration block `create_before_destroy` argument](https://www.terraform.io/language/meta-arguments/lifecycle#create_before_destroy) in this resource configuration to properly order redeployments in Terraform.

## Example Usage

```terraform
resource "aws_cloudformation_type" "example" {
  schema_handler_package = "s3://${aws_s3_object.example.bucket}/${aws_s3_object.example.key}"
  type                   = "RESOURCE"
  type_name              = "ExampleCompany::ExampleService::ExampleResource"

  logging_config {
    log_group_name = aws_cloudwatch_log_group.example.name
    log_role_arn   = aws_iam_role.example.arn
  }

  lifecycle {
    create_before_destroy = true
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `execution_role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role for CloudFormation to assume when invoking the extension. If your extension calls AWS APIs in any of its handlers, you must create an IAM execution role that includes the necessary permissions to call those AWS APIs, and provision that execution role in your account. When CloudFormation needs to invoke the extension handler, CloudFormation assumes this execution role to create a temporary session token, which it then passes to the extension handler, thereby supplying your extension with the appropriate credentials.
* `logging_config` - (Optional) Configuration block containing logging configuration.
* `schema_handler_package` - (Required) URL to the S3 bucket containing the extension project package that contains the necessary files for the extension you want to register. Must begin with `s3://` or `https://`. For example, `s3://example-bucket/example-object`.
* `type` - (Optional) CloudFormation Registry Type. For example, `RESOURCE` or `MODULE`.
* `type_name` - (Optional) CloudFormation Type name. For example, `ExampleCompany::ExampleService::ExampleResource`.

### logging_config

The `logging_config` configuration block supports the following arguments:

* `log_group_name` - (Required) Name of the CloudWatch Log Group where CloudFormation sends error logging information when invoking the type's handlers.
* `log_role_arn` - (Required) Amazon Resource Name (ARN) of the IAM Role CloudFormation assumes when sending error logging information to CloudWatch Logs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - (Optional) Amazon Resource Name (ARN) of the CloudFormation Type version. See also `type_arn`.
* `default_version_id` - Identifier of the CloudFormation Type default version.
* `deprecated_status` - Deprecation status of the version.
* `description` - Description of the version.
* `documentation_url` - URL of the documentation for the CloudFormation Type.
* `is_default_version` - Whether the CloudFormation Type version is the default version.
* `provisioning_type` - Provisioning behavior of the CloudFormation Type.
* `schema` - JSON document of the CloudFormation Type schema.
* `source_url` - URL of the source code for the CloudFormation Type.
* `type_arn` - (Optional) Amazon Resource Name (ARN) of the CloudFormation Type. See also `arn`.
* `version_id` - (Optional) Identifier of the CloudFormation Type version.
* `visibility` - Scope of the CloudFormation Type.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_cloudformation_type` using the type version Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_cloudformation_type.example
  id = "arn:aws:cloudformation:us-east-1:123456789012:type/resource/ExampleCompany-ExampleService-ExampleType/1"
}
```

Using `terraform import`, import `aws_cloudformation_type` using the type version Amazon Resource Name (ARN). For example:

```console
% terraform import aws_cloudformation_type.example arn:aws:cloudformation:us-east-1:123456789012:type/resource/ExampleCompany-ExampleService-ExampleType/1
```
