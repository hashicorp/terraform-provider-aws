---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_cluster_policy"
description: |-
  Terraform resource for managing an AWS Managed Streaming for Kafka Cluster Policy.
---
# Resource: aws_msk_cluster_policy

Terraform resource for managing an AWS Managed Streaming for Kafka Cluster Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_msk_cluster_policy" "example" {
  cluster_arn = aws_msk_cluster.example.arn

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Sid    = "ExampleMskClusterPolicy"
      Effect = "Allow"
      Principal = {
        "AWS" = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "kafka:Describe*",
        "kafka:Get*",
        "kafka:CreateVpcConnection",
        "kafka:GetBootstrapBrokers",
      ]
      Resource = aws_msk_cluster.example.arn
    }]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_arn` - (Required) The Amazon Resource Name (ARN) that uniquely identifies the cluster.
* `policy` - (Required) Resource policy for cluster.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Same as `cluster_arn`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Managed Streaming for Kafka Cluster Policy using the `cluster_arn. For example:

```terraform
import {
  to = aws_msk_cluster_policy.example
  id = "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"
}
```

Using `terraform import`, import Managed Streaming for Kafka Cluster Policy using the `cluster_arn`. For example:

```console
% terraform import aws_msk_cluster_policy.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
