---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_resource_policy"
description: |-
  Provides a Redshift Resource Policy resource.
---

# Resource: aws_redshift_resource_policy

Creates a new Amazon Redshift Resource Policy.

## Example Usage

```terraform
resource "aws_redshift_resource_policy" "example" {
  resource_arn = aws_redshift_cluster.example.cluster_namespace_arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:aws:iam::12345678901:root"
      }
      Action   = "redshift:CreateInboundIntegration"
      Resource = aws_redshift_cluster.example.cluster_namespace_arn
      Sid      = ""
    }]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the account to create or update a resource policy for.
* `policy` - (Required) The content of the resource policy being updated.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the account to create or update a resource policy for.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Resource Policies using the `resource_arn`. For example:

```terraform
import {
  to = aws_redshift_resource_policy.example
  id = "example"
}
```

Using `terraform import`, import Redshift Resource Policies using the `resource_arn`. For example:

```console
% terraform import aws_redshift_resource_policy.example example
```
