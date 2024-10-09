---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_domain"
description: |-
  Provides a Cognito User Pool Domain resource.
---

# Resource: aws_cognito_user_pool_domain

Provides a Cognito User Pool Domain resource.

## Example Usage

### Amazon Cognito domain

```terraform
resource "aws_cognito_user_pool_domain" "main" {
  domain       = "example-domain"
  user_pool_id = aws_cognito_user_pool.example.id
}

resource "aws_cognito_user_pool" "example" {
  name = "example-pool"
}
```

### Custom Cognito domain

```terraform
resource "aws_cognito_user_pool_domain" "main" {
  domain          = "example-domain"
  certificate_arn = aws_acm_certificate.cert.arn
  user_pool_id    = aws_cognito_user_pool.example.id
}

resource "aws_cognito_user_pool" "example" {
  name = "example-pool"
}

data "aws_route53_zone" "example" {
  name = "example.com"
}

resource "aws_route53_record" "auth-cognito-A" {
  name    = aws_cognito_user_pool_domain.main.domain
  type    = "A"
  zone_id = data.aws_route53_zone.example.zone_id
  alias {
    evaluate_target_health = false

    name    = aws_cognito_user_pool_domain.main.cloudfront_distribution
    zone_id = aws_cognito_user_pool_domain.main.cloudfront_distribution_zone_id
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `domain` - (Required) For custom domains, this is the fully-qualified domain name, such as auth.example.com. For Amazon Cognito prefix domains, this is the prefix alone, such as auth.
* `user_pool_id` - (Required) The user pool ID.
* `certificate_arn` - (Optional) The ARN of an ISSUED ACM certificate in us-east-1 for a custom domain.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `aws_account_id` - The AWS account ID for the user pool owner.
* `cloudfront_distribution` - The Amazon CloudFront endpoint (e.g. `dpp0gtxikpq3y.cloudfront.net`) that you use as the target of the alias that you set up with your Domain Name Service (DNS) provider.
* `cloudfront_distribution_arn` - The URL of the CloudFront distribution. This is required to generate the ALIAS `aws_route53_record`
* `cloudfront_distribution_zone_id` - The Route 53 hosted zone ID of the CloudFront distribution.
* `s3_bucket` - The S3 bucket where the static files for this domain are stored.
* `version` - The app version.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cognito User Pool Domains using the `domain`. For example:

```terraform
import {
  to = aws_cognito_user_pool_domain.main
  id = "auth.example.org"
}
```

Using `terraform import`, import Cognito User Pool Domains using the `domain`. For example:

```console
% terraform import aws_cognito_user_pool_domain.main auth.example.org
```
