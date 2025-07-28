
As a convenience, for compatibility with the [Terraform S3 Backend](https://www.terraform.io/language/settings/backends/s3),
the following service endpoints can also be configured using the **deprecated** environment variables:
{{ range .Services -}}
  {{- if or .TfAwsEnvVar .DeprecatedEnvVar }}
* {{ .HumanFriendly }}: `{{ .TfAwsEnvVar }}` or `{{ .DeprecatedEnvVar }}`
  {{- end }}
{{- end }}

## Connecting to Local AWS Compatible Solutions

~> **NOTE:** This information is not intended to be exhaustive for all local AWS compatible solutions or necessarily authoritative configurations for those documented. Check the documentation for each of these solutions for the most up to date information.

### DynamoDB Local

The Amazon DynamoDB service offers a downloadable version for writing and testing applications without accessing the DynamoDB web service. For more information about this solution, see the [DynamoDB Local documentation in the Amazon DynamoDB Developer Guide](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html).

An example provider configuration:

```terraform
provider "aws" {
  access_key                  = "mock_access_key"
  region                      = "us-east-1"
  secret_key                  = "mock_secret_key"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    dynamodb = "http://localhost:8000"
  }
}
```

### LocalStack

[LocalStack](https://localstack.cloud/) provides an easy-to-use test/mocking framework for developing Cloud applications.

An example provider configuration:

```terraform
provider "aws" {
  access_key                  = "mock_access_key"
  region                      = "us-east-1"
  s3_use_path_style           = true
  secret_key                  = "mock_secret_key"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    apigateway     = "http://localhost:4566"
    cloudformation = "http://localhost:4566"
    cloudwatch     = "http://localhost:4566"
    dynamodb       = "http://localhost:4566"
    es             = "http://localhost:4566"
    firehose       = "http://localhost:4566"
    iam            = "http://localhost:4566"
    kinesis        = "http://localhost:4566"
    lambda         = "http://localhost:4566"
    route53        = "http://localhost:4566"
    redshift       = "http://localhost:4566"
    s3             = "http://localhost:4566"
    secretsmanager = "http://localhost:4566"
    ses            = "http://localhost:4566"
    sns            = "http://localhost:4566"
    sqs            = "http://localhost:4566"
    ssm            = "http://localhost:4566"
    stepfunctions  = "http://localhost:4566"
    sts            = "http://localhost:4566"
  }
}
```
