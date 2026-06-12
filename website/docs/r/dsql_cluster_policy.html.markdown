---
subcategory: "DSQL"
layout: "aws"
page_title: "AWS: aws_dsql_cluster_policy"
description: |-
  Terraform resource for managing an Amazon Aurora DSQL Cluster Policy.
---

# Resource: aws_dsql_cluster_policy

Terraform resource for managing an Amazon Aurora DSQL Cluster resource-based policy.

~> **NOTE:** Aurora DSQL resource-based policies can grant access to principals within the same AWS account as the cluster. Cross-account access is not currently supported by Aurora DSQL resource-based policies.

~> **NOTE:** Aurora DSQL resource-based policy changes are eventually consistent and typically take effect within one minute.

## Example Usage

### Block Public Internet Access

```terraform
resource "aws_dsql_cluster" "example" {}

resource "aws_dsql_cluster_policy" "example" {
  identifier = aws_dsql_cluster.example.identifier

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyAccessFromOutsideVPC"
        Effect = "Deny"
        Principal = {
          AWS = "*"
        }
        Action = [
          "dsql:DbConnect",
          "dsql:DbConnectAdmin",
        ]
        Resource = "*"
        Condition = {
          Null = {
            "aws:SourceVpc" = "true"
          }
        }
      }
    ]
  })
}
```

This policy denies `dsql:DbConnect` and `dsql:DbConnectAdmin` requests from the public internet. It only checks whether the request came from a VPC. To limit access to a specific VPC, use `aws:SourceVpc` with `StringNotEquals`.

The calling principal still requires an identity-based IAM policy that allows the required Aurora DSQL actions on the cluster.

### Restrict Access to a Specific VPC

```terraform
resource "aws_dsql_cluster" "example" {}

resource "aws_dsql_cluster_policy" "example" {
  identifier = aws_dsql_cluster.example.identifier

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyAccessFromOtherVPCs"
        Effect = "Deny"
        Principal = {
          AWS = "*"
        }
        Action = [
          "dsql:DbConnect",
          "dsql:DbConnectAdmin",
        ]
        Resource = aws_dsql_cluster.example.arn
        Condition = {
          StringNotEquals = {
            "aws:SourceVpc" = aws_vpc.example.id
          }
        }
      }
    ]
  })
}
```

### Restrict Access to an AWS Organization

```terraform
resource "aws_dsql_cluster" "example" {}

resource "aws_dsql_cluster_policy" "example" {
  identifier = aws_dsql_cluster.example.identifier

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyAccessFromOutsideOrganization"
        Effect = "Deny"
        Principal = {
          AWS = "*"
        }
        Action = [
          "dsql:DbConnect",
          "dsql:DbConnectAdmin",
        ]
        Resource = aws_dsql_cluster.example.arn
        Condition = {
          StringNotEquals = {
            "aws:PrincipalOrgID" = "o-exampleorgid"
          }
        }
      }
    ]
  })
}
```

For more examples, including specific organizational units and multi-Region cluster policies, see the [Aurora DSQL resource-based policy examples](https://docs.aws.amazon.com/aurora-dsql/latest/userguide/rbp-examples.html).

## Argument Reference

This resource supports the following arguments:

* `bypass_policy_lockout_safety_check` - (Optional) Whether to bypass the policy lockout safety check. Setting this value to `true` increases the risk that the cluster becomes unmanageable. Defaults to `false`.
* `identifier` - (Required) Identifier of the Aurora DSQL Cluster.
* `policy` - (Required) Resource-based policy document as JSON.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the Aurora DSQL Cluster.
* `policy_version` - Version of the policy document.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `1m`)
* `update` - (Default `1m`)
* `delete` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Aurora DSQL Cluster Policies using the cluster `identifier`. For example:

```terraform
import {
  to = aws_dsql_cluster_policy.example
  id = "abcde1f234ghijklmnop5qr6st"
}
```

Using `terraform import`, import Aurora DSQL Cluster Policies using the cluster `identifier`. For example:

```console
% terraform import aws_dsql_cluster_policy.example abcde1f234ghijklmnop5qr6st
```
