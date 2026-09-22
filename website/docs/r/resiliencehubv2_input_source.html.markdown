---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_input_source"
description: |-
  Terraform resource for managing an AWS Resilience Hub V2 Input Source.
---

# Resource: aws_resiliencehubv2_input_source

Terraform resource for managing an AWS Resilience Hub V2 Input Source.

An input source defines where Resilience Hub discovers AWS resources for a service. Supported source types include CloudFormation stacks, Terraform state files (stored in S3), and EKS clusters.

~> **Note:** This resource does not support in-place updates. Any change to the resource configuration will destroy and recreate the input source.

## Example Usage

### CloudFormation Stack

```terraform
resource "aws_resiliencehubv2_input_source" "example" {
  service_arn = aws_resiliencehubv2_service.example.arn

  resource_configuration {
    cfn_stack_arn = "arn:aws:cloudformation:us-west-2:123456789012:stack/my-stack/abc123"
  }
}
```

### Terraform State File

```terraform
resource "aws_resiliencehubv2_input_source" "example" {
  service_arn = aws_resiliencehubv2_service.example.arn

  resource_configuration {
    tf_state_file_url = "s3://my-bucket/terraform.tfstate"
  }
}
```

### EKS Cluster

```terraform
resource "aws_resiliencehubv2_input_source" "example" {
  service_arn = aws_resiliencehubv2_service.example.arn

  resource_configuration {
    eks {
      cluster_arn = "arn:aws:eks:us-west-2:123456789012:cluster/my-cluster"
      namespaces  = ["default", "production"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `resource_configuration` - (Required) Resource configuration for an input source. See [`resource_configuration` Block](#resource_configuration-block) below.
* `service_arn` - (Required) ARN of the service this input source belongs to.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `resource_configuration` Block

The `resource_configuration` block supports exactly one of the following:

* `cfn_stack_arn` - (Optional) CloudFormation stack ARN.
* `design_file_s3_url` - (Optional) S3 URL.
* `eks` - (Optional) EKS configuration. See [`eks` Block](#eks-block) below.
* `resource_tag` - (Optional) Resource tags used for discovery. See [`resource_tag` Block](#resource_tag-block) below.
* `tf_state_file_url` - (Optional) S3 URL.

### `eks` Block

The `eks` block supports:

* `cluster_arn` - (Required) Cluster ARN.
* `namespaces` - (Required) List of Kubernetes namespaces within the EKS cluster.

### `resource_tag` Block

The `resource_tag` block supports:

* `key` - (Required) Tag key.
* `values` - (Required) List of tag values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `input_source_id` - Unique identifier of the input source.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_resiliencehubv2_input_source.example
  identity = {
    service_arn     = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123"
    input_source_id = "12345678-1234-1234-1234-123456789012"
  }
}

resource "aws_resiliencehubv2_input_source" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `service_arn` (String) ARN of the service this input source belongs to.
* `input_source_id` (String) Unique identifier of the input source.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub V2 Input Source using the `service_arn` and `input_source_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_resiliencehubv2_input_source.example
  id = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123,12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import Resilience Hub V2 Input Source using the `service_arn` and `input_source_id` separated by a comma (`,`). For example:

```console
% terraform import aws_resiliencehubv2_input_source.example arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123,12345678-1234-1234-1234-123456789012
```
