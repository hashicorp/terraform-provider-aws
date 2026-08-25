---
subcategory: "X-Ray"
layout: "aws"
page_title: "AWS: aws_xray_trace_segment_destination"
description: |-
    Manages the destination of data sent to `PutTraceSegments` by AWS X-Ray.
---

# Resource: aws_xray_trace_segment_destination

Manages the destination of data sent to `PutTraceSegments` by AWS X-Ray.

-> **Note:** Removing this resource from Terraform has no effect on the destination configuration within AWS X-Ray.

## Example Usage

```terraform
resource "aws_xray_trace_segment_destination" "example" {
  destination = "CloudWatchLogs"
}
```

## Argument Reference

This resource supports the following arguments:

* `destination` - (Required) Destination of trace segments. Valid values: `XRay`, `CloudWatchLogs`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_xray_trace_segment_destination.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_xray_trace_segment_destination" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import XRay Trace Segment Destinations using the region name. For example:

```terraform
import {
  to = aws_xray_trace_segment_destination.example
  id = "us-west-2"
}
```

Using `terraform import`, import XRay Trace Segment Destinations using the region name. For example:

```console
% terraform import aws_xray_trace_segment_destination.example us-west-2
```
