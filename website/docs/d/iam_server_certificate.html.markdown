---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_server_certificate"
description: |-
  Get information about a server certificate
---

# Data Source: aws_iam_server_certificate

Use this data source to lookup information about IAM Server Certificates.

## Example Usage

```terraform
data "aws_iam_server_certificate" "my-domain" {
  name_prefix = "my-domain.org"
  latest      = true
}

resource "aws_elb" "elb" {
  name = "my-domain-elb"

  listener {
    instance_port      = 8000
    instance_protocol  = "https"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = data.aws_iam_server_certificate.my-domain.arn
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `name_prefix` - prefix of cert to filter by
* `path_prefix` - prefix of path to filter by
* `name` - exact name of the cert to lookup
* `latest` - sort results by expiration date. returns the certificate with expiration date in furthest in the future.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` is set to the unique id of the IAM Server Certificate
* `arn` is set to the ARN of the IAM Server Certificate
* `path` is set to the path of the IAM Server Certificate
* `expiration_date` is set to the expiration date of the IAM Server Certificate
* `upload_date` is the date when the server certificate was uploaded
* `certificate_body` is the public key certificate (PEM-encoded). This is useful when [configuring back-end instance authentication](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-create-https-ssl-load-balancer.html) policy for load balancer
* `certificate_chain` is the public key certificate chain (PEM-encoded) if exists, empty otherwise

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an IAM server certificate using `name`. For example:

```terraform
import {
  to = aws_iam_server_certificate.example
  id = "example"
}
```

Using `terraform import`, import an IAM server certificate using `name`. For example:

```console
% terraform import aws_iam_server_certificate.example example
```

Import will read in the certificate body, certificate chain (if it exists), ID, name, path, and ARN. It will not retrieve the private key which is not available through the AWS API.
