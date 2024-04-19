---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_traffic_policy"
description: |-
    Manages a Route53 Traffic Policy
---

# Resource: aws_route53_traffic_policy

Manages a Route53 Traffic Policy.

## Example Usage

```terraform
resource "aws_route53_traffic_policy" "example" {
  name     = "example"
  comment  = "example comment"
  document = <<EOF
{
  "AWSPolicyFormatVersion": "2015-10-01",
  "RecordType": "A",
  "Endpoints": {
    "endpoint-start-NkPh": {
      "Type": "value",
      "Value": "10.0.0.2"
    }
  },
  "StartEndpoint": "endpoint-start-NkPh"
}
EOF
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the traffic policy.
* `document` - (Required) Policy document. This is a JSON formatted string. For more information about building Route53 traffic policy documents, see the [AWS Route53 Traffic Policy document format](https://docs.aws.amazon.com/Route53/latest/APIReference/api-policies-traffic-policy-document-format.html)

The following arguments are optional:

* `comment` - (Optional) Comment for the traffic policy.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the traffic policy
* `type` - DNS type of the resource record sets that Amazon Route 53 creates when you use a traffic policy to create a traffic policy instance.
* `version` - Version number of the traffic policy. This value is automatically incremented by AWS after each update of this resource.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Traffic Policy using the `id` and `version`. For example:

```terraform
import {
  to = aws_route53_traffic_policy.example
  id = "01a52019-d16f-422a-ae72-c306d2b6df7e/1"
}
```

Using `terraform import`, import Route53 Traffic Policy using the `id` and `version`. For example:

```console
% terraform import aws_route53_traffic_policy.example 01a52019-d16f-422a-ae72-c306d2b6df7e/1
```
