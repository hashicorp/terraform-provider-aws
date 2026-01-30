---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_organization"
description: |-
  Get information about the organization that the users account belongs to.
---

# Data Source: aws_organizations_organization

Get information about the organization that the users account belongs to.

## Example Usage

### Basic Usage

```terraform
data "aws_organizations_organization" "example" {}

output "account_ids" {
  value = data.aws_organizations_organization.example.accounts[*].id
}
```

### Limit SNS Topic Access to an Organization

```terraform
data "aws_organizations_organization" "example" {}

resource "aws_sns_topic" "sns_topic" {
  name = "my-sns-topic"
}

resource "aws_sns_topic_policy" "sns_topic_policy" {
  arn = aws_sns_topic.sns_topic.arn

  policy = data.aws_iam_policy_document.sns_topic_policy.json
}

data "aws_iam_policy_document" "sns_topic_policy" {
  statement {
    effect = "Allow"

    actions = [
      "SNS:Subscribe",
      "SNS:Publish",
    ]

    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"

      values = [
        data.aws_organizations_organization.example.id,
      ]
    }

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    resources = [
      aws_sns_topic.sns_topic.arn,
    ]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `return_organization_only` - (Optional) Return (as attributes) only the results of the [`DescribeOrganization`](https://docs.aws.amazon.com/organizations/latest/APIReference/API_DescribeOrganization.html) API to avoid [API limits](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_reference_limits.html#throttling-limits). When configured to `true` only the `arn`, `feature_set`, `master_account_arn`, `master_account_email` and `master_account_id` attributes will be returned. All others will be empty. Default: `false`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the organization.
* `feature_set` - FeatureSet of the organization.
* `id` - ID of the organization.
* `master_account_arn` - ARN of the account that is designated as the master account for the organization.
* `master_account_email` - The email address that is associated with the AWS account that is designated as the master account for the organization.
* `master_account_id` - Unique identifier (ID) of the master account of an organization.
* `master_account_name` - Name of the master account of an organization.

### Master Account or Delegated Administrator Attribute Reference

If the account is the master account or a delegated administrator for the organization, the following attributes are also exported:

* `accounts` - List of organization accounts including the master account. For a list excluding the master account, see the `non_master_accounts` attribute. All elements have these attributes:
    * `arn` - ARN of the account.
    * `email` - Email of the account.
    * `id` - Identifier of the account.
    * `joined_method` - Method by which the account joined the organization.
    * `joined_timestamp` - Date the account became a part of the organization.
    * `name` - Name of the account.
    * `state` - State of the account.
    * `status` - (**Deprecated** use `state` instead) Status of the account.
* `aws_service_access_principals` - A list of AWS service principal names that have integration enabled with your organization. Organization must have `feature_set` set to `ALL`. For additional information, see the [AWS Organizations User Guide](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_integrate_services.html).
* `enabled_policy_types` - A list of Organizations policy types that are enabled in the Organization Root. Organization must have `feature_set` set to `ALL`. For additional information about valid policy types (e.g., `SERVICE_CONTROL_POLICY`), see the [AWS Organizations API Reference](https://docs.aws.amazon.com/organizations/latest/APIReference/API_EnablePolicyType.html).
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
