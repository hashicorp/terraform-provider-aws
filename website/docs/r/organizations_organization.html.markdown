---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_organization"
description: |-
  Provides a resource to create an organization.
---

# Resource: aws_organizations_organization

Provides a resource to create an organization.

!> **WARNING:** When migrating from a `feature_set` of `CONSOLIDATED_BILLING` to `ALL`, the Organization account owner will received an email stating the following: "You started the process to enable all features for your AWS organization. As part of that process, all member accounts that joined your organization by invitation must approve the change. You don’t need approval from member accounts that you directly created from within your AWS organization." After all member accounts have accepted the invitation, the Organization account owner must then finalize the changes via the [AWS Console](https://console.aws.amazon.com/organizations/home#/organization/settings/migration-progress). Until these steps are performed, Terraform will perpetually show a difference, and the `DescribeOrganization` API will continue to show the `FeatureSet` as `CONSOLIDATED_BILLING`. See the [AWS Organizations documentation](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_org_support-all-features.html) for more information.

!> **WARNING:** [Warning from the AWS Docs](https://docs.aws.amazon.com/organizations/latest/APIReference/API_EnableAWSServiceAccess.html): "We recommend that you enable integration between AWS Organizations and the specified AWS service by using the console or commands that are provided by the specified service. Doing so ensures that the service is aware that it can create the resources that are required for the integration. How the service creates those resources in the organization's accounts depends on that service. For more information, see the documentation for the other AWS service."

## Example Usage

```terraform
resource "aws_organizations_organization" "org" {
  aws_service_access_principals = [
    "cloudtrail.amazonaws.com",
    "config.amazonaws.com",
  ]

  feature_set = "ALL"
}
```

## Argument Reference

This resource supports the following arguments:

* `aws_service_access_principals` - (Optional) List of AWS service principal names for which you want to enable integration with your organization. This is typically in the form of a URL, such as service-abbreviation.amazonaws.com. Organization must have `feature_set` set to `ALL`. Some services do not support enablement via this endpoint, see [warning in aws docs](https://docs.aws.amazon.com/organizations/latest/APIReference/API_EnableAWSServiceAccess.html).
* `enabled_policy_types` - (Optional) List of Organizations policy types to enable in the Organization Root. Organization must have `feature_set` set to `ALL`. For additional information about valid policy types (e.g., `AISERVICES_OPT_OUT_POLICY`, `BACKUP_POLICY`, `BEDROCK_POLICY`, `CHATBOT_POLICY`, `DECLARATIVE_POLICY_EC2`, `INSPECTOR_POLICY`, `RESOURCE_CONTROL_POLICY`, `S3_POLICY`, `SECURITYHUB_POLICY`, `SERVICE_CONTROL_POLICY`, `TAG_POLICY` and `UPGRADE_ROLLOUT_POLICY`), see the [AWS Organizations API Reference](https://docs.aws.amazon.com/organizations/latest/APIReference/API_EnablePolicyType.html). To enable `INSPECTOR_POLICY`, `aws_service_access_principals` must include `inspector2.amazonaws.com`. To enable `SECURITYHUB_POLICY`, `aws_service_access_principals` must include `securityhub.amazonaws.com`.
* `feature_set` - (Optional) Specify `ALL` (default) or `CONSOLIDATED_BILLING`.
* `return_organization_only` - (Optional) Return (as attributes) only the results of the [`DescribeOrganization`](https://docs.aws.amazon.com/organizations/latest/APIReference/API_DescribeOrganization.html) API to avoid [API limits](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_reference_limits.html#throttling-limits). When configured to `true` only the `arn`, `feature_set`, `master_account_arn`, `master_account_email` and `master_account_id` attributes will be returned. All others will be empty. Default: `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `accounts` - List of organization accounts including the master account. For a list excluding the master account, see the `non_master_accounts` attribute. All elements have these attributes:
    * `arn` - ARN of the account.
    * `email` - Email of the account.
    * `id` - Identifier of the account.
    * `joined_method` - Method by which the account joined the organization.
    * `joined_timestamp` - Date the account became a part of the organization.
    * `name` - Name of the account.
    * `state` - State of the account.
    * `status` - (**Deprecated** use `state` instead) Status of the account.
* `arn` - ARN of the organization.
* `id` - Identifier of the organization.
* `master_account_arn` - ARN of the master account.
* `master_account_email` - Email address of the master account.
* `master_account_id` - Identifier of the master account.
* `master_account_name` - Name of the master account.
* `non_master_accounts` - List of organization accounts excluding the master account. For a list including the master account, see the `accounts` attribute. All elements have these attributes:
    * `arn` - ARN of the account.
    * `email` - Email of the account.
    * `id` - Identifier of the account.
    * `joined_method` - Method by which the account joined the organization.
    * `joined_timestamp` - Date the account became a part of the organization.
    * `name` - Name of the account.
    * `state` - State of the account.
    * `status` - (**Deprecated** use `state` instead) Status of the account.
* `roots` - List of organization roots. All elements have these attributes:
    * `arn` - ARN of the root.
    * `id` - Identifier of the root.
    * `name` - Name of the root.
    * `policy_types` - List of policy types enabled for this root. All elements have these attributes:
        * `name` - Name of the policy type.
        * `status` - Status of the policy type as it relates to the associated root.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_organizations_organization.example
  identity = {
    id = "o-1234567"
  }
}

resource "aws_organizations_organization" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the AWS Organizations organization.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the AWS organization using the `id`. For example:

```terraform
import {
  to = aws_organizations_organization.example
  id = "o-1234567"
}
```

Using `terraform import`, import the AWS organization using the `id`. For example:

```console
% terraform import aws_organizations_organization.example o-1234567
```
