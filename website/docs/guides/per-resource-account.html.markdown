---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Per-Resource Account"
description: |-
  Managing resources in multiple AWS accounts from one provider configuration.
---

# Per-Resource Account

Resources, data sources and ephemeral resources accept an optional top-level `account` argument. List resources and actions accept it in their `config` block. One provider configuration can then manage resources in many AWS accounts of the same partition, the way [Enhanced Region Support](enhanced-region-support.html) lets it manage resources in many Regions.

<!-- TOC depthFrom:2 depthTo:2 -->

- [Why use it](#why-use-it)
- [Example](#example)
- [How the provider reaches an account](#how-the-provider-reaches-an-account)
- [How `account` works](#how-account-works)
- [Migrating from one provider configuration per account](#migrating-from-one-provider-configuration-per-account)
- [Resources without `account`](#resources-without-account)
- [Limitations](#limitations)

<!-- /TOC -->

## Why use it

Each provider configuration runs as a separate provider process and must be written into the configuration up front. A configuration that manages the same baseline in every account of an organization otherwise needs one provider configuration, and one process, per account.

## Example

```terraform
provider "aws" {
  region = "eu-central-1"

  assume_role_in_account {
    role_name = "terraform"
  }
}

variable "accounts" {
  type = map(string) # account name => account ID
}

resource "aws_cloudwatch_log_group" "baseline" {
  for_each = var.accounts

  name              = "baseline"
  retention_in_days = 30
  account           = each.value
}
```

## How the provider reaches an account

A resource names the account it is in. The provider configuration says how to get there, in `assume_role_in_account` blocks:

```terraform
provider "aws" {
  # Every other account.
  assume_role_in_account {
    role_name = var.role_name
  }

  # These accounts.
  assume_role_in_account {
    account_ids = ["333333333333", "444444444444"]
    role_name   = "legacy-admin"
    external_id = var.external_id
  }
}
```

The `assume_role_in_account` block supports the following:

* `account_ids` - (Optional) AWS account IDs the block applies to. The block without `account_ids` applies to every other account. An account can be in only one block, and only one block can omit `account_ids`.
* `duration` - (Optional) Duration of the role session, such as `1h` or `30m`.
* `external_id` - (Optional) External ID to pass when assuming the role.
* `policy` - (Optional) IAM policy JSON that further restricts the session's permissions.
* `policy_arns` - (Optional) ARNs of IAM policies that further restrict the session's permissions.
* `role_name` - (Required) Name, with any path, of the IAM role to assume in the resource's account.
* `session_name` - (Optional) Name of the role session.
* `source_identity` - (Optional) Source identity of the role session.
* `tags` - (Optional) Session tags.
* `transitive_tag_keys` - (Optional) Keys of the session tags that are passed on to a chained role.

The role is `arn:<partition>:iam::<account>:role/<role_name>`. It is assumed from the provider's own credentials. If the provider configuration has an `assume_role` block, those are the credentials of that role, and AWS limits the resulting chained session to one hour.

The blocks are provider configuration and are not stored in state. A different role, for example a read-only role for `plan` and another for `apply`, changes no resource.

Use a separate provider configuration for anything else, such as different source credentials for some accounts.

## How `account` works

* A resource without `account` is created in the account of the provider's own credentials, which are used directly. The same goes for `account` set to that account. If the provider does not know its own account (`skip_requesting_account_id`), any `account`, also a recorded one, needs a block that applies to it, no account is recorded for resources without `account`, and adding `account` to such a resource replaces it.
* With another account, the role of the `assume_role_in_account` block that applies to it is assumed. If none applies, the plan fails, or the apply if the account is only known then.
* The account must be permitted by the provider's `allowed_account_ids` and `forbidden_account_ids`.
* The provider records the account of every resource in state, also when `account` is not set. A resource never changes account silently: if its recorded account is not the provider's own, `account` must be set, and the plan fails until it is.
* Changing the account replaces the resource: it is destroyed in the old account and created in the new one. An account that is not known until apply does the same for an existing resource.
* The provider's Region, endpoints, `default_tags`, `ignore_tags` and retry settings apply in every account.

## Migrating from one provider configuration per account

Because the account is recorded in state, resources that were created by a provider configuration for one account can be managed by a provider configuration for another, without changing them:

1. Upgrade the provider and run `terraform apply` (or `terraform apply -refresh-only`) with the existing provider configurations. This records each resource's account. Do not skip this step: a resource without a recorded account is looked for in the account of whichever provider configuration manages it next.
2. Point the resources at a single provider configuration that has an `assume_role_in_account` block for their accounts.
3. Set `account` on the resources that are not in that provider's own account. `terraform plan` fails for each resource that still needs it and names its account. Once set, the plan shows no changes.

A data source reads from the account in its configuration, not from one recorded earlier. Without `account` it reads from the provider's own account.

## Resources without `account`

The following already have a top-level `account` attribute with another meaning, and the argument has not been added to them:

* `aws_arn` data source
* `aws_opensearch_authorize_vpc_endpoint_access` resource
* `aws_redshift_endpoint_authorization` resource

## Limitations

* `terraform import` and `import` blocks import from the provider's own account. An import into a resource that sets `account` to another account imports the object of that name in the provider's account and then plans its replacement. To import a resource in another account, use a provider configuration for that account.
* While `account` is not yet known, planning uses the provider's own account. The few resources that call AWS while planning may fail to plan a new resource until the account is known.
* Ephemeral resources are renewed and closed with the provider's credentials.
* Tag policy enforcement uses the required tags of the provider's account.
