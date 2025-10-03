---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_web_app"
description: |-
  Terraform resource for managing an AWS Transfer Family Web App.
---

# Resource: aws_transfer_web_app

Terraform resource for managing an AWS Transfer Family Web App.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "example" {}

data "aws_iam_policy_document" "assume_role_transfer" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
      "sts:SetContext"
    ]
    principals {
      type        = "Service"
      identifiers = ["transfer.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "example"
  assume_role_policy = data.aws_iam_policy_document.assume_role_transfer.json
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetDataAccess",
      "s3:ListCallerAccessGrants",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:access-grants/*"
    ]
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "s3:ResourceAccount"
    }
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:ListAccessGrantsInstances"
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "s3:ResourceAccount"
    }
  }
}

resource "aws_iam_role_policy" "example" {
  policy = data.aws_iam_policy_document.example.json
  role   = aws_iam_role.example.name
}

resource "aws_transfer_web_app" "example" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
      role         = aws_iam_role.example.arn
    }
  }
  web_app_units {
    provisioned = 1
  }
  tags = {
    Name = "test"
  }
}
```

## Argument Reference

The following arguments are required:

* `identity_provider_details` - (Required) Block for details of the identity provider to use with the web app. See [Identity provider details](#identity-provider-details) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `access_endpoint` - (Optional) URL provided to interact with the Transfer Family web app.
* `tags` - (Optional) Key-value pairs that can be used to group and search for web apps.
* `web_app_endpoint_policy` - (Optional) Type of endpoint policy for the web app. Valid values are: `STANDARD`(default) or `FIPS`.
* `web_app_units` - (Optional) Block for number of concurrent connections or the user sessions on the web app.
    * provisioned - (Optional) Number of units of concurrent connections.

### Identity provider details

* `identity_center_config` - (Optional) Block that describes the values to use for the IAM Identity Center settings. See [Identity center config](#identity-center-config) below.

### Identity center config

* `instance_arn` - (Optional) ARN of the IAM Identity Center used for the web app.
* `role` - (Optional) ARN of an identity bearer role for your web app.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Web App.
* `web_app_id` - ID of the Wep App resource.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transfer Family Web App using the `web_app_id`. For example:

```terraform
import {
  to = aws_transfer_web_app.example
  id = "web_app-id-12345678"
}
```

Using `terraform import`, import Transfer Family Web App using the `web_app_id`. For example:

```console
% terraform import aws_transfer_web_app.example web_app-id-12345678
```
