---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_assumed_role_source"
description: |-
  Get information on the IAM source role of an STS assumed role
---

# Data Source: aws_iam_assumed_role_source

This data source provides information on the IAM source role of an STS assumed role. For non-role ARNs, this data source simply passes the ARN through.

For some AWS resources, multiple types of principals are allowed in the same argument (e.g., IAM users and IAM roles). However, these arguments often do not allow assumed-role (i.e., STS, temporary credential) principals. Given an STS ARN, this data source provides the ARN for the source IAM role.

## Example Usage

### Basic Example

```terraform
data "aws_iam_assumed_role_source" "example" {
  arn = "arn:aws:sts::123456789012:assumed-role/Audien-Heaven/MatyNoyes"
}
```

### Find the Terraform Runner's Source Role

Combined with `aws_caller_identity`, you can get the current user's source IAM role ARN (`source_arn`) if you're using an assumed role. If you're not using an assumed role, the caller's (e.g., an IAM user's) ARN will simply be passed through. In environments where both IAM users and individuals using assumed roles need to apply the same configurations, this data source enables seamless use.

```terraform
data "aws_caller_identity" "current" {}

data "aws_iam_assumed_role_source" "example" {
  arn = data.aws_called_identity.current.arn
}
```

## Argument Reference

* `arn` - (Required) ARN for an assumed role.

~> If `arn` is a non-role ARN (or non-ARN), Terraform gives no error and `source_arn` will be equal to the `arn` value. For IAM role and STS assumed-role ARNs, Terraform gives an error if the identified IAM role does not exist.

## Attributes Reference

* `role_path` - Path of the source role. Only available if `arn` corresponds to an IAM role or STS assumed role.
* `role_name` - Name of the source role. Only available if `arn` corresponds to an IAM role or STS assumed role.
* `source_arn` - IAM source role ARN if `arn` corresponds to an STS assumed role. Otherwise, `source_arn` is equal to `arn`.
* `session_name` - Name of the STS session. Only available if `arn` corresponds to an STS assumed role.
