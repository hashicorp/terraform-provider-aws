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
* `enabled_policy_types` - (Optional) List of Organizations policy types to enable in the Organization Root. Organization must have `feature_set` set to `ALL`. For additional information about valid policy types (e.g., `AISERVICES_OPT_OUT_POLICY`, `BACKUP_POLICY`, `SERVICE_CONTROL_POLICY`, and `TAG_POLICY`), see the [AWS Organizations API Reference](https://docs.aws.amazon.com/organizations/latest/APIReference/API_EnablePolicyType.html).
* `feature_set` - (Optional) Specify "ALL" (default) or "CONSOLIDATED_BILLING".

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `accounts` - List of organization accounts including the master account. For a list excluding the master account, see the `non_master_accounts` attribute. All elements have these attributes:
    * `arn` - ARN of the account
    * `email` - Email of the account
    * `id` - Identifier of the account
    * `name` - Name of the account
    * `status` - Current status of the account
* `arn` - ARN of the organization
* `id` - Identifier of the organization
* `master_account_arn` - ARN of the master account
* `master_account_email` - Email address of the master account
* `master_account_id` - Identifier of the master account
* `master_account_name` - Name of the master account
* `non_master_accounts` - List of organization accounts excluding the master account. For a list including the master account, see the `accounts` attribute. All elements have these attributes:
    * `arn` - ARN of the account
    * `email` - Email of the account
    * `id` - Identifier of the account
    * `name` - Name of the account
    * `status` - Current status of the account
* `roots` - List of organization roots. All elements have these attributes:
    * `arn` - ARN of the root
    * `id` - Identifier of the root
    * `name` - Name of the root
    * `policy_types` - List of policy types enabled for this root. All elements have these attributes:
        * `name` - The name of the policy type
        * `status` - The status of the policy type as it relates to the associated root

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the AWS organization using the `id`. For example:

```terraform
import {
  to = aws_organizations_organization.my_org
  id = "o-1234567"
}
```

Using `terraform import`, import the AWS organization using the `id`. For example:

```console
% terraform import aws_organizations_organization.my_org o-1234567
```
