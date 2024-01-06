---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_trust_store"
description: |-
  Provides a Trust Store resource for use with Load Balancers.
---

# Resource: aws_lb_trust_store

Provides a ELBv2 Trust Store for use with Application Load Balancer Listener resources.

## Example Usage

### Trust Store Load Balancer Listener

```terraform
resource "aws_lb_trust_store" "test" {
  name = "tf-example-lb-ts"

  ca_certificates_bundle_s3_bucket = "..."
  ca_certificates_bundle_s3_key    = "..."

}

resource "aws_lb_listener" "example" {
  load_balancer_arn = aws_lb.example.id

  default_action {
    target_group_arn = aws_lb_target_group.example.id
    type             = "forward"
  }

  mutual_authentication = {
    mode            = "verify"
    trust_store_arn = aws_lb_trust_store.test.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `ca_certificates_bundle_s3_bucket` - (Required) S3 Bucket name holding the client certificate CA bundle.
* `ca_certificates_bundle_s3_key` - (Required) S3 object key holding the client certificate CA bundle.
* `ca_certificates_bundle_s3_object_version` - (Optional) Version Id of CA bundle S3 bucket object, if versioned, defaults to latest if omitted.

* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`. Cannot be longer than 6 characters.
* `name` - (Optional, Forces new resource) Name of the Trust Store. If omitted, Terraform will assign a random, unique name. This name must be unique per region per account, can have a maximum of 32 characters, must contain only alphanumeric characters or hyphens, and must not begin or end with a hyphen.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn_suffix` - ARN suffix for use with CloudWatch Metrics.
* `arn` - ARN of the Trust Store (matches `id`).
* `id` - ARN of the Trust Store (matches `arn`).
* `name` - Name of the Trust Store.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Trust Stores using their ARN. For example:

```terraform
import {
  to = aws_lb_trust_store.example
  id = "arn:aws:elasticloadbalancing:us-west-2:187416307283:truststore/my-trust-store/20cfe21448b66314"
}
```

Using `terraform import`, import Target Groups using their ARN. For example:

```console
% terraform import aws_lb_trust_store.example arn:aws:elasticloadbalancing:us-west-2:187416307283:truststore/my-trust-store/20cfe21448b66314
```
