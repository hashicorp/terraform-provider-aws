---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_collector"
description: |-
  Manages the AWS Continuum collector.
---

# Resource: aws_continuum_collector

Manages the AWS Continuum collector for an account and Region. The collector provisions an AWS Config service-linked recorder, a versioned S3 bucket and a KMS key that Continuum uses to gather infrastructure context. Each account can have one collector.

## Example Usage

```terraform
resource "aws_continuum_collector" "example" {}
```

## Argument Reference

This resource supports the following arguments:

* `kubernetes_cluster_role_arn` - (Optional) ARN of the IAM role that Continuum assumes to read Kubernetes cluster context.
* `kubernetes_cluster_role_external_id` - (Optional, Sensitive) External ID that Continuum presents when assuming `kubernetes_cluster_role_arn`. Continuum does not return this value, so it is not populated by import.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `type` - (Optional, Forces new resource) Collector type. Valid values are `AWS_COLLECTOR`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the collector.
* `config_recorder_arn` - ARN of the AWS Config recorder that the collector provisioned.
* `created_at` - Time the collector was created.
* `id` - ID of the collector.
* `status` - Collector status: `ACTIVE` or `PENDING_DELETION`.
* `updated_at` - Time the collector was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_continuum_collector.example
  identity = {
    id = "01a17e22728b7d25b3a2294e0ca5c35e"
  }
}

resource "aws_continuum_collector" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the collector.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_continuum_collector` using the ID. For example:

```terraform
import {
  to = aws_continuum_collector.example
  id = "ID_PLACEHOLDER"
}
```

Using `terraform import`, import `aws_continuum_collector` using the ID. For example:

```console
% terraform import aws_continuum_collector.example ID_PLACEHOLDER
```

