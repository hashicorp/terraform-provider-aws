---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_query_log_config"
description: |-
  Provides a Route 53 Resolver query logging configuration resource.
---

# Resource: aws_route53_resolver_query_log_config

Provides a Route 53 Resolver query logging configuration resource.

## Example Usage

```terraform
resource "aws_route53_resolver_query_log_config" "example" {
  name            = "example"
  destination_arn = aws_s3_bucket.example.arn

  tags = {
    Environment = "Prod"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `destination_arn` - (Required) The ARN of the resource that you want Route 53 Resolver to send query logs.
You can send query logs to an [S3 bucket](s3_bucket.html), a [CloudWatch Logs log group](cloudwatch_log_group.html), or a [Kinesis Data Firehose delivery stream](kinesis_firehose_delivery_stream.html).
* `name` - (Required) The name of the Route 53 Resolver query logging configuration.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the Route 53 Resolver query logging configuration.
* `arn` - The ARN (Amazon Resource Name) of the Route 53 Resolver query logging configuration.
* `owner_id` - The AWS account ID of the account that created the query logging configuration.
* `share_status` - An indication of whether the query logging configuration is shared with other AWS accounts, or was shared with the current account by another AWS account.
Sharing is configured through AWS Resource Access Manager (AWS RAM).
Values are `NOT_SHARED`, `SHARED_BY_ME` or `SHARED_WITH_ME`
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import  Route 53 Resolver query logging configurations using the Route 53 Resolver query logging configuration ID. For example:

```terraform
import {
  to = aws_route53_resolver_query_log_config.example
  id = "rqlc-92edc3b1838248bf"
}
```

Using `terraform import`, import  Route 53 Resolver query logging configurations using the Route 53 Resolver query logging configuration ID. For example:

```console
% terraform import aws_route53_resolver_query_log_config.example rqlc-92edc3b1838248bf
```
