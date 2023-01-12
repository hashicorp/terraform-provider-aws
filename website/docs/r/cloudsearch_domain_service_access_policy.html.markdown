---
subcategory: "CloudSearch"
layout: "aws"
page_title: "AWS: aws_cloudsearch_domain_service_access_policy"
description: |-
  Provides an CloudSearch domain service access policy resource. 
---

# Resource: aws_cloudsearch_domain_service_access_policy

Provides an CloudSearch domain service access policy resource.

Terraform waits for the domain service access policy to become `Active` when applying a configuration.

## Example Usage

```terraform
resource "aws_cloudsearch_domain" "example" {
  name = "example-domain"
}

resource "aws_cloudsearch_domain_service_access_policy" "example" {
  domain_name = aws_cloudsearch_domain.example.id

  access_policy = <<POLICY
{
  "Version":"2012-10-17",
  "Statement":[{
    "Sid":"search_only",
    "Effect":"Allow",
    "Principal":"*",
    "Action":[
      "cloudsearch:search",
      "cloudsearch:document"
    ],
    "Condition":{"IpAddress":{"aws:SourceIp":"192.0.2.0/32"}}
  }]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `access_policy` - (Required) The access rules you want to configure. These rules replace any existing rules. See the [AWS documentation](https://docs.aws.amazon.com/cloudsearch/latest/developerguide/configuring-access.html) for details.
* `domain_name` - (Required) The CloudSearch domain name the policy applies to.

## Attributes Reference

No additional attributes are exported.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `update` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

CloudSearch domain service access policies can be imported using the domain name, e.g.,

```
$ terraform import aws_cloudsearch_domain_service_access_policy.example example-domain
```
