---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_access_grants_instance"
description: |-
  Provides a resource to manage an S3 Access Grants instance.
---

# Resource: aws_s3control_access_grants_instance

Provides a resource to manage an S3 Access Grants instance, which serves as a logical grouping for access grants.
You can have one S3 Access Grants instance per Region in your account.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3control_access_grants_instance" "example" {}
```

### AWS IAM Identity Center

```terraform
resource "aws_s3control_access_grants_instance" "example" {
  identity_center_arn = "arn:aws:sso:::instance/ssoins-890759e9c7bfdc1d"
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The AWS account ID for the S3 Access Grants instance. Defaults to automatically determined account ID of the Terraform AWS provider.
* `identity_center_arn` - (Optional) The ARN of the AWS IAM Identity Center instance associated with the S3 Access Grants instance.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `access_grants_instance_arn` - Amazon Resource Name (ARN) of the S3 Access Grants instance.
* `access_grants_instance_id` - Unique ID of the S3 Access Grants instance.
* `identity_center_application_arn` - The ARN of the AWS IAM Identity Center instance application; a subresource of the original Identity Center instance.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Access Grants instances using the `account_id`. For example:

```terraform
import {
  to = aws_s3control_access_grants_instance.example
  id = "123456789012"
}
```

Using `terraform import`, import S3 Access Grants instances using the `account_id`. For example:

```console
% terraform import aws_s3control_access_grants_instance.example 123456789012
```
