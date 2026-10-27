---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_resource_policy"
description: |-
  Manages the resource policy of an AWS Continuum resource.
---

# Resource: aws_continuum_resource_policy

Manages the resource policy attached to an AWS Continuum resource, such as a collector. Removing the resource removes the policy.

## Example Usage

```terraform
resource "aws_continuum_resource_policy" "example" {
  resource_arn = aws_continuum_collector.example.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = "arn:aws:iam::123456789012:root" }
      Action    = "continuum:GetCollectorDataAccessAuthorization"
      Resource  = aws_continuum_collector.example.arn
    }]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `policy` - (Required) JSON policy document.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_arn` - (Required, Forces new resource) ARN of the Continuum resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ARN of the resource.
* `revision_id` - Revision of the policy. Updates send it as the expected revision.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_continuum_resource_policy.example
  identity = {
    resource_arn = "arn:aws:continuum:us-east-1:123456789012:collectors/01a17e22728b7d25b3a2294e0ca5c35f"
  }
}

resource "aws_continuum_resource_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `resource_arn` (String) ARN of the Continuum resource.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_continuum_resource_policy` using the ID. For example:

```terraform
import {
  to = aws_continuum_resource_policy.example
  id = "arn:aws:continuum:us-east-1:123456789012:collectors/01a17e22728b7d25b3a2294e0ca5c35f"
}
```

Using `terraform import`, import `aws_continuum_resource_policy` using the ID. For example:

```console
% terraform import aws_continuum_resource_policy.example arn:aws:continuum:us-east-1:123456789012:collectors/01a17e22728b7d25b3a2294e0ca5c35f
```

