---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_role"
description: |-
  Provides an IAM role.
---

# Resource: aws_iam_role

Provides an IAM role.

~> *NOTE:* If policies are attached to the role via the [`aws_iam_policy_attachment` resource](/docs/providers/aws/r/iam_policy_attachment.html) and you are modifying the role `name` or `path`, the `force_detach_policies` argument must be set to `true` and applied before attempting the operation otherwise you will encounter a `DeleteConflict` error. The [`aws_iam_role_policy_attachment` resource (recommended)](/docs/providers/aws/r/iam_role_policy_attachment.html) does not have this requirement.

## Example Usage

```hcl
resource "aws_iam_role" "test_role" {
  name = "test_role"

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })

  tags = {
    tag-key = "tag-value"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resource) The name of the role. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `assume_role_policy` - (Required) The policy that grants an entity permission to assume the role.

~> **NOTE:** This `assume_role_policy` is very similar but slightly different than just a standard IAM policy and cannot use an `aws_iam_policy` resource.  It _can_ however, use an `aws_iam_policy_document` [data source](/docs/providers/aws/d/iam_policy_document.html), see example below for how this could work.

* `force_detach_policies` - (Optional) Specifies to force detaching any policies the role has before destroying it. Defaults to `false`.
* `path` - (Optional) The path to the role.
  See [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html) for more information.
* `description` - (Optional) The description of the role.

* `max_session_duration` - (Optional) The maximum session duration (in seconds) that you want to set for the specified role. If you do not specify a value for this setting, the default maximum of one hour is applied. This setting can have a value from 1 hour to 12 hours.
* `permissions_boundary` - (Optional) The ARN of the policy that is used to set the permissions boundary for the role.
* `tags` - Key-value map of tags for the IAM role
* `tags` - Key-value mapping of tags for the IAM role
* `managed_policy_arns` - (Optional) An exclusive set of IAM managed policy ARNs to attach to the IAM role. If the attribute is not used, the resource will not attach or detach the role's managed policies on the next `apply`. If the set is empty, all managed policies that are attached out of band, will be detached on the next `apply`. 

~> **NOTE:** The `managed_policy_arns` attribute, which provides an _exclusive_ set of managed policies for an IAM role, will conflict with using the `iam_role_policy_attachment` resource, which provides non-exclusive, managed policy-role attachment. See [`iam_role_policy_attachment`](/docs/providers/aws/r/iam_role_policy_attachment.html).

* `inline_policy` - (Optional) An exclusive set of IAM inline policies associated with the IAM role.  If the attribute is not used, the resource will not add or remove the role's inline policies on the next `apply`. If one empty `inline_policy` attribute is used, any inline policies that are added outside of Terraform will be removed on the next `apply`.

~> **NOTE:** The `inline_policy` attribute, which provides an _exclusive_ set of inline policies for an IAM role, will conflict with the `iam_role_policy` resource, which provides non-exclusive, inline policy-role association. See [`iam_role_policy`](/docs/providers/aws/r/iam_role_policy.html).

### inline_policy

The following arguments are supported:

* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://www.terraform.io/docs/providers/aws/guides/iam-policy-documents.html).
* `name` - (Optional) The name of the role policy. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. If both `name` and `name_prefix` are used, `name_prefix` will be ignored.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) specifying the role.
* `create_date` - The creation date of the IAM role.
* `description` - The description of the role.
* `id` - The name of the role.
* `name` - The name of the role.
* `unique_id` - The stable and unique string identifying the role.

## Example of Using Data Source for Assume Role Policy

```hcl
data "aws_iam_policy_document" "instance-assume-role-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "instance" {
  name               = "instance_role"
  path               = "/system/"
  assume_role_policy = data.aws_iam_policy_document.instance-assume-role-policy.json
}
```

## Example of Using Exclusive Inline Policies

This example will create an IAM role with two inline IAM policies. If a third policy were added out of band, on the next apply, that policy would be removed. If one of the two original policies were removed, out of band, on the next apply, the policy would be recreated.

```hcl
resource "aws_iam_role" "example" {
  name               = "yak_role"
  assume_role_policy = "${data.aws_iam_policy_document.instance_assume_role_policy.json}" # (not shown)

  inline_policy {
    name = "my_inline_policy"
    policy = <<EOF
{
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ],
  "Version": "2012-10-17"
}
EOF

  inline_policy {
    name_prefix = "inline_"
    policy = "${data.aws_iam_policy_document.inline_policy.json}"
  }
}

data "aws_iam_policy_document" "inline_policy" {
  statement {
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }
}
```

## Example of Not Allowing Inline Policies

This example will create an IAM role resource, which, on the next apply, will remove any out-of-band inline policies.

```hcl
resource "aws_iam_role" "example" {
  name               = "yak_role"
  assume_role_policy = "${data.aws_iam_policy_document.instance_assume_role_policy.json}" # (not shown)

  inline_policy {}
}
```

## Example of Using Exclusive Managed Policies

This example will create an IAM role resource with two attached managed IAM policies. If a third policy were attached out of band, on the next apply, that policy would be detached. If one of the two original policies were detached out of band, on the next apply, the policy would be recreated and re-attached.

```hcl
resource "aws_iam_role" "example" {
  name               = "yak_role"
  assume_role_policy = "${data.aws_iam_policy_document.instance_assume_role_policy.json}" # (not shown)

  managed_policy_arns = [
    "${aws_iam_policy.policy_one.arn}",
    "${aws_iam_policy.policy_two.arn}",
  ]
}

resource "aws_iam_policy" "policy_one" {
  name        = "managed_policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy_two" {
  name        = "managed_policy2"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:ListBucket",
        "s3:HeadBucket"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
```

## Example of Not Allowing Managed Policies

This example will create an IAM role resource, which, on the next apply, will detached all managed policies that were attached out of band.

```hcl
resource "aws_iam_role" "example" {
  name               = "yak_role"
  assume_role_policy = "${data.aws_iam_policy_document.instance_assume_role_policy.json}" # (not shown)

  managed_policy_arns = []
}
```

## Import

IAM Roles can be imported using the `name`, e.g.

```
$ terraform import aws_iam_role.developer developer_name
```
