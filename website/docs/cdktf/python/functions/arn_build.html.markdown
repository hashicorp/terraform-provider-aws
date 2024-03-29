---
subcategory: ""
layout: "aws"
page_title: "AWS: arn_build"
description: |-
  Builds an ARN from its constituent parts.
---


<!-- Please do not edit this file, it is generated. -->
# Function: arn_build

~> Provider-defined function support is in technical preview and offered without compatibility promises until Terraform 1.8 is generally available.

Builds an ARN from its constituent parts.

See the [AWS documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference-arns.html) for additional information on Amazon Resource Names.

## Example Usage

```terraform
# result: arn:aws:iam::444455556666:role/example
output "example" {
  value = provider::aws::arn_build("aws", "iam", "", "444455556666", "role/example")
}
```

## Signature

```text
arn_build(partition string, service string, region string, account_id string, resource string) string
```

## Arguments

1. `partition` (String) Partition in which the resource is located. Supported partitions include `aws`, `aws-cn`, and `aws-us-gov`.
1. `service` (String) Service namespace.
1. `region` (String) Region code.
1. `account_id` (String) AWS account identifier.
1. `resource` (String) Resource section, typically composed of a resource type and identifier.

<!-- cache-key: cdktf-0.20.1 input-c5fb8fd6b6cf40ec3f3190f7e66547798964c21e1627e8dad5375ac37261a14a -->