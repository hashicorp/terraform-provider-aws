---
subcategory: "CloudWatch Observability Access Manager"
layout: "aws"
page_title: "AWS: aws_oam_sink_policy"
description: |-
  Terraform resource for managing an AWS CloudWatch Observability Access Manager Sink Policy.
---

# Resource: aws_oam_sink_policy

Terraform resource for managing an AWS CloudWatch Observability Access Manager Sink Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_oam_sink" "example" {
  name = "ExampleSink"
}

resource "aws_oam_sink_policy" "example" {
  sink_identifier = aws_oam_sink.example.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["oam:CreateLink", "oam:UpdateLink"]
        Effect   = "Allow"
        Resource = "*"
        Principal = {
          "AWS" = ["1111111111111", "222222222222"]
        }
        Condition = {
          "ForAllValues:StringEquals" = {
            "oam:ResourceTypes" = ["AWS::CloudWatch::Metric", "AWS::Logs::LogGroup"]
          }
        }
      }
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `sink_identifier` - (Required) ARN of the sink to attach this policy to.
* `policy` - (Required) JSON policy to use. If you are updating an existing policy, the entire existing policy is replaced by what you specify here.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Sink.
* `sink_id` - ID string that AWS generated as part of the sink ARN.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `1m`)
* `update` - (Default `1m`)

## Import

CloudWatch Observability Access Manager Sink Policy can be imported using the `sink_identifier`, e.g.,

```
$ terraform import aws_oam_sink_policy.example arn:aws:oam:us-west-2:123456789012:sink/sink-id
```
