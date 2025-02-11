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
* `dynamodb_config` - (Optional) DynamoDB settings. See [`dynamodb_config` Block](#dynamodb_config-block) for details.
* `elasticsearch_config` - (Optional) Amazon Elasticsearch settings. See [`elasticsearch_config` Block](#elasticsearch_config-block) for details.
* `event_bridge_config` - (Optional) AWS EventBridge settings. See [`event_bridge_config` Block](#event_bridge_config-block) for details.
* `http_config` - (Optional) HTTP settings. See [`http_config` Block](#http_config-block) for details.
* `lambda_config` - (Optional) AWS Lambda settings. See [`lambda_config` Block](#lambda_config-block) for details.
* `opensearchservice_config` - (Optional) Amazon OpenSearch Service settings. See [`opensearchservice_config` Block](#opensearchservice_config-block) for details.
* `relational_database_config` (Optional) AWS RDS settings. See [`relational_database_config` Block](#relational_database_config-block) for details.
* `service_role_arn` - (Optional) IAM service role ARN for the data source. Required if `type` is specified as `AWS_LAMBDA`, `AMAZON_DYNAMODB`, `AMAZON_ELASTICSEARCH`, `AMAZON_EVENTBRIDGE`, or `AMAZON_OPENSEARCH_SERVICE`.

### `dynamodb_config` Block

The `dynamodb_config` configuration block supports the following arguments:

* `table_name` - (Required) Name of the DynamoDB table.
* `region` - (Optional) AWS region of the DynamoDB table. Defaults to current region.
* `use_caller_credentials` - (Optional) Set to `true` to use Amazon Cognito credentials with this data source.
* `delta_sync_config` - (Optional) The DeltaSyncConfig for a versioned data source. See [`delta_sync_config` Block](#delta_sync_config-block) for details.
* `versioned` - (Optional) Detects Conflict Detection and Resolution with this data source.

### `delta_sync_config` Block

The `delta_sync_config` configuration block supports the following arguments:

* `base_table_ttl` - (Optional) The number of minutes that an Item is stored in the data source.
* `delta_sync_table_name` - (Required) The table name.
* `delta_sync_table_ttl` - (Optional) The number of minutes that a Delta Sync log entry is stored in the Delta Sync table.

### `elasticsearch_config` Block

The `elasticsearch_config` configuration block supports the following arguments:

* `endpoint` - (Required) HTTP endpoint of the Elasticsearch domain.
* `region` - (Optional) AWS region of Elasticsearch domain. Defaults to current region.

### `event_bridge_config` Block

The `event_bridge_config` configuration block supports the following arguments:

* `event_bus_arn` - (Required) ARN for the EventBridge bus.

### `http_config` Block

The `http_config` configuration block supports the following arguments:

* `endpoint` - (Required) HTTP URL.
* `authorization_config` - (Optional) Authorization configuration in case the HTTP endpoint requires authorization. See [`authorization_config` Block](#authorization_config-block) for details.

### `authorization_config` Block

The `authorization_config` configuration block supports the following arguments:

* `authorization_type` - (Optional) Authorization type that the HTTP endpoint requires. Default values is `AWS_IAM`.
* `aws_iam_config` - (Optional) Identity and Access Management (IAM) settings. See [`aws_iam_config` Block](#aws_iam_config-block) for details.

### `aws_iam_config` Block

The `aws_iam_config` configuration block supports the following arguments:

* `signing_region` - (Optional) Signing Amazon Web Services Region for IAM authorization.
* `signing_service_name`- (Optional) Signing service name for IAM authorization.

### `lambda_config` Block

The `lambda_config` configuration block supports the following arguments:

* `function_arn` - (Required) ARN for the Lambda function.

### `opensearchservice_config` Block

The `opensearchservice_config` configuration block supports the following arguments:

* `endpoint` - (Required) HTTP endpoint of the OpenSearch domain.
* `region` - (Optional) AWS region of the OpenSearch domain. Defaults to current region.

### `relational_database_config` Block

The `relational_database_config` configuration block supports the following arguments:

* `http_endpoint_config` - (Required) Amazon RDS HTTP endpoint configuration. See [`http_endpoint_config` Block](#http_endpoint_config-block) for details.
* `source_type` - (Optional) Source type for the relational database. Valid values: `RDS_HTTP_ENDPOINT`.

### `http_endpoint_config` Block

The `http_endpoint_config` configuration block supports the following arguments:

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
