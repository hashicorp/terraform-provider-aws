---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_domain_name"
description: |-
  Manages an Amazon API Gateway Version 2 domain name.
---

# Resource: aws_apigatewayv2_domain_name

Manages an Amazon API Gateway Version 2 domain name.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html).

-> **Note:** This resource establishes ownership of and the TLS settings for
a particular domain name. An API stage can be associated with the domain name using the `aws_apigatewayv2_api_mapping` resource.

## Example Usage

### Basic

```terraform
resource "aws_apigatewayv2_domain_name" "example" {
  domain_name = "ws-api.example.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.example.arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
```

### Associated Route 53 Resource Record

```terraform
resource "aws_apigatewayv2_domain_name" "example" {
  domain_name = "http-api.example.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.example.arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}

resource "aws_route53_record" "example" {
  name    = aws_apigatewayv2_domain_name.example.domain_name
  type    = "A"
  zone_id = aws_route53_zone.example.zone_id

  alias {
    name                   = aws_apigatewayv2_domain_name.example.domain_name_configuration[0].target_domain_name
    zone_id                = aws_apigatewayv2_domain_name.example.domain_name_configuration[0].hosted_zone_id
    evaluate_target_health = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) Domain name. Must be between 1 and 512 characters in length.
* `domain_name_configuration` - (Required) Domain name configuration. See below.
* `mutual_tls_authentication` - (Optional) Mutual TLS authentication configuration for the domain name.
* `tags` - (Optional) Map of tags to assign to the domain name. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `domain_name_configuration`

* `certificate_arn` - (Required) ARN of an AWS-managed certificate that will be used by the endpoint for the domain name. AWS Certificate Manager is the only supported source. Use the [`aws_acm_certificate`](/docs/providers/aws/r/acm_certificate.html) resource to configure an ACM certificate.
* `endpoint_type` - (Required) Endpoint type. Valid values: `REGIONAL`.
* `hosted_zone_id` - (Computed) Amazon Route 53 Hosted Zone ID of the endpoint.
* `ownership_verification_certificate_arn` - (Optional) ARN of the AWS-issued certificate used to validate custom domain ownership (when `certificate_arn` is issued via an ACM Private CA or `mutual_tls_authentication` is configured with an ACM-imported certificate.)
* `security_policy` - (Required) Transport Layer Security (TLS) version of the [security policy](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-custom-domain-tls-version.html) for the domain name. Valid values: `TLS_1_2`.
* `target_domain_name` - (Computed) Target domain name.

### `mutual_tls_authentication`

* `truststore_uri` - (Required) Amazon S3 URL that specifies the truststore for mutual TLS authentication, for example, `s3://bucket-name/key-name`. The truststore can contain certificates from public or private certificate authorities. To update the truststore, upload a new version to S3, and then update your custom domain name to use the new version.
* `truststore_version` - (Optional) Version of the S3 object that contains the truststore. To specify a version, you must have versioning enabled for the S3 bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `api_mapping_selection_expression` - [API mapping selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-mapping-selection-expressions) for the domain name.
* `arn` - ARN of the domain name.
* `id` - Domain name identifier.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `60m`)

## Import

`aws_apigatewayv2_domain_name` can be imported by using the domain name, e.g.,

```
$ terraform import aws_apigatewayv2_domain_name.example ws-api.example.com
```
