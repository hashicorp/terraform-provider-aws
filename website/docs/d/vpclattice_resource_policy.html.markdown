---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_resource_policy"
description: |-
  Terraform data source for managing an AWS VPC Lattice Resource Policy.
---

# Data Source: aws_vpclattice_resource_policy

Terraform data source for managing an AWS VPC Lattice Resource Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_vpclattice_resource_policy" "example" {
  resource_arn = aws_vpclattice_service_network.example.arn
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) Resource ARN of the resource for which a policy is retrieved.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `policy` - JSON-encoded string representation of the applied resource policy.
