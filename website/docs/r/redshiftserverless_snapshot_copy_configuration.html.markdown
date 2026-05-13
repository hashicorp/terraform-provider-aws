---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_snapshot_copy_configuration"
description: |-
  Terraform resource for managing an AWS Redshift Serverless Snapshot Copy Configuration.
---

# Resource: aws_redshiftserverless_snapshot_copy_configuration

Terraform resource for managing an AWS Redshift Serverless Snapshot Copy Configuration.

## Example Usage

```terraform
provider "aws" {
  region = "us-east-1"
}

provider "aws" {
  alias  = "west"
  region = "us-west-2"
}

resource "aws_kms_key" "example" {
  provider = aws.west
}

resource "aws_redshiftserverless_namespace" "example" {
  namespace_name = "example-namespace"
}

resource "aws_redshiftserverless_snapshot_copy_configuration" "example" {
  namespace_name            = aws_redshiftserverless_namespace.example.namespace_name
  destination_kms_key_id    = aws_kms_key.example.arn
  destination_region        = "us-west-2"
  snapshot_retention_period = 7
}
```

## Argument Reference

The following arguments are required:

* `namespace_name` - (Required) Name of the namespace to copy snapshots from.
* `destination_region` - (Required) Region to copy snapshots to.

The following arguments are optional:

* `destination_kms_key_id` - (Optional) KMS Key to encrypt snapshots at destination.
* `snapshot_retention_period` - (Optional) Retention period of the snapshots at destination.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Snapshot Copy Configuration.
* `id` - Globally unique identifier for Snapshot Copy Configuration.
