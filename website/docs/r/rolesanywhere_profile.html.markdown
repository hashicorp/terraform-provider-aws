---
subcategory: "Roles Anywhere"
layout: "aws"
page_title: "AWS: aws_rolesanywhere_profile"
description: |-
  Provides a Roles Anywhere Profile resource
---

# Resource: aws_rolesanywhere_profile

Terraform resource for managing a Roles Anywhere Profile.

## Example Usage

```terraform
resource "aws_iam_role" "test" {
  name = "test"
  path = "/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = [
        "sts:AssumeRole",
        "sts:TagSession",
        "sts:SetSourceIdentity"
      ]
      Principal = {
        Service = "rolesanywhere.amazonaws.com",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_rolesanywhere_profile" "test" {

  name      = "example"
  role_arns = [aws_iam_role.test.arn]
}
```

## Argument Reference

This resource supports the following arguments:

* `duration_seconds` - (Optional) The number of seconds the vended session credentials are valid for. Defaults to 3600.
* `enabled` - (Optional) Whether or not the Profile is enabled.
* `managed_policy_arns` - (Optional) A list of managed policy ARNs that apply to the vended session credentials.
* `name` - (Required) The name of the Profile.
* `require_instance_properties` - (Optional) Specifies whether instance properties are required in [CreateSession](https://docs.aws.amazon.com/rolesanywhere/latest/APIReference/API_CreateSession.html) requests with this profile.
* `role_arns` - (Required) A list of IAM roles that this profile can assume
* `session_policy` - (Optional) A session policy that applies to the trust boundary of the vended session credentials.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Profile
* `id` - The Profile ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_rolesanywhere_profile` using its `id`. For example:

```terraform
import {
  to = aws_rolesanywhere_profile.example
  id = "db138a85-8925-4f9f-a409-08231233cacf"
}
```

Using `terraform import`, import `aws_rolesanywhere_profile` using its `id`. For example:

```console
% terraform import aws_rolesanywhere_profile.example db138a85-8925-4f9f-a409-08231233cacf
```
