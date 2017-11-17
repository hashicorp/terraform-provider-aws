---
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_resource_policy"
sidebar_current: "docs-aws-resource-cloudwatch-log-resource-policy"
description: |-
  Provides a CloudWatch Logs Resource Policy.
---

# aws_cloudwatch_log_resource_policy

Provides a CloudWatch Logs Resource Policy resource.

## Example Usage

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_cloudwatch_log_resource_policy" "example" {
  policy_name = "example"
  policy_document = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Route53LogsToCloudWatchLogs",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "route53.amazonaws.com"
        ]
      },
      "Action": "logs:PutLogEvents",
      "Resource": "${aws_cloudwatch_log_group.example.arn}"
    }
  ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `policy_name` - (Optional) Name of the policy.
* `policy_document` - (Optional) The policy document. This is a JSON formatted string.
