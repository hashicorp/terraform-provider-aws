---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_query_definition"
description: |-
  Provides a CloudWatch Logs query definition resource.
---

# Resource: aws_cloudwatch_query_definition

Provides a CloudWatch Logs query definition resource.

## Example Usage

```hcl
resource "aws_cloudwatch_query_definition" "query" {
  name = "custom_query"
  log_groups = [
    "/aws/logGroup1",
    "/aws/logGroup2"
  ]
  query = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 25
EOF
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the query.
* `query` - (Required) The query to save. You can read more about CloudWatch Logs Query Syntax in the [documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CWL_QuerySyntax.html).
* `log_groups` - (Optional) The names of the log groups to save with the query.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `query_definition_id` - The query definition ID of the saved query.

## Import

CloudWatch query definitions can be imported using the query name and query definition ID,
separated by an underscore (`_`).

```
$ terraform import aws_cloudwatch_query_definition.query custom_query_269951d7-6f75-496d-9d7b-6b7a5486bdbd
```
