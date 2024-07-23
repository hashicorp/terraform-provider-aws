---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_trust_store_revocation"
description: |-
  Provides a Trust Store Revocation resource for use with Load Balancers.
---

# Resource: aws_lb_trust_store_revocation

Provides a ELBv2 Trust Store Revocation for use with Application Load Balancer Listener resources.

## Example Usage

### Trust Store With Revocations

```terraform
resource "aws_lb_trust_store" "test" {
  name = "tf-example-lb-ts"

  ca_certificates_bundle_s3_bucket = "..."
  ca_certificates_bundle_s3_key    = "..."

}

resource "aws_lb_trust_store_revocation" "test" {
  trust_store_arn = aws_lb_trust_store.test.arn

  revocations_s3_bucket = "..."
  revocations_s3_key    = "..."

}

```

## Argument Reference

This resource supports the following arguments:

* `trust_store_arn` - (Required) Trust Store ARN.
* `revocations_s3_bucket` - (Required) S3 Bucket name holding the client certificate CA bundle.
* `revocations_s3_key` - (Required) S3 object key holding the client certificate CA bundle.
* `revocations_s3_object_version` - (Optional) Version Id of CA bundle S3 bucket object, if versioned, defaults to latest if omitted.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `revocation_id` - AWS assigned RevocationId, (number).
* `id` - "combination of the Trust Store ARN and RevocationId `${trust_store_arn},{revocation_id}`"

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Trust Store Revocations using their ARN. For example:

```terraform
import {
  to = aws_lb_trust_store_revocation.example
  id = "arn:aws:elasticloadbalancing:us-west-2:187416307283:truststore/my-trust-store/20cfe21448b66314,6"
}
```

Using `terraform import`, import Trust Store Revocations using their ARN. For example:

```console
% terraform import aws_lb_trust_store_revocation.example arn:aws:elasticloadbalancing:us-west-2:187416307283:truststore/my-trust-store/20cfe21448b66314,6
```
