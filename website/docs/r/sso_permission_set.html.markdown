---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_sso_permission_set"
description: |-
  Manages an AWS Single Sign-On permission set
---

# Resource: aws_sso_permission_set

Provides an AWS Single Sign-On Permission Set resource

## Example Usage

```hcl
data "aws_sso_instance" "selected" { }

data "aws_iam_policy_document" "example" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "arn:aws:s3:::*",
    ]
  }
}

resource "aws_sso_permission_set" "example" {
  name                = "Example"
  description         = "An example"
  instance_arn        = data.aws_sso_instance.selected.arn
  session_duration    = "PT1H"
  relay_state         = "https://console.aws.amazon.com/console/home"
  inline_policy       = data.aws_iam_policy_document.example.json
  managed_policy_arns = [
  "arn:aws:iam::aws:policy/ReadOnlyAccess",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `instance_arn` - (Required) The AWS ARN associated with the AWS Single Sign-On Instance.
* `name` - (Required) The name of the AWS Single Sign-On Permission Set.
* `description` - (Optional) The description of the AWS Single Sign-On Permission Set.
* `session_duration` - (Optional) The session duration of the AWS Single Sign-On Permission Set in the ISO-8601 standard. The default value is `PT1H`. 
* `relay_state` - (Optional) The relay state of AWS Single Sign-On Permission Set. 
* `inline_policy` - (Optional) The inline policy of the AWS Single Sign-On Permission Set.
* `managed_policy_arns` - (Optional) The managed policies attached to the AWS Single Sign-On Permission Set.
* `tags` - (Optional) Key-value map of resource tags.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The arn of the AWS Single Sign-On Permission Set.
* `arn` - The arn of the AWS Single Sign-On Permission Set.
* `created_date` - The created date of the AWS Single Sign-On Permission Set.

## Import

`aws_sso_permission_set` can be imported by using the AWS Single Sign-On Permission Set Resource Name (ARN), e.g.

```
$ terraform import aws_sso_permission_set.example arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk
```
