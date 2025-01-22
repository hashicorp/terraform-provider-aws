---
subcategory: "DevOps Guru"
layout: "aws"
page_title: "AWS: aws_devopsguru_service_integration"
description: |-
  Terraform resource for managing an AWS DevOps Guru Service Integration.
---
# Resource: aws_devopsguru_service_integration

Terraform resource for managing an AWS DevOps Guru Service Integration.

~> To prevent unintentional deletion of account wide settings, destruction of this resource will only remove it from the Terraform state. To disable any configured settings, explicitly set the opt-in value to `DISABLED` and apply again before destroying.

## Example Usage

### Basic Usage

```terraform
resource "aws_devopsguru_service_integration" "example" {
  kms_server_side_encryption {
    opt_in_status = "ENABLED"
    type          = "AWS_OWNED_KMS_KEY"
  }

  logs_anomaly_detection {
    opt_in_status = "ENABLED"
  }

  ops_center {
    opt_in_status = "ENABLED"
  }
}
```

### Customer Managed KMS Key

```terraform
resource "aws_kms_key" "example" {}

resource "aws_devopsguru_service_integration" "example" {
  kms_server_side_encryption {
    kms_key_id    = aws_kms_key.test.arn
    opt_in_status = "ENABLED"
    type          = "CUSTOMER_MANAGED_KEY"
  }

  logs_anomaly_detection {
    opt_in_status = "DISABLED"
  }

  ops_center {
    opt_in_status = "DISABLED"
  }
}
```

## Argument Reference

The following arguments are required:

* `kms_server_side_encryption` - (Required) Information about whether DevOps Guru is configured to encrypt server-side data using KMS. See [`kms_server_side_encryption`](#kms_server_side_encryption-argument-reference) below.
* `logs_anomaly_detection` - (Required) Information about whether DevOps Guru is configured to perform log anomaly detection on Amazon CloudWatch log groups. See [`logs_anomaly_detection`](#logs_anomaly_detection-argument-reference) below.
* `ops_center` - (Required) Information about whether DevOps Guru is configured to create an OpsItem in AWS Systems Manager OpsCenter for each created insight. See [`ops_center`](#ops_center-argument-reference) below.

### `kms_server_side_encryption` Argument Reference

* `kms_key_id` - (Optional) KMS key ID. This value can be a key ID, key ARN, alias name, or alias ARN.
* `opt_in_status` - (Optional) Specifies whether KMS integration is enabled. Valid values are `DISABLED` and `ENABLED`.
* `type` - (Optional) Type of KMS key used. Valid values are `CUSTOMER_MANAGED_KEY` and `AWS_OWNED_KMS_KEY`.

### `logs_anomaly_detection` Argument Reference

* `opt_in_status` - (Optional) Specifies if DevOps Guru is configured to perform log anomaly detection on CloudWatch log groups. Valid values are `DISABLED` and `ENABLED`.

### `ops_center` Argument Reference

* `opt_in_status` - (Optional) Specifies if DevOps Guru is enabled to create an AWS Systems Manager OpsItem for each created insight. Valid values are `DISABLED` and `ENABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS region.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DevOps Guru Service Integration using the `id`. For example:

```terraform
import {
  to = aws_devopsguru_service_integration.example
  id = "us-east-1"
}
```

Using `terraform import`, import DevOps Guru Service Integration using the `id`. For example:

```console
% terraform import aws_devopsguru_service_integration.example us-east-1
```
