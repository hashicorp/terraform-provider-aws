---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_organization_features"
description: |-
  Terraform resource for managing an AWS IAM (Identity & Access Management) Organization Features.
---

# Resource: aws_iam_organization_features

Manages a IAM Organization Features for centralized root access for member accounts.. More information about managing root access in IAM can be found in the [Centralize root access for member accounts](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_root-enable-root-access.html).

~> **NOTE:** Before managing IAM Organization features, the AWS account utilizing this resource must be an Organizations management account. Also, you must enable trusted access for AWS Identity and Access Management in AWS Organizations.

## Example Usage

```terraform
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["iam.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_iam_organization_features" "example" {
  features = [
    "RootCredentialsManagement", 
    "RootSessions"
  ]
}
```

## Argument Reference

The following arguments are required:

* `features` - (Required) List of IAM features to enable. Valid values are `RootCredentialsManagement` and `RootSessions`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS Organization identifier.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM (Identity & Access Management) Organization Features using the `id`. For example:

```terraform
import {
  to = aws_iam_organization_features.example
  id = "o-1234567"
}
```

Using `terraform import`, import IAM (Identity & Access Management) Organization Features using the `id`. For example:

```console
% terraform import aws_iam_organization_features.example o-1234567
```
