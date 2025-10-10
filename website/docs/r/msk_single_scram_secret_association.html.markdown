---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_single_scram_secret_association"
description: |-
  Associates a single SCRAM secret with a Managed Streaming for Kafka (MSK) cluster.
---

# Resource: aws_msk_single_scram_secret_association

Associates a single SCRAM secret with a Managed Streaming for Kafka (MSK) cluster.

## Example Usage

```terraform
resource "aws_msk_single_scram_secret_association" "example" {
  cluster_arn = aws_msk_cluster.example.arn
  secret_arn  = aws_secretsmanager_secret.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_arn` - (Required, Forces new resource) Amazon Resource Name (ARN) of the MSK cluster.
* `secret_arn` -  (Required, Forces new resource) AWS Secrets Manager secret ARN.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an MSK SCRAM Secret Association using the `cluster_arn` and `secret_arn`. For example:

```terraform
import {
  to = aws_msk_single_scram_secret_association.example
  id = "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3,arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456"
}
```

Using `terraform import`, import an MSK SCRAM Secret Association using the `cluster_arn` and `secret_arn`. For example:

```console
% terraform import aws_msk_single_scram_secret_association.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3,arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456
```
