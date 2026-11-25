---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_instance_profile"
description: |-
  Lists IAM (Identity & Access Management) Instance Profile resources.
---

# List Resource: aws_iam_instance_profile

Lists IAM (Identity & Access Management) Instance Profile resources.

## Example Usage

### Basic Usage

```terraform
list "aws_iam_instance_profile" "example" {
  provider = aws
}
```

### Filter by Path Prefix

This example will return IAM Instance Profiles with a `path` equal to or beginning with `/example/`.

```terraform
list "aws_iam_instance_profile" "example" {
  provider = aws

  config {
    path_prefix = "/example/"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `path_prefix` - (Optional) Limits the returned IAM Instance Profiles to those within this path.If `path_prefix` is not specified, or is `"/"`, returns all IAM Instance Profiles.
