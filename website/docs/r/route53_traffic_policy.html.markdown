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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the traffic policy
* `type` - DNS type of the resource record sets that Amazon Route 53 creates when you use a traffic policy to create a traffic policy instance.
* `version` - Version number of the traffic policy. This value is automatically incremented by AWS after each update of this resource.

## Import

Route53 Traffic Policy can be imported using the `id` and `version`, e.g.

```
$ terraform import aws_route53_traffic_policy.example 01a52019-d16f-422a-ae72-c306d2b6df7e/1
```
