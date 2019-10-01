---
layout: "aws"
page_title: "AWS: aws_shield_drt_role_association"
description: |-
  Associates an IAM role with AWS Shield Advanced for DDoS Response Team (DRT) access
---

# Resource: aws_shield_drt_role_association

Associates an IAM Role for the AWS Shield Advanced DDoS Response Team (DRT) to use to access your account.

## Example Usage

### Associate role

```hcl
resource "aws_iam_role" "shield_drt" {
  name = "shield-drt-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": [
          "drt.shield.amazonaws.com"
        ]
       }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "shield_drt" {
  role       = "${aws_iam_role.shield_drt.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSShieldDRTAccessPolicy"
}

resource "aws_shield_drt_role_association" "shield_drt" {
  role_arn = "${aws_iam_role.shield_drt.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `role_arn` - (Required) The ARN of an IAM role that the DRT team will be given access to

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the DRT role association.

## Import

Shield DRT role association resources can be imported by specifying an arbitrary ID e.g.

```
$ terraform import aws_shield_drt_role_association.foo bar
```
