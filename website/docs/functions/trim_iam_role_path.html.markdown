---
subcategory: ""
layout: "aws"
page_title: "AWS: trim_iam_role_path"
description: |-
  Trims the path prefix from an IAM role Amazon Resource Name (ARN).
---

# Function: trim_iam_role_path

Trims the path prefix from an IAM role Amazon Resource Name (ARN).
This function can be used when services require role ARNs to be passed without a path.

See the [AWS IAM documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/list_awsidentityandaccessmanagementiam.html#awsidentityandaccessmanagementiam-resources-for-iam-policies) for additional information on IAM role ARNs.

## Example Usage

```terraform
# result: arn:aws:iam::444455556666:role/example
output "example" {
  value = provider::aws::trim_iam_role_path("arn:aws:iam::444455556666:role/with/path/example")
}
```

## Signature

```text
trim_iam_role_path(arn string) string
```

## Arguments

1. `arn` (String) IAM role Amazon Resource Name (ARN).
