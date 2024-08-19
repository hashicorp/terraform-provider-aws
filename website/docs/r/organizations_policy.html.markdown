---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_policy"
description: |-
  Provides a resource to manage an AWS Organizations policy.
---

# Resource: aws_organizations_policy

Provides a resource to manage an [AWS Organizations policy](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies.html).

## Example Usage

```terraform
data "aws_iam_policy_document" "example" {
  statement {
    effect    = "Allow"
    actions   = ["*"]
    resources = ["*"]
  }
}

resource "aws_organizations_policy" "example" {
  name    = "example"
  content = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `content` - (Required) The policy content to add to the new policy. For example, if you create a [service control policy (SCP)](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scp.html), this string must be JSON text that specifies the permissions that admins in attached accounts can delegate to their users, groups, and roles. For more information about the SCP syntax, see the [Service Control Policy Syntax documentation](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_reference_scp-syntax.html) and for more information on the Tag Policy syntax, see the [Tag Policy Syntax documentation](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_example-tag-policies.html).
* `name` - (Required) The friendly name to assign to the policy.
* `description` - (Optional) A description to assign to the policy.
* `skip_destroy` - (Optional) If set to `true`, destroy will **not** delete the policy and instead just remove the resource from state. This can be useful in situations where the policies (and the associated attachment) must be preserved to meet the AWS minimum requirement of 1 attached policy.
* `type` - (Optional) The type of policy to create. Valid values are `AISERVICES_OPT_OUT_POLICY`, `BACKUP_POLICY`, `SERVICE_CONTROL_POLICY` (SCP), and `TAG_POLICY`. Defaults to `SERVICE_CONTROL_POLICY`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique identifier (ID) of the policy.
* `arn` - Amazon Resource Name (ARN) of the policy.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_organizations_policy` using the policy ID. For example:

```terraform
import {
  to = aws_organizations_policy.example
  id = "p-12345678"
}
```

Using `terraform import`, import `aws_organizations_policy` using the policy ID. For example:

```console
% terraform import aws_organizations_policy.example p-12345678
```
