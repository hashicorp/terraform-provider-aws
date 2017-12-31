---
layout: "aws"
page_title: "AWS: aws_codebuild_webhook"
sidebar_current: "docs-aws-resource-codebuild-webhook"
description: |-
  Provides a CodeBuild Webhook resource.
---

# aws_codebuild_webhook

Provides a CodeBuild Webhook resource.

## Example Usage

```hcl
resource "aws_iam_role" "codebuild_role" {
  name = "codebuild-role-"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codebuild.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "codebuild_policy" {
  name        = "codebuild-policy"
  path        = "/service-role/"
  description = "Policy used in trust relationship with CodeBuild"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "codebuild_policy_attachment" {
  name       = "codebuild-policy-attachment"
  policy_arn = "${aws_iam_policy.codebuild_policy.arn}"
  roles      = ["${aws_iam_role.codebuild_role.id}"]
}

resource "aws_codebuild_project" "foo" {
  name         = "test-project"
  description  = "test_codebuild_project"
  build_timeout      = "5"
  service_role = "${aws_iam_role.codebuild_role.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable {
      "name"  = "SOME_KEY1"
      "value" = "SOME_VALUE1"
    }

    environment_variable {
      "name"  = "SOME_KEY2"
      "value" = "SOME_VALUE2"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/mitchellh/packer.git"
  }

  tags {
    "Environment" = "Test"
  }
}

resource "aws_codebuild_webhook" "foo" {
  name = "${aws_codebuild_project.foo.name}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the build project.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the build project.
* `url` - The URL to the webhook.
