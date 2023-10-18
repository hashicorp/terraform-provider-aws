---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_datasource"
description: |-
  Provides an AppSync Data Source.
---

# Resource: aws_appsync_datasource

Provides an AppSync Data Source.

## Example Usage

```terraform
resource "aws_dynamodb_table" "example" {
  name           = "example"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserId"

  attribute {
    name = "UserId"
    type = "S"
  }
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["appsync.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}
resource "aws_iam_role" "example" {
  name               = "example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "example" {
  statement {
    effect    = "Allow"
    actions   = ["dynamodb:*"]
    resources = [aws_dynamodb_table.example.arn]
  }
}

resource "aws_iam_role_policy" "example" {
  name   = "example"
  role   = aws_iam_role.example.id
  policy = data.aws_iam_policy_document.example.json
}

resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "tf_appsync_example"
}

resource "aws_appsync_datasource" "example" {
  api_id           = aws_appsync_graphql_api.example.id
  name             = "tf_appsync_example"
  service_role_arn = aws_iam_role.example.arn
  type             = "AMAZON_DYNAMODB"

  dynamodb_config {
    table_name = aws_dynamodb_table.example.name
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `api_id` - (Required) API ID for the GraphQL API for the data source.
* `name` - (Required) User-supplied name for the data source.
* `type` - (Required) Type of the Data Source. Valid values: `AWS_LAMBDA`, `AMAZON_DYNAMODB`, `AMAZON_ELASTICSEARCH`, `HTTP`, `NONE`, `RELATIONAL_DATABASE`, `AMAZON_EVENTBRIDGE`, `AMAZON_OPENSEARCH_SERVICE`.
* `description` - (Optional) Description of the data source.
* `dynamodb_config` - (Optional) DynamoDB settings. See [DynamoDB Config](#dynamodb-config)
* `elasticsearch_config` - (Optional) Amazon Elasticsearch settings. See [ElasticSearch Config](#elasticsearch-config)
* `event_bridge_config` - (Optional) AWS EventBridge settings. See [Event Bridge Config](#event-bridge-config)
* `http_config` - (Optional) HTTP settings. See [HTTP Config](#http-config)
* `lambda_config` - (Optional) AWS Lambda settings. See [Lambda Config](#lambda-config)
* `opensearchservice_config` - (Optional) Amazon OpenSearch Service settings. See [OpenSearch Service Config](#opensearch-service-config)
* `relational_database_config` (Optional) AWS RDS settings. See [Relational Database Config](#relational-database-config)
* `service_role_arn` - (Optional) IAM service role ARN for the data source.

### DynamoDB Config

This argument supports the following arguments:

* `table_name` - (Required) Name of the DynamoDB table.
* `region` - (Optional) AWS region of the DynamoDB table. Defaults to current region.
* `use_caller_credentials` - (Optional) Set to `true` to use Amazon Cognito credentials with this data source.
* `delta_sync_config` - (Optional) The DeltaSyncConfig for a versioned data source. See [Delta Sync Config](#delta-sync-config)
* `versioned` - (Optional) Detects Conflict Detection and Resolution with this data source.

### Delta Sync Config

* `base_table_ttl` - (Optional) The number of minutes that an Item is stored in the data source.
* `delta_sync_table_name` - (Required) The table name.
* `delta_sync_table_ttl` - (Optional) The number of minutes that a Delta Sync log entry is stored in the Delta Sync table.

### ElasticSearch Config

This argument supports the following arguments:

* `endpoint` - (Required) HTTP endpoint of the Elasticsearch domain.
* `region` - (Optional) AWS region of Elasticsearch domain. Defaults to current region.

### Event Bridge Config

This argument supports the following arguments:

* `event_bus_arn` - (Required) ARN for the EventBridge bus.

### HTTP Config

This argument supports the following arguments:

* `endpoint` - (Required) HTTP URL.
* `authorization_config` - (Optional) Authorization configuration in case the HTTP endpoint requires authorization. See [Authorization Config](#authorization-config).

#### Authorization Config

This argument supports the following arguments:

* `authorization_type` - (Optional) Authorization type that the HTTP endpoint requires. Default values is `AWS_IAM`.
* `aws_iam_config` - (Optional) Identity and Access Management (IAM) settings. See [AWS IAM Config](#aws-iam-config).

##### AWS IAM Config

This argument supports the following arguments:

* `signing_region` - (Optional) Signing Amazon Web Services Region for IAM authorization.
* `signing_service_name`- (Optional) Signing service name for IAM authorization.

### Lambda Config

This argument supports the following arguments:

* `function_arn` - (Required) ARN for the Lambda function.

### OpenSearch Service Config

This argument supports the following arguments:

* `endpoint` - (Required) HTTP endpoint of the OpenSearch domain.
* `region` - (Optional) AWS region of the OpenSearch domain. Defaults to current region.

### Relational Database Config

This argument supports the following arguments:

* `http_endpoint_config` - (Required) Amazon RDS HTTP endpoint configuration. See [HTTP Endpoint Config](#http-endpoint-config).
* `source_type` - (Optional) Source type for the relational database. Valid values: `RDS_HTTP_ENDPOINT`.

#### HTTP Endpoint Config

This argument supports the following arguments:

* `db_cluster_identifier` - (Required) Amazon RDS cluster identifier.
* `aws_secret_store_arn` - (Required) AWS secret store ARN for database credentials.
* `database_name` - (Optional) Logical database name.
* `region` - (Optional) AWS Region for RDS HTTP endpoint. Defaults to current region.
* `schema` - (Optional) Logical schema name.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appsync_datasource` using the `api_id`, a hyphen, and `name`. For example:

```terraform
import {
  to = aws_appsync_datasource.example
  id = "abcdef123456-example"
}
```

Using `terraform import`, import `aws_appsync_datasource` using the `api_id`, a hyphen, and `name`. For example:

```console
% terraform import aws_appsync_datasource.example abcdef123456-example
```
