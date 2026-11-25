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
* `role_arns` - (Optional) A list of IAM roles that this profile can assume
* `session_policy` - (Optional) A session policy that applies to the trust boundary of the vended session credentials.
* `accept_role_session_name` - (Optional) Whether or not a custom role session name is accepted.
* `tip_configuration` - (Optional) Trusted identity propagation (TIP) settings for the profile. When TIP is enabled, sessions from the profile include the identity from the X.509 certificate in `aws:OriginIdentity` condition keys. See [`tip_configuration`](#tip_configuration) below. To turn TIP off, set `status` to `DISABLED`; removing the block from the configuration does not change the setting in AWS.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### tip_configuration

* `status` - (Required) Whether trusted identity propagation is used for the profile. Valid values: `ENABLED`, `DISABLED`.
* `subject_source` - (Optional) The certificate field whose value becomes the `aws:OriginIdentity/Subject` condition key. Valid values: `x509Subject/CN`, `x509SAN/URI`. If not specified, AWS uses `x509Subject/CN`. The profile's attribute mappings must include a rule for this field.
* `subject_type` - (Optional) Whether the certificates for the profile identify workloads or people. Valid values: `workload`, `user`. If not specified, sessions don't include the `aws:OriginIdentity/SubjectType` condition key.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Profile
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
