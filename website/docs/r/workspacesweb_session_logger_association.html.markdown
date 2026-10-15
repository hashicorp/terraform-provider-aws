---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_session_logger_association"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Session Logger Association.
---

# Resource: aws_workspacesweb_session_logger_association

Terraform resource for managing an AWS WorkSpaces Web Session Logger Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_portal" "example" {
  display_name = "example"
}

resource "aws_s3_bucket" "example" {
  bucket        = "example-session-logs"
  force_destroy = true
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["workspaces-web.amazonaws.com"]
    }
    actions = [
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.example.arn}/*"
    ]
  }
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.id
  policy = data.aws_iam_policy_document.example.json
}

resource "aws_workspacesweb_session_logger" "example" {
  display_name = "example"

  event_filter {
    all = {}
  }

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.example.id
      folder_structure = "Flat"
      log_file_format  = "Json"
    }
  }

  depends_on = [aws_s3_bucket_policy.example]
}

resource "aws_workspacesweb_session_logger_association" "example" {
  portal_arn         = aws_workspacesweb_portal.example.portal_arn
  session_logger_arn = aws_workspacesweb_session_logger.example.session_logger_arn
}
```

## Argument Reference

The following arguments are required:

* `portal_arn` - (Required) ARN of the web portal.
* `session_logger_arn` - (Required) ARN of the session logger.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Session Logger Association using the `session_logger_arn,portal_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_session_logger_association.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:sessionLogger/session_logger-id-12345678,arn:aws:workspaces-web:us-west-2:123456789012:portal/portal-id-12345678"
}
```

Using `terraform import`, import WorkSpaces Web Session Logger Association using the `session_logger_arn,portal_arn`. For example:

```console
% terraform import aws_workspacesweb_session_logger_association.example arn:aws:workspaces-web:us-west-2:123456789012:sessionLogger/session_logger-id-12345678,arn:aws:workspaces-web:us-west-2:123456789012:portal/portal-id-12345678
```
