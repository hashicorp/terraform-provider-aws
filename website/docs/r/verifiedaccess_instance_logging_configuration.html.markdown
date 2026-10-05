---
subcategory: "Verified Access"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_instance_logging_configuration"
description: |-
  Terraform resource for managing a Verified Access Instance Logging Configuration.
---

# Resource: aws_verifiedaccess_instance_logging_configuration

Terraform resource for managing a Verified Access Logging Configuration.

~> Deletion of this resource resets the Verified Access Instance Logging Configuration to its default values.

## Example Usage

### With CloudWatch Logging

```terraform
resource "aws_verifiedaccess_instance_logging_configuration" "example" {
  access_logs {
    cloudwatch_logs {
      enabled   = true
      log_group = aws_cloudwatch_log_group.example.id
    }
  }
  verifiedaccess_instance_id = aws_verifiedaccess_instance.example.id
}
```

### With Kinesis Data Firehose Logging

```terraform
resource "aws_verifiedaccess_instance_logging_configuration" "example" {
  access_logs {
    kinesis_data_firehose {
      delivery_stream = aws_kinesis_firehose_delivery_stream.example.name
      enabled         = true
    }
  }
  verifiedaccess_instance_id = aws_verifiedaccess_instance.example.id
}
```

### With S3 logging

```terraform
resource "aws_verifiedaccess_instance_logging_configuration" "example" {
  access_logs {
    s3 {
      bucket_name = aws_s3_bucket.example.id
      enabled     = true
      prefix      = "example"
    }
  }
  verifiedaccess_instance_id = aws_verifiedaccess_instance.example.id
}
```

### With all three logging options

```terraform
resource "aws_verifiedaccess_instance_logging_configuration" "example" {
  access_logs {
    cloudwatch_logs {
      enabled   = true
      log_group = aws_cloudwatch_log_group.example.id
    }
    kinesis_data_firehose {
      delivery_stream = aws_kinesis_firehose_delivery_stream.example.name
      enabled         = true
    }
    s3 {
      bucket_name = aws_s3_bucket.example.id
      enabled     = true
    }
  }
  verifiedaccess_instance_id = aws_verifiedaccess_instance.example.id
}
```

### With `include_trust_context`

```terraform
resource "aws_verifiedaccess_instance_logging_configuration" "example" {
  access_logs {
    include_trust_context = true
  }
  verifiedaccess_instance_id = aws_verifiedaccess_instance.example.id
}
```

### With `log_version`

```terraform
resource "aws_verifiedaccess_instance_logging_configuration" "example" {
  access_logs {
    log_version = "ocsf-1.0.0-rc.2"
  }
  verifiedaccess_instance_id = aws_verifiedaccess_instance.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `access_logs` - (Required) Block that specifies the configuration options for Verified Access instances. [Detailed below](#access_logs-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `verifiedaccess_instance_id` - (Required, Forces new resource) ID of the Verified Access instance.

### `access_logs` Block

A `access_logs` block supports the following arguments:

* `cloudwatch_logs` - (Optional) Block that specifies configures sending Verified Access logs to CloudWatch Logs. [Detailed below](#cloudwatch_logs-block).
* `include_trust_context` - (Optional) Include trust data sent by trust providers into the logs.
* `kinesis_data_firehose` - (Optional) Block that specifies configures sending Verified Access logs to Kinesis. [Detailed below](#kinesis_data_firehose-block).
* `log_version` - (Optional) Logging version to use. Refer to [VerifiedAccessLogOptions](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_VerifiedAccessLogOptions.html) for the allowed values.
* `s3` - (Optional) Block that specifies configures sending Verified Access logs to S3. [Detailed below](#s3-block).

#### `cloudwatch_logs` Block

A `cloudwatch_logs` block supports the following arguments:

* `enabled` - (Required) Whether logging is enabled.
* `log_group` - (Optional) Name of the CloudWatch Logs Log Group.

#### `kinesis_data_firehose` Block

A `kinesis_data_firehose` block supports the following arguments:

* `delivery_stream` - (Optional) Name of the delivery stream.
* `enabled` - (Required) Whether logging is enabled.

#### `s3` Block

A `s3` block supports the following arguments:

* `bucket_name` - (Optional) Name of S3 bucket.
* `bucket_owner` - (Optional) ID of the AWS account that owns the Amazon S3 bucket.
* `enabled` - (Required) Whether logging is enabled.
* `prefix` - (Optional) Bucket prefix.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Verified Access Logging Configuration using the Verified Access Instance `id`. For example:

```terraform
import {
  to = aws_verifiedaccess_instance_logging_configuration.example
  id = "vai-1234567890abcdef0"
}
```

Using `terraform import`, import Verified Access Logging Configuration using the Verified Access Instance `id`. For example:

```console
% terraform import aws_verifiedaccess_instance_logging_configuration.example vai-1234567890abcdef0
```
