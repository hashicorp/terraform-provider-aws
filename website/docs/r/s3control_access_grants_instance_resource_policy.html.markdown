---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_access_grants_instance_resource_policy"
description: |-
  Provides a resource to manage an S3 Access Grants instance resource policy.
---

# Resource: aws_s3control_access_grants_instance_resource_policy

Provides a resource to manage an S3 Access Grants instance resource policy.
Use a resource policy to manage cross-account access to your S3 Access Grants instance.

## Example Usage

```terraform
resource "aws_s3control_access_grants_instance" "example" {}

resource "aws_s3control_access_grants_instance_resource_policy" "example" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "S3AccessGrantsPolicy",
  "Statement": [{
    "Sid": "AllowAccessToS3AccessGrants",
    "Effect": "Allow",
    "Principal": {
      "AWS": "123456789456"
    },
    "Action": [
      "s3:ListAccessGrants",
      "s3:ListAccessGrantsLocations",
      "s3:GetDataAccess"
    ],
    "Resource": "${aws_s3control_access_grants_instance.example.access_grants_instance_arn}"
  }]
}
EOF
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The AWS account ID for the S3 Access Grants instance. Defaults to automatically determined account ID of the Terraform AWS provider.
* `policy` - (Optional) The policy document.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Access Grants instance resource policies using the `account_id`. For example:

```terraform
import {
  to = aws_s3control_access_grants_instance_resource_policy.example
  id = "123456789012"
}
```

Using `terraform import`, import S3 Access Grants instance resource policies using the `account_id`. For example:

```console
% terraform import aws_s3control_access_grants_instance_resource_policy.example 123456789012
```
