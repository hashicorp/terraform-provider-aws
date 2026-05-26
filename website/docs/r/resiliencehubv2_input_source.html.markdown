---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_input_source"
description: |-
  Terraform resource for managing an AWS Resilience Hub V2 Input Source.
---

# Resource: aws_resiliencehubv2_input_source

Terraform resource for managing an AWS Resilience Hub V2 Input Source.

An input source defines where Resilience Hub discovers AWS resources for a service. Supported source types include CloudFormation stacks, Terraform state files (stored in S3), and EKS clusters. Exactly one of `cfn_stack_arn`, `tf_state_file_url`, or `eks_cluster_arn` must be specified.

~> **Note:** This resource does not support in-place updates. Any change to the resource configuration will destroy and recreate the input source.

~> **Note:** The referenced resources (CloudFormation stacks, S3 state files, EKS clusters) must exist before creating the input source. Use `depends_on` to ensure proper ordering.

## Example Usage

### CloudFormation Stack

```hcl
resource "aws_resiliencehubv2_input_source" "example" {
  service_arn   = aws_resiliencehubv2_service.example.arn
  cfn_stack_arn = "arn:aws:cloudformation:us-west-2:123456789012:stack/my-stack/abc123"
}
```

### Terraform State File

```hcl
resource "aws_resiliencehubv2_input_source" "example" {
  service_arn      = aws_resiliencehubv2_service.example.arn
  tf_state_file_url = "s3://my-bucket/terraform.tfstate"
}
```

### EKS Cluster

```hcl
resource "aws_resiliencehubv2_input_source" "example" {
  service_arn     = aws_resiliencehubv2_service.example.arn
  eks_cluster_arn = "arn:aws:eks:us-west-2:123456789012:cluster/my-cluster"
  eks_namespaces  = ["default", "production"]
}
```

## Argument Reference

The following arguments are required:

* `service_arn` - (Required) ARN of the service this input source belongs to. Changing this value requires creating a new resource.

Exactly one of the following arguments is required:

* `cfn_stack_arn` - (Optional) ARN of a CloudFormation stack to use as an input source. Changing this value requires creating a new resource.
* `eks_cluster_arn` - (Optional) ARN of an EKS cluster to use as an input source. Changing this value requires creating a new resource.
* `tf_state_file_url` - (Optional) S3 URL of a Terraform state file to use as an input source. Changing this value requires creating a new resource.

The following arguments are optional:

* `eks_namespaces` - (Optional) List of Kubernetes namespaces to include when using an EKS cluster input source. Changing this value requires creating a new resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Composite identifier in the format `service_arn,input_source_id`.
* `input_source_id` - Unique identifier of the input source.

## Import

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
