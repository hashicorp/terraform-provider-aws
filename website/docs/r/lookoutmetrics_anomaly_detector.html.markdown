---
subcategory: "Lookout for Metrics"
layout: "aws"
page_title: "AWS: aws_lookoutmetrics_anomaly_detector"
description: |-
  Terraform resource for managing an AWS Lookout for Metrics Anomaly Detector.
---

# Resource: aws_lookoutmetrics_anomaly_detector

Terraform resource for managing an AWS Lookout for Metrics Anomaly Detector.

## Example Usage

### Basic Usage

```terraform
resource "aws_lookoutmetrics_anomaly_detector" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Anomaly Detector. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

Lookout for Metrics Anomaly Detector can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_lookoutmetrics_anomaly_detector.example rft-8012925589
```
