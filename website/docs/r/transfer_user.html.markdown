---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_user"
description: |-
  Provides a AWS Transfer User resource.
---

# Resource: aws_transfer_user

Provides a AWS Transfer User resource. Managing SSH keys can be accomplished with the [`aws_transfer_ssh_key` resource](/docs/providers/aws/r/transfer_ssh_key.html).

~> **NOTE:** We suggest using [`jsonencode()`](https://developer.hashicorp.com/terraform/language/functions/jsonencode) or [`aws_iam_policy_document`](/docs/providers/aws/d/iam_policy_document.html) when assigning a value to `policy`. They seamlessly translate Terraform language into JSON, enabling you to maintain consistency within your configuration without the need for context switches. Also, you can sidestep potential complications arising from formatting discrepancies, whitespace inconsistencies, and other nuances inherent to JSON.

## Example Usage

```terraform
resource "aws_transfer_server" "foo" {
  identity_provider_type = "SERVICE_MANAGED"

  tags = {
    NAME = "tf-acc-test-transfer-server"
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

resource "aws_iam_role" "foo" {
  name               = "tf-test-transfer-user-iam-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "foo" {
  statement {
    sid       = "AllowFullAccesstoS3"
    effect    = "Allow"
    actions   = ["s3:*"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "foo" {
  name   = "tf-test-transfer-user-iam-policy"
  role   = aws_iam_role.foo.id
  policy = data.aws_iam_policy_document.foo.json
}

resource "aws_transfer_user" "foo" {
  server_id = aws_transfer_server.foo.id
  user_name = "tftestuser"
  role      = aws_iam_role.foo.arn

  home_directory_type = "LOGICAL"
  home_directory_mappings {
    entry  = "/test.pdf"
    target = "/bucket3/test-path/tftestuser.pdf"
  }
}
```

To restrict a user to their own home directory, use a `home_directory_mappings` block like the following:

```terraform
resource "aws_transfer_user" "example" {
  # ...

  home_directory_type = "LOGICAL"
  home_directory_mappings {
    entry  = "/"
    target = "/${aws_s3_bucket.foo.id}/$${Transfer:UserName}"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `home_directory` - (Optional) Landing directory (folder) for a user when they log in to the server using their SFTP client.  It should begin with a `/`.  The first item in the path is the name of the home bucket (accessible as `${Transfer:HomeBucket}` in the policy) and the rest is the home directory (accessible as `${Transfer:HomeDirectory}` in the policy). For example, `/example-bucket-1234/username` would set the home bucket to `example-bucket-1234` and the home directory to `username`.
* `home_directory_mappings` - (Optional) Logical directory mappings that specify what S3 paths and keys should be visible to your user and how you want to make them visible. See [`home_directory_mappings` Block](#home_directory_mappings-block) below.
* `home_directory_type` - (Optional) Type of landing directory (folder) you mapped for your users' home directory. Valid values are `PATH` and `LOGICAL`.
* `policy` - (Optional) IAM JSON policy document that scopes down user access to portions of their Amazon S3 bucket. IAM variables you can use inside this policy include `${Transfer:UserName}`, `${Transfer:HomeDirectory}`, and `${Transfer:HomeBucket}`. Since the IAM variable syntax matches Terraform's interpolation syntax, they must be escaped inside Terraform configuration strings (`$${Transfer:UserName}`).  These are evaluated on-the-fly when navigating the bucket.
* `posix_profile` - (Optional) Full POSIX identity, including user ID (Uid), group ID (Gid), and any secondary groups IDs (SecondaryGids), that controls your users' access to your Amazon EFS file systems. See [`posix_profile` Block](#posix_profile-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role` - (Required) ARN of an IAM role that allows the service to control your user’s access to your Amazon S3 bucket.
* `server_id` - (Required) Server ID of the Transfer Server (e.g., `s-12345678`)
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_name` - (Required) Name used for log in to your SFTP server.

### `home_directory_mappings` Block

* `entry` - (Required) Logical directory entry that appears to your user.
* `target` - (Required) Map target that maps the entry to an actual S3 path.

### `posix_profile` Block

* `gid` - (Required) POSIX group ID used for all EFS operations by this user.
* `secondary_gids` - (Optional) Secondary POSIX group IDs used for all EFS operations by this user.
* `uid` - (Required) POSIX user ID used for all EFS operations by this user.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of Transfer User
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transfer Users using the `server_id` and `user_name` separated by `/`. For example:

```terraform
import {
  to = aws_transfer_user.bar
  id = "s-12345678/test-username"
}
```

Using `terraform import`, import Transfer Users using the `server_id` and `user_name` separated by `/`. For example:

```console
% terraform import aws_transfer_user.bar s-12345678/test-username
```
