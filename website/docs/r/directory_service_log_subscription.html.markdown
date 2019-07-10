---
layout: "aws"
page_title: "AWS: aws_directory_service_log_subscription"
sidebar_current: "docs-aws-resource-directory-service-log-subscription"
description: |-
  Provides a Log subscription for AWS Directory Service that pushes logs to cloudwatch.
---

# Resource: aws_directory_service_log_subscription

Provides a Log subscription for AWS Directory Service that pushes logs to cloudwatch.

## Example Usage

```hcl
resource "aws_directory_service_log_subscription" "msad-cwl" {
  directory_id = "d-1234567890"
  log_group_name = "/aws/directoryservice/msad"
}

data "aws_iam_policy_document" "ad-log-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]

    principals {
      identifiers = ["ds.amazonaws.com"]
      type = "Service"
    }

    resources = ["/aws/directoryservice/msad"]

    effect = "Allow"
  }
}

resource "aws_cloudwatch_log_resource_policy" "ad-log-policy" {
  policy_document = "${data.aws_iam_policy_document.ad-log-policy.json}"
  policy_name = "ad-log-policy"
}
```

## Argument Reference

The following arguments are supported:

* `directory_id` - (Required) The id of directory.
* `log_group_name` - (Required) Name of the cloudwatch log group to which the logs should be published. The log group should be already created and the directory service principal should be provided with required permission to create stream and publish logs. Changing this value would delete the current subscription and create a new one. A directory can only have one log subscription at a time.

## Import

Conditional forwarders can be imported using the directory id and remote_domain_name, e.g.

```
$ terraform import aws_directory_service_log_subscription.msad d-1234567890
```
