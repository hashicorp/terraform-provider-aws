---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user_ssh_key"
description: |-
  Get information on a SSH public key associated with the specified IAM user.
---

# Data Source: aws_iam_user_ssh_key

Use this data source to get information about a SSH public key associated with the specified IAM user.

## Example Usage

```terraform
data "aws_iam_user_ssh_key" "example" {
  encoding          = "SSH"
  ssh_public_key_id = "APKARUZ32GUTKIGARLXE"
  username          = "test-user"
}
```

## Argument Reference

* `encoding` - (Required) Specifies the public key encoding format to use in the response. To retrieve the public key in ssh-rsa format, use `SSH`. To retrieve the public key in PEM format, use `PEM`.
* `ssh_public_key_id` - (Required) Unique identifier for the SSH public key.
* `username` - (Required) Name of the IAM user associated with the SSH public key.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `fingerprint` - MD5 message digest of the SSH public key.
* `public_key` - SSH public key.
* `status` - Status of the SSH public key. Active means that the key can be used for authentication with an CodeCommit repository. Inactive means that the key cannot be used.
