---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_domain_name"
description: |-
  Registers a custom domain name for use with AWS API Gateway.
---

# Resource: aws_api_gateway_domain_name

Registers a custom domain name for use with AWS API Gateway. Additional information about this functionality
can be found in the [API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html).

This resource just establishes ownership of and the TLS settings for
a particular domain name. An API can be attached to a particular path
under the registered domain name using
[the `aws_api_gateway_base_path_mapping` resource](api_gateway_base_path_mapping.html).

API Gateway domains can be defined as either 'edge-optimized' or 'regional'.  In an edge-optimized configuration,
API Gateway internally creates and manages a CloudFront distribution to route requests on the given hostname. In
addition to this resource it's necessary to create a DNS record corresponding to the given domain name which is an alias
(either Route53 alias or traditional CNAME) to the Cloudfront domain name exported in the `cloudfront_domain_name`
attribute.

In a regional configuration, API Gateway does not create a CloudFront distribution to route requests to the API, though
a distribution can be created if needed. In either case, it is necessary to create a DNS record corresponding to the
given domain name which is an alias (either Route53 alias or traditional CNAME) to the regional domain name exported in
the `regional_domain_name` attribute.

~> **Note:** API Gateway requires the use of AWS Certificate Manager (ACM) certificates instead of Identity and Access Management (IAM) certificates in regions that support ACM. Regions that support ACM can be found in the [Regions and Endpoints Documentation](https://docs.aws.amazon.com/general/latest/gr/rande.html#acm_region). To import an existing private key and certificate into ACM or request an ACM certificate, see the [`aws_acm_certificate` resource](/docs/providers/aws/r/acm_certificate.html).

~> **Note:** The `aws_api_gateway_domain_name` resource expects dependency on the `aws_acm_certificate_validation` as
only verified certificates can be used. This can be made either explicitly by adding the
`depends_on = [aws_acm_certificate_validation.cert]` attribute. Or implicitly by referring certificate ARN
from the validation resource where it will be available after the resource creation:
`regional_certificate_arn = aws_acm_certificate_validation.cert.certificate_arn`.

~> **Note:** All arguments including the private key will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

An end-to-end example of a REST API configured with OpenAPI can be found in the [`/examples/api-gateway-rest-api-openapi` directory within the GitHub repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/api-gateway-rest-api-openapi).

### Edge Optimized (ACM Certificate)

```terraform
resource "aws_api_gateway_domain_name" "example" {
  certificate_arn = aws_acm_certificate_validation.example.certificate_arn
  domain_name     = "api.example.com"
}

# Example DNS record using Route53.
# Route53 is not specifically required; any DNS host can be used.
resource "aws_route53_record" "example" {
  name    = aws_api_gateway_domain_name.example.domain_name
  type    = "A"
  zone_id = aws_route53_zone.example.id

  alias {
    evaluate_target_health = true
    name                   = aws_api_gateway_domain_name.example.cloudfront_domain_name
    zone_id                = aws_api_gateway_domain_name.example.cloudfront_zone_id
  }
}
```

### Edge Optimized (IAM Certificate)

```terraform
resource "aws_api_gateway_domain_name" "example" {
  domain_name = "api.example.com"

  certificate_name        = "example-api"
  certificate_body        = file("${path.module}/example.com/example.crt")
  certificate_chain       = file("${path.module}/example.com/ca.crt")
  certificate_private_key = file("${path.module}/example.com/example.key")
}

# Example DNS record using Route53.
# Route53 is not specifically required; any DNS host can be used.
resource "aws_route53_record" "example" {
  zone_id = aws_route53_zone.example.id # See aws_route53_zone for how to create this

  name = aws_api_gateway_domain_name.example.domain_name
  type = "A"

  alias {
    name                   = aws_api_gateway_domain_name.example.cloudfront_domain_name
    zone_id                = aws_api_gateway_domain_name.example.cloudfront_zone_id
    evaluate_target_health = true
  }
}
```

### Regional (ACM Certificate)

```terraform
resource "aws_api_gateway_domain_name" "example" {
  domain_name              = "api.example.com"
  regional_certificate_arn = aws_acm_certificate_validation.example.certificate_arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

# Example DNS record using Route53.
# Route53 is not specifically required; any DNS host can be used.
resource "aws_route53_record" "example" {
  name    = aws_api_gateway_domain_name.example.domain_name
  type    = "A"
  zone_id = aws_route53_zone.example.id

  alias {
    evaluate_target_health = true
    name                   = aws_api_gateway_domain_name.example.regional_domain_name
    zone_id                = aws_api_gateway_domain_name.example.regional_zone_id
  }
}
```

### Regional (IAM Certificate)

```terraform
resource "aws_api_gateway_domain_name" "example" {
  certificate_body          = file("${path.module}/example.com/example.crt")
  certificate_chain         = file("${path.module}/example.com/ca.crt")
  certificate_private_key   = file("${path.module}/example.com/example.key")
  domain_name               = "api.example.com"
  regional_certificate_name = "example-api"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

# Example DNS record using Route53.
# Route53 is not specifically required; any DNS host can be used.
resource "aws_route53_record" "example" {
  name    = aws_api_gateway_domain_name.example.domain_name
  type    = "A"
  zone_id = aws_route53_zone.example.id

  alias {
    evaluate_target_health = true
    name                   = aws_api_gateway_domain_name.example.regional_domain_name
    zone_id                = aws_api_gateway_domain_name.example.regional_zone_id
  }
}
```

### Enhanced Security Policy

```terraform
resource "aws_api_gateway_domain_name" "example" {
  domain_name              = "api.example.com"
  regional_certificate_arn = aws_acm_certificate_validation.example.certificate_arn
  security_policy          = "SecurityPolicy_TLS13_1_3_2025_09"
  endpoint_access_mode     = "STRICT"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `domain_name` - (Required) Fully-qualified domain name to register.
* `endpoint_access_mode` - (Optional) Endpoint access mode of the DomainName. Only available for domain names that use security policies that start with `SecurityPolicy_`. Valid values: `BASIC`, `STRICT`.
* `endpoint_configuration` - (Optional) Configuration block defining API endpoint information including type. See below.
* `mutual_tls_authentication` - (Optional) Mutual TLS authentication configuration for the domain name. See below.
* `ownership_verification_certificate_arn` - (Optional) ARN of the AWS-issued certificate used to validate custom domain ownership (when `certificate_arn` is issued via an ACM Private CA or `mutual_tls_authentication` is configured with an ACM-imported certificate.)
* `policy` - (Optional) A stringified JSON policy document that applies to the execute-api service for this DomainName regardless of the caller and Method configuration. Supported only for private custom domain names.
* `routing_mode` - (Optional) Mode to route traffic for the domain name. Valid values: `BASE_PATH_MAPPING_ONLY`, `ROUTING_RULE_ONLY`, `ROUTING_RULE_THEN_BASE_PATH_MAPPING`.
* `security_policy` - (Optional) Transport Layer Security (TLS) version + cipher suite for this DomainName. Must be configured to perform drift detection. For a list of valid security policies, see [DomainName](https://docs.aws.amazon.com/apigateway/latest/api/API_DomainName.html) in the Amazon API Gateway API Reference.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

When referencing an AWS-managed certificate, the following arguments are supported:

* `certificate_arn` - (Optional) ARN for an AWS-managed certificate. AWS Certificate Manager is the only supported source. Used when an edge-optimized domain name is desired. Conflicts with `certificate_name`, `certificate_body`, `certificate_chain`, `certificate_private_key`, `regional_certificate_arn`, and `regional_certificate_name`.
* `regional_certificate_arn` - (Optional) ARN for an AWS-managed certificate. AWS Certificate Manager is the only supported source. Used when a regional domain name is desired. Conflicts with `certificate_arn`, `certificate_name`, `certificate_body`, `certificate_chain`, and `certificate_private_key`.

When uploading a certificate, the following arguments are supported:

* `certificate_body` - (Optional) Certificate issued for the domain name being registered, in PEM format. Only valid for `EDGE` endpoint configuration type. Conflicts with `certificate_arn`, `regional_certificate_arn`, and `regional_certificate_name`.
* `certificate_chain` - (Optional) Certificate for the CA that issued the certificate, along with any intermediate CA certificates required to create an unbroken chain to a certificate trusted by the intended API clients. Only valid for `EDGE` endpoint configuration type. Conflicts with `certificate_arn`, `regional_certificate_arn`, and `regional_certificate_name`.
* `certificate_name` - (Optional) Unique name to use when registering this certificate as an IAM server certificate. Conflicts with `certificate_arn`, `regional_certificate_arn`, and `regional_certificate_name`. Required if `certificate_arn` is not set.
* `certificate_private_key` - (Optional) Private key associated with the domain certificate given in `certificate_body`. Only valid for `EDGE` endpoint configuration type. Conflicts with `certificate_arn`, `regional_certificate_arn`, and `regional_certificate_name`.
* `regional_certificate_name` - (Optional) User-friendly name of the certificate that will be used by regional endpoint for this domain name. Conflicts with `certificate_arn`, `certificate_name`, `certificate_body`, `certificate_chain`, and `certificate_private_key`.

### endpoint_configuration

* `ip_address_type` - (Optional) The IP address types that can invoke a DomainName. Valid values: `ipv4`, `dualstack`. Use `ipv4` to allow only IPv4 addresses to invoke a DomainName, or use `dualstack` to allow both IPv4 and IPv6 addresses to invoke a DomainName. For the `PRIVATE` endpoint type, only `dualstack` is supported. Terraform performs drift detection for this argument only when the value is provided.
* `types` - (Required) A list of endpoint types of an API or its custom domain name. For an edge-optimized API and its custom domain name, the endpoint type is `EDGE`. For a regional API and its custom domain name, the endpoint type is `REGIONAL`. For a private API, the endpoint type is `PRIVATE`.

### mutual_tls_authentication

* `truststore_uri` - (Required) Amazon S3 URL that specifies the truststore for mutual TLS authentication, for example, `s3://bucket-name/key-name`. The truststore can contain certificates from public or private certificate authorities. To update the truststore, upload a new version to S3, and then update your custom domain name to use the new version.
* `truststore_version` - (Optional) Version of the S3 object that contains the truststore. To specify a version, you must have versioning enabled for the S3 bucket.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of domain name.
* `certificate_upload_date` - Upload date associated with the domain certificate.
* `cloudfront_domain_name` - Hostname created by Cloudfront to represent the distribution that implements this domain name mapping.
* `cloudfront_zone_id` - For convenience, the hosted zone ID (`Z2FDTNDATAQYW2`) that can be used to create a Route53 alias record for the distribution.
* `domain_name_id` - The identifier for the domain name resource. Supported only for private custom domain names.
* `id` - Internal identifier assigned to this domain name by API Gateway.
* `regional_domain_name` - Hostname for the custom domain's regional endpoint.
* `regional_zone_id` - Hosted zone ID that can be used to create a Route53 alias record for the regional endpoint.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `60m`)
- `update` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway domain names using their `name` or `name` and `domain_name_id` (for private custom domain names). For example:

```terraform
import {
  to = aws_api_gateway_domain_name.example
  id = "dev.example.com"
}
```

For a private custom domain name:

```terraform
import {
  to = aws_api_gateway_domain_name.example
  id = "api.internal.example.com/abcde12345"
}
```

Using `terraform import`, import API Gateway domain names using their `name` or `name` and `domain_name_id` (for private custom domain names). For example:

```console
% terraform import aws_api_gateway_domain_name.example dev.example.com
```

For a private custom domain name:

```console
% terraform import aws_api_gateway_domain_name.example dev.api.internal.example.com/abcde12345
```
