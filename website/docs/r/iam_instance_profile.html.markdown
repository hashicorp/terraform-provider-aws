---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_instance_profile"
description: |-
  Provides an IAM instance profile.
---

# Resource: aws_iam_instance_profile

Provides an IAM instance profile.

~> **NOTE:** When managing instance profiles, remember that the `name` attribute must always be unique. This means that even if you have different `role` or `path` values, duplicating an existing instance profile `name` will lead to an `EntityAlreadyExists` error.

## Example Usage

```terraform
resource "aws_iam_instance_profile" "test_profile" {
  name = "test_profile"
  role = aws_iam_role.role.name
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "role" {
  name               = "test_role"
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}
```

## Argument Reference

The following arguments are optional:

* `name` - (Optional, Forces new resource) Name of the instance profile. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`. Can be a string of characters consisting of upper and lowercase alphanumeric characters and these special characters: `_`, `+`, `=`, `,`, `.`, `@`, `-`. Spaces are not allowed. The `name` must be unique, regardless of the `path` or `role`. In other words, if there are different `role` or `path` values but the same `name` as an existing instance profile, it will still cause an `EntityAlreadyExists` error.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `path` - (Optional, default "/") Path to the instance profile. For more information about paths, see [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html) in the IAM User Guide. Can be a string of characters consisting of either a forward slash (`/`) by itself or a string that must begin and end with forward slashes. Can include any ASCII character from the ! (\u0021) through the DEL character (\u007F), including most punctuation characters, digits, and upper and lowercase letters.
* `role` - (Optional) Name of the role to add to the profile.
* `tags` - (Optional) Map of resource tags for the IAM Instance Profile. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS to the instance profile.
* `create_date` - Creation timestamp of the instance profile.
* `id` - Instance profile's ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `unique_id` - [Unique ID][1] assigned by AWS.

  [1]: https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html#GUIDs

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Instance Profiles using the `name`. For example:

```terraform
import {
  to = aws_iam_instance_profile.test_profile
  id = "app-instance-profile-1"
}
```

Using `terraform import`, import Instance Profiles using the `name`. For example:

```console
% terraform import aws_iam_instance_profile.test_profile app-instance-profile-1
```
