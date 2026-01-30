---
subcategory: "Multi-party Approval"
layout: "aws"
page_title: "AWS: aws_mpa_approval_team"
description: |-
  Manages an AWS Multi-party Approval Approval Team.
---

# Resource: aws_mpa_approval_team

Manages an AWS Multi-party Approval Approval Team.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_mpa_approval_team" "example" {
  name        = "example-approval-team"
  description = "Example approval team for multi-party approval"

  approval_strategy {
    m_of_n {
      min_approvals_required = 2
    }
  }

  approver {
    primary_identity_id         = "user-id-1"
    primary_identity_source_arn = "arn:aws:mpa:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:identity-source/example"
  }

  approver {
    primary_identity_id         = "user-id-2"
    primary_identity_source_arn = "arn:aws:mpa:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:identity-source/example"
  }

  policy {
    policy_arn = "arn:aws:mpa:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:policy/example-policy"
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the approval team. Forces new resource.
* `description` - (Required) Description of the approval team.
* `approval_strategy` - (Required) Approval strategy configuration block. See [approval_strategy](#approval_strategy) below.
* `approver` - (Required) One or more approver configuration blocks. See [approver](#approver) below.
* `policy` - (Required) One or more policy configuration blocks. Forces new resource. See [policy](#policy) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### approval_strategy

* `m_of_n` - (Required) M-of-N approval strategy configuration block. See [m_of_n](#m_of_n) below.

### m_of_n

* `min_approvals_required` - (Required) Minimum number of approvals required. Must be at least 1.

### approver

* `primary_identity_id` - (Required) Primary identity ID of the approver.
* `primary_identity_source_arn` - (Required) ARN of the primary identity source for the approver.

### policy

* `policy_arn` - (Required) ARN of the MPA policy to attach to the approval team.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Approval Team.
* `creation_time` - Time the approval team was created.
* `id` - ARN of the Approval Team.
* `last_update_time` - Time the approval team was last updated.
* `number_of_approvers` - Number of approvers in the team.
* `status` - Status of the approval team.
* `status_code` - Status code of the approval team.
* `status_message` - Status message of the approval team.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version_id` - Version ID of the approval team.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Multi-party Approval Approval Team using the ARN. For example:

```terraform
import {
  to = aws_mpa_approval_team.example
  id = "arn:aws:mpa:us-east-1:123456789012:approval-team/example"
}
```

Using `terraform import`, import Multi-party Approval Approval Team using the ARN. For example:

```console
% terraform import aws_mpa_approval_team.example arn:aws:mpa:us-east-1:123456789012:approval-team/example
```
