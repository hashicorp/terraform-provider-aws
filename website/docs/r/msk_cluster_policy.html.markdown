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
resource "aws_msk_cluster_policy" "test" {
	  cluster_arn = aws_msk_cluster.test.arn
	  policy = jsonencode({
	    Version = "2012-10-17",
	    Statement = [{
	      Sid    = "testMskClusterPolicy"
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
	      Resource = aws_msk_cluster.test.arn
	    }]
	  })
}
```

## Argument Reference

The following arguments are required:

* `cluster_arn` - (Required) The Amazon Resource Name (ARN) that uniquely identifies the cluster.

* `policy` - (Required) Resource policy for cluster.

The following arguments are optional:

* `current_version` - (Optional) Current cluster policy version.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `current_version` - Resource policy version

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Managed Streaming for Kafka Cluster Policy using the `example_id_arg`. For example:

```terraform
import {
  to = aws_msk_cluster_policy.example
  id = "cluster_policy-id-12345678"
}
```

Using `terraform import`, import Managed Streaming for Kafka Cluster Policy using the `example_id_arg`. For example:

```console
% terraform import aws_msk_cluster_policy.example cluster_policy-id-12345678
```
