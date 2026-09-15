---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_replica"
description: |-
  Manages a Cognito User Pool multi-Region replica.
---

# Resource: aws_cognito_user_pool_replica

Manages a replica of a Cognito User Pool in an additional AWS Region for multi-Region replication. Multi-Region replication requires the primary user pool to be encrypted with a customer-managed, multi-Region KMS key (see the `key_configuration` argument of the [`aws_cognito_user_pool`](cognito_user_pool.html.markdown) resource).

## Example Usage

```terraform
resource "aws_kms_key" "example" {
  description             = "cognito-multi-region"
  deletion_window_in_days = 7
  multi_region            = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

# The KMS key must be replicated to the target region before creating the replica user pool.
resource "aws_kms_replica_key" "example" {
  provider        = aws.replica
  primary_key_arn = aws_kms_key.example.arn
  description     = "cognito-multi-region replica"

  deletion_window_in_days = 7

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Cognito to use this key"
        Effect = "Allow"
        Principal = {
          Service = "cognito-idp.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant",
          "kms:DescribeKey",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_cognito_user_pool" "example" {
  name = "example"

  key_configuration {
    key_type    = "CUSTOMER_MANAGED_KEY"
    kms_key_arn = aws_kms_key.example.arn
  }
}

resource "aws_cognito_user_pool_replica" "example" {
  user_pool_id = aws_cognito_user_pool.example.id
  region_name  = "us-west-2" # must differ from the primary user pool's Region
  status       = "ACTIVE"

  depends_on = [aws_kms_replica_key.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/cognito_identity.html). Defaults to the Region set in the provider configuration. This is the **primary** user pool's Region, not the replica's.
* `user_pool_id` - (Required) ID of the primary user pool to replicate. Changing this forces a new resource.
* `region_name` - (Required) AWS Region in which to create the replica. Changing this forces a new resource.
* `status` - (Optional) Desired status of the replica. Valid values are `ACTIVE` and `INACTIVE`. Defaults to `INACTIVE`.
* `tags` - (Optional) Map of tags to assign to the replica user pool. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `role` - Role of the replica. Either `PRIMARY` or `SECONDARY`.
* `user_pool_arn` - ARN of the replica user pool.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an `import` block to import Cognito User Pool Replicas using a comma-delimited string of `user_pool_id` and `region_name`. For example:

```terraform
import {
  to = aws_cognito_user_pool_replica.example
  id = "us-east-1_abc123,us-west-2"
}
```

Using `terraform import`, import Cognito User Pool Replicas using a comma-delimited string of `user_pool_id` and `region_name`. For example:

```console
% terraform import aws_cognito_user_pool_replica.example us-east-1_abc123,us-west-2
```
