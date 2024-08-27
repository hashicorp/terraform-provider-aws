---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_ssh_key"
description: |-
  Provides a AWS Transfer SSH Public Key resource.
---

# Resource: aws_transfer_ssh_key

Provides a AWS Transfer User SSH Key resource.

## Example Usage

```terraform
resource "tls_private_key" "example" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_transfer_ssh_key" "example" {
  server_id = aws_transfer_server.example.id
  user_name = aws_transfer_user.example.user_name
  body      = trimspace(tls_private_key.example.public_key_openssh)
}

resource "aws_transfer_server" "example" {
  identity_provider_type = "SERVICE_MANAGED"

  tags = {
    NAME = "tf-acc-test-transfer-server"
  }
}

resource "aws_transfer_user" "example" {
  server_id = aws_transfer_server.example.id
  user_name = "tftestuser"
  role      = aws_iam_role.example.arn

  tags = {
    NAME = "tftestuser"
  }
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["transfer.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "tf-test-transfer-user-iam-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "example" {
  statement {
    sid       = "AllowFullAccesstoS3"
    effect    = "Allow"
    actions   = ["s3:*"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "example" {
  name   = "tf-test-transfer-user-iam-policy"
  role   = aws_iam_role.example.id
  policy = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `server_id` - (Requirement) The Server ID of the Transfer Server (e.g., `s-12345678`)
* `user_name` - (Requirement) The name of the user account that is assigned to one or more servers.
* `body` - (Requirement) The public key portion of an SSH key pair.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transfer SSH Public Key using the `server_id` and `user_name` and `ssh_public_key_id` separated by `/`. For example:

```terraform
import {
  to = aws_transfer_ssh_key.bar
  id = "s-12345678/test-username/key-12345"
}
```

Using `terraform import`, import Transfer SSH Public Key using the `server_id` and `user_name` and `ssh_public_key_id` separated by `/`. For example:

```console
% terraform import aws_transfer_ssh_key.bar s-12345678/test-username/key-12345
```
