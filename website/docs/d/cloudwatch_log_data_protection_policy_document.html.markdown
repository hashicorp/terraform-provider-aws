---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_data_protection_policy_document"
description: |-
  Generates a CloudWatch Log Group Data Protection Policy document in JSON format
---

# Data Source: aws_cloudwatch_log_data_protection_policy_document

Generates a CloudWatch Log Group Data Protection Policy document in JSON format for use with the `aws_cloudwatch_log_data_protection_policy` resource.

-> For more information about data protection policies, see the [Help protect sensitive log data with masking](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/mask-sensitive-log-data.html).

## Example Usage

```terraform
resource "aws_cloudwatch_log_data_protection_policy" "example" {
  log_group_name  = aws_cloudwatch_log_group.example.name
  policy_document = data.aws_cloudwatch_log_data_protection_policy_document.example.json
}

data "aws_cloudwatch_log_data_protection_policy_document" "example" {
  name = "Example"

  statement {
    sid = "Audit"

    data_identifiers = [
      "arn:aws:dataprotection::aws:data-identifier/EmailAddress",
      "arn:aws:dataprotection::aws:data-identifier/DriversLicense-US",
    ]

    operation {
      audit {
        findings_destination {
          cloudwatch_logs {
            log_group = aws_cloudwatch_log_group.audit.name
          }
          firehose {
            delivery_stream = aws_kinesis_firehose_delivery_stream.audit.name
          }
          s3 {
            bucket = aws_s3_bucket.audit.bucket
          }
        }
      }
    }
  }

  statement {
    sid = "Deidentify"

    data_identifiers = [
      "arn:aws:dataprotection::aws:data-identifier/EmailAddress",
      "arn:aws:dataprotection::aws:data-identifier/DriversLicense-US",
    ]

    operation {
      deidentify {
        mask_config {}
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the data protection policy document.
* `statement` - (Required) Configures the data protection policy.

-> There must be exactly two statements: the first with an `audit` operation, and the second with a `deidentify` operation.

The following arguments are optional:

* `description` - (Optional)
* `version` - (Optional)

### statement Configuration Block

* `data_identifiers` - (Required) Set of at least 1 sensitive data identifiers that you want to mask. Read more in [Types of data that you can protect](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/protect-sensitive-log-data-types.html).
* `operation` - (Required) Configures the data protection operation applied by this statement.
* `sid` - (Optional) Name of this statement.

#### operation Configuration Block

* `audit` - (Optional) Configures the detection of sensitive data.
* `deidentify` - (Optional) Configures the masking of sensitive data.

-> Every policy statement must specify exactly one operation.

##### audit Configuration Block

* `findings_destination` - (Required) Configures destinations to send audit findings to.

##### findings_destination Configuration Block

* `cloudwatch_logs` - (Optional) Configures CloudWatch Logs as a findings destination.
* `firehose` - (Optional) Configures Kinesis Firehose as a findings destination.
* `s3` - (Optional) Configures S3 as a findings destination.

###### cloudwatch_logs Configuration Block

* `log_group` - (Required) Name of the CloudWatch Log Group to send findings to.

###### firehose Configuration Block

* `delivery_stream` - (Required) Name of the Kinesis Firehose Delivery Stream to send findings to.

###### s3 Configuration Block

* `bucket` - (Required) Name of the S3 Bucket to send findings to.

##### deidentify Configuration Block

* `mask_config` - (Required) An empty object that configures masking.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `json` - Standard JSON policy document rendered based on the arguments above.
