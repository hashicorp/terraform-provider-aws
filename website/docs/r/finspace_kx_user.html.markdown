---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_user"
description: |-
  Terraform resource for managing an AWS FinSpace Kx User.
---

# Resource: aws_finspace_kx_user

Terraform resource for managing an AWS FinSpace Kx User.

## Example Usage

### Basic Usage

```terraform
resource "aws_kms_key" "example" {
  description             = "Example KMS Key"
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "example" {
  name       = "my-tf-kx-environment"
  kms_key_id = aws_kms_key.example.arn
}

resource "aws_iam_role" "example" {
  name = "example-role"

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
}

resource "aws_finspace_kx_user" "example" { 
  name           = "my-tf-kx-user"
  environment_id = aws_finspace_kx_environment.example.id
  iam_role       = aws_iam_role.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) A unique identifier for the user.
* `environment_id` - (Required) Unique identifier for the KX environment.
* `iam_role` - (Required) IAM role ARN to be associated with the user.

The following arguments are optional:

* `tags` - (Optional) List of key-value pairs to label the KX user, such as first name and last name. You can add up to 50 tags to a user.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) identifier of the KX user.
* `tags_all` - Map of tags assigned to the resource.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)
