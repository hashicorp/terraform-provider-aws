---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_auth_policy"
description: |-
  Terraform data source for managing an AWS VPC Lattice Auth Policy.
---

# Data Source: aws_vpclattice_auth_policy

Terraform data source for managing an AWS VPC Lattice Auth Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_vpclattice_auth_policy" "test" {
  resource_identifier = aws_vpclattice_auth_policy.test.resource_identifier
}
```

## Argument Reference

The following arguments are required:

* `resource_identifier` - (Required) The ID or Amazon Resource Name (ARN) of the service network or service for which the policy is created.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `policy` - The auth policy. The policy string in JSON must not contain newlines or blank lines.
* `state` - The state of the auth policy. The auth policy is only active when the auth type is set to AWS_IAM. If you provide a policy, then authentication and authorization decisions are made based on this policy and the client's IAM policy. If the Auth type is NONE, then, any auth policy you provide will remain inactive.
