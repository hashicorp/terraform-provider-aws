---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_egress_only_internet_gateway"
description: |-
  Provides details about a specific Egress-Only Internet Gateway.
---

# Data Source: aws_egress_only_internet_gateway

`aws_egress_only_internet_gateway` provides details about a specific Egress-Only Internet Gateway.

## Example Usage

### Basic Usage

```terraform
variable "eoig_id" {}

data "aws_egress_only_internet_gateway" "default" {
  egress_only_internet_gateway_id = var.eoig_id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `egress_only_internet_gateway_id` - (Optional) ID of the specific Egress-Only Internet Gateway to retrieve.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired Egress-Only Internet Gateway.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Egress-Only Internet Gateway.
* `owner_id` - ID of the AWS account that owns the egress-only internet gateway.
  
`attachments` are also exported with the following attributes, when there are relevants:
Each attachment supports the following:

* `state` - Current state of the attachment between the gateway and the VPC. Present only if a VPC is attached
* `vpc_id` - ID of an attached VPC.
