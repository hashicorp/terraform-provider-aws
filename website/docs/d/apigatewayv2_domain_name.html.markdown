---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_domain_name"
description: |-
  Terraform data source for managing an AWS API Gateway V2 Domain Name.
---

# Data Source: aws_apigatewayv2_domain_name

Terraform data source for managing an AWS API Gateway V2 Domain Name. 

More information can be found in [Terraform Apigateway V2 Domain Name](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/apigatewayv2_domain_name) and [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html).


## Example Usage

### Basic Usage

```terraform
resource "aws_apigatewayv2_domain_name" "example" {
  domain_name              = "example.terraform.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.example.arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}

data "aws_apigatewayv2_domain_name" "example" {
  domain_name = aws_apigatewayv2_domain_name.example.domain_name
}
```

## Argument Reference

The following arguments are required:

* `domain_name` - (Required) Domain name of resource.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `api_mapping_selection_expression` - This expression is evaluated to determine which API stage is selected when a request is made using a custom domain. [AWS Docs](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-mapping-selection-expressions)
* `domain_name` - The name of the domain_name resource.
* `domain_name_configurations` - The list of domain name configurations.
* `mutual_tls_authentication` - The mutual TLS authentication configuration for a custom domain name.
* `tags` - The collection of tags associated with a domain name.

### `domain_name_configuration`

* `certificate_arn` - An AWS-managed certificate that will be used by the edge-optimized endpoint for this domain name. AWS Certificate Manager is the only supported source.
* `certificate_name` - The user-friendly name of the certificate that will be used by the edge-optimized endpoint for this domain name.
* `certificate_upload_date` - The timestamp when the certificate that was used by edge-optimized endpoint for this domain name was uploaded.
* `domain_name_status` - The status of the domain name migration. The valid values are AVAILABLE, UPDATING, PENDING_CERTIFICATE_REIMPORT, and PENDING_OWNERSHIP_VERIFICATION. If the status is UPDATING, the domain cannot be modified further until the existing operation is complete. If it is AVAILABLE, the domain can be updated.
* `domain_name_status_message` - An optional text message containing detailed information about status of the domain name migration.
* `endpoint_type` - The endpoint type.
* `hosted_zone_id` - The Amazon Route 53 Hosted Zone ID of the endpoint.
* `ownership_verification_certificate_arn` - The ARN of the public certificate issued by ACM to validate ownership of your custom domain. Only required when configuring mutual TLS and using an ACM imported or private CA certificate ARN as the regionalCertificateArn
* `security_policy` - The Transport Layer Security (TLS) version of the security policy for this domain name.

### `mutual_tls_authentication`

* `truststore_uri` - An Amazon S3 URL that specifies the truststore for mutual TLS authentication, for example, s3://bucket-name/key-name. The truststore can contain certificates from public or private certificate authorities. To update the truststore, upload a new version to S3, and then update your custom domain name to use the new version. To update the truststore, you must have permissions to access the S3 object.
* `truststore_version` - The version of the S3 object that contains your truststore. To specify a version, you must have versioning enabled for the S3 bucket.
* `truststore_warnings` - A list of warnings that API Gateway returns while processing your truststore. Invalid certificates produce warnings. Mutual TLS is still enabled, but some clients might not be able to access your API. To resolve warnings, upload a new truststore to S3, and then update you domain name to use the new version.