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

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The arn of the permission set.
* `arn` - The arn of the permission set.
* `created_date` - The created date of the permission set.
* `description` - The description of the permission set.
* `session_duration` - The session duration of the permission set in the ISO-8601 standard.
* `relay_state` - The relay state of the permission set.
* `inline_policy` - The inline policy of the permission set.
* `managed_policy_arns` - The managed policies attached to the permission set.
* `tags` - The tags of the permission set.

## Import

`aws_sso_permission_set` can be imported by using the AWS Single Sign-On Permission Set Resource Name (ARN), e.g.

```
$ terraform import aws_sso_permission_set.example arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk
```
