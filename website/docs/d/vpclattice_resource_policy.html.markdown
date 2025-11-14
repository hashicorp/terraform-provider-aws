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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_arn` - (Required) Resource ARN of the resource for which a policy is retrieved.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `policy` - JSON-encoded string representation of the applied resource policy.
