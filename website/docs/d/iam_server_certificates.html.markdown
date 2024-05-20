---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_server_certificates"
description: |-
  List all server certificates
---

# Data Source: aws_iam_server_certificates

Use this data source to list IAM Server Certificates.

## Example Usage

```terraform
data "aws_iam_server_certificates" "my-domain" {
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

* `name_prefix` - prefix of cert to filter by
* `path_prefix` - prefix of path to filter by
* `name` - exact name of the cert to lookup
* `latest` - sort results by expiration date. returns certificates with expiration date furthest in the future first.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` is set to the unique id of the IAM Server Certificate
* `arn` is set to the ARN of the IAM Server Certificate
* `path` is set to the path of the IAM Server Certificate
* `expiration_date` is set to the expiration date of the IAM Server Certificate
* `upload_date` is the date when the server certificate was uploaded

