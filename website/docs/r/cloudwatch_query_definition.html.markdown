---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_query_definition"
description: |-
  Provides a CloudWatch Logs query definition resource.
---

# Resource: aws_cloudwatch_query_definition

Provides a CloudWatch Logs query definition resource.

## Example Usage

```terraform
resource "aws_cloudwatch_query_definition" "example" {
  name = "custom_query"

  log_group_names = [
    "/aws/logGroup1",
    "/aws/logGroup2"
  ]

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 25
EOF
}
```

### Query Parameters

```terraform
resource "aws_cloudwatch_query_definition" "example" {
  name           = "errors_by_status"
  query_language = "CWLI"

  log_group_names = [
    "/aws/logGroup1",
  ]

  query_string = <<EOF
fields @timestamp, @message
| filter status = {{status}}
| sort @timestamp desc
| limit 50
EOF

  parameter {
    name          = "status"
    default_value = "500"
    description   = "HTTP status code to filter on"
  }
}
```

### OpenSearch SQL

```terraform
resource "aws_cloudwatch_query_definition" "example" {
  name           = "recent_messages_sql"
  query_language = "SQL"

  query_string = "SELECT `@timestamp`, `@message` FROM `/aws/logGroup1` LIMIT 25"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the query.
* `query_string` - (Required) The query to save. You can read more about CloudWatch Logs Query Syntax in the [documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CWL_QuerySyntax.html).

The following arguments are optional:

* `log_group_names` - (Optional) Specific log groups to use with the query. For OpenSearch SQL queries, specify the log groups in `query_string` instead.
* `parameter` - (Optional) Query parameters that define placeholder variables for the query string. Reference a parameter in `query_string` as `{{name}}`. Query parameters are supported only for Logs Insights QL (`CWLI`) queries. See [`parameter` Block](#parameter-block) below.
* `query_language` - (Optional) Query language of the query. Valid values are `CWLI` (Logs Insights QL), `PPL` (OpenSearch PPL) and `SQL` (OpenSearch SQL). If not set, the language reported by AWS is used. See [Supported query languages](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CWL_AnalyzeLogData_Languages.html).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `parameter` Block

The `parameter` block supports the following arguments:

* `default_value` - (Optional) Value to use for the parameter if no value is supplied when the query runs. An empty string means no default value.
* `description` - (Optional) Description of the parameter.
* `name` - (Required) Name of the parameter. Must start with a letter or underscore and contain only letters, digits, and underscores.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The query definition ARN.
* `query_definition_id` - The query definition ID.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_query_definition.example
  identity = {
    query_definition_id = "269951d7-6f75-496d-9d7b-6b7a5486bdbd"
  }
}

resource "aws_cloudwatch_query_definition" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `query_definition_id` (String) ID of the query definition.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Query Definitions using `arn`. The ARN can be found on the "Edit Query" page for the query in the AWS Console. For example:

```terraform
import {
  to = aws_cloudwatch_query_definition.example
  id = "arn:aws:logs:us-west-2:123456789012:query-definition:269951d7-6f75-496d-9d7b-6b7a5486bdbd"
}
```

Using `terraform import`, import Query Definitions using `arn`. The ARN can be found on the "Edit Query" page for the query in the AWS Console. For example:

```console
% terraform import aws_cloudwatch_query_definition.example arn:aws:logs:us-west-2:123456789012:query-definition:269951d7-6f75-496d-9d7b-6b7a5486bdbd
```
