---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_bus_policy"
description: |-
  Provides a resource to create an EventBridge policy to support cross-account events.
---

# Resource: aws_cloudwatch_event_bus_policy

Provides a resource to create an EventBridge resource policy to support cross-account events.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

~> **Note:** The EventBridge bus policy resource  (`aws_cloudwatch_event_bus_policy`) is incompatible with the EventBridge permission resource (`aws_cloudwatch_event_permission`) and will overwrite permissions.

## Example Usage

### Account Access

```hcl
data "aws_iam_policy_document" "test" {
  statement {
    sid    = "DevAccountAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      "arn:aws:events:eu-west-1:123456789012:event-bus/default"
    ]

    principals {
      type        = "AWS"
      identifiers = ["123456789012"]
    }
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.test.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
```

### Organization Access

```hcl
data "aws_iam_policy_document" "test" {
  statement {
    sid    = "OrganizationAccess"
    effect = "Allow"
    actions = [
      "events:DescribeRule",
      "events:ListRules",
      "events:ListTargetsByRule",
      "events:ListTagsForResource",
    ]
    resources = [
      "arn:aws:events:eu-west-1:123456789012:rule/*",
      "arn:aws:events:eu-west-1:123456789012:event-bus/default"
    ]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"
      values   = [aws_organizations_organization.example.id]
    }
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.test.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
```

### Multiple Statements

```hcl
data "aws_iam_policy_document" "test" {

  statement {
    sid    = "DevAccountAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      "arn:aws:events:eu-west-1:123456789012:event-bus/default"
    ]

    principals {
      type        = "AWS"
      identifiers = ["123456789012"]
    }
  }

  statement {
    sid    = "OrganizationAccess"
    effect = "Allow"
    actions = [
      "events:DescribeRule",
      "events:ListRules",
      "events:ListTargetsByRule",
      "events:ListTagsForResource",
    ]
    resources = [
      "arn:aws:events:eu-west-1:123456789012:rule/*",
      "arn:aws:events:eu-west-1:123456789012:event-bus/default"
    ]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"
      values   = [aws_organizations_organization.example.id]
    }
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.test.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
```

## Argument Reference

This resource supports the following arguments:

* `policy` - (Required) The text of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `event_bus_name` - (Optional) The name of the event bus to set the permissions on.
  If you omit this, the permissions are set on the `default` event bus.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the EventBridge event bus.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an EventBridge policy using the `event_bus_name`. For example:

```terraform
import {
  to = aws_cloudwatch_event_bus_policy.DevAccountAccess
  id = "example-event-bus"
}
```

Using `terraform import`, import an EventBridge policy using the `event_bus_name`. For example:

```console
% terraform import aws_cloudwatch_event_bus_policy.DevAccountAccess example-event-bus
```
