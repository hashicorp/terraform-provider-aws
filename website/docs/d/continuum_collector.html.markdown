---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_collector"
description: |-
  Provides details about the AWS Continuum collector.
---

# Data Source: aws_continuum_collector

Provides details about the AWS Continuum collector.

## Example Usage

```terraform
data "aws_continuum_collector" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) ID of the collector. Omit to look up the only collector in the account and Region.
* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the collector.
* `config_recorder_arn` - ARN of the AWS Config recorder.
* `created_at` - Time the collector was created.
* `kubernetes_cluster_role_arn` - ARN of the Kubernetes cluster role.
* `status` - Collector status.
* `type` - Collector type.
* `updated_at` - Time the collector was last updated.
