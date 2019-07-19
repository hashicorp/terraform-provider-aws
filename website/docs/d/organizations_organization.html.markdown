---
layout: "aws"
page_title: "AWS: aws_organizations_organization"
sidebar_current: "docs-aws-datasource-organizations-organization"
description: |-
  Get information the organization that the user's account belongs to
---

# Data Source: aws_organizations_organization

Get information the organization that the user's account belongs to

## Example Usage

### SNS topic that can be interacted by the organization only

```hcl
data "aws_organizations_organization" "my_org" {}

resource "aws_sns_topic" "sns_topic" {
  name = "my-sns-topic"
}

resource "aws_sns_topic_policy" "sns_topic_policy" {
  arn = "${aws_sns_topic.sns_topic.arn}"

  policy = "${data.aws_iam_policy_document.sns_topic_policy.json}"
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
        "${data.aws_organizations_organization.my_org.id}",
      ]
    }

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    resources = [
      "${aws_sns_topic.sns_topic.arn}",
    ]
  }
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the organization.
* `available_policy_types` - A list of policy types that are enabled for this organization.
* `feature_set` - The FeatureSet of the organization.
* `id` - The ID of the organization. 
* `master_account_arn` - The Amazon Resource Name (ARN) of the account that is designated as the master account for the organization.
* `master_account_email` - The email address that is associated with the AWS account that is designated as the master account for the organization.
* `master_account_id` - The unique identifier (ID) of the master account of an organization.

The policy type object contains the following attributes:

* `type` - The name of the policy type
* `status` - The status of the policy type as it relates to the associated root.

[1]: https://docs.aws.amazon.com/organizations/latest/APIReference/API_DescribeOrganization.html#API_DescribeOrganization_ResponseSyntax
