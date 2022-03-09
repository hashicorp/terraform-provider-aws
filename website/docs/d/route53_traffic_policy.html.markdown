---
subcategory: "Route53"
layout: "aws"
page_title: "AWS: aws_route53_zone"
description: |-
    Provides details about a specific Route 53 Hosted Zone
---

# Data Source: aws_route53_zone

`aws_route53_zone` provides details about a specific Route 53 Hosted Zone.

This data source allows to find a Hosted Zone ID given Hosted Zone name and certain search criteria.

## Example Usage

The following example shows how to get a Hosted Zone from its name and from this data how to create a Record Set.


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


data "aws_route53_traffic_policy" "example" {
  traffic_policy_id = aws_route53_traffic_policy.test.id
}
```

## Argument Reference

* `traffic_policy_id` - (Required) ID of the traffic policy.

## Attributes Reference

The following attribute is additionally exported:

* `id` - ID of the traffic policy
* `comment` - Comment for the traffic policy.
* `document` - Policy document. This is a JSON formatted string. For more information about building Route53 traffic policy documents, see the [AWS Route53 Traffic Policy document format](https://docs.aws.amazon.com/Route53/latest/APIReference/api-policies-traffic-policy-document-format.html)
* `name` - Name of the traffic policy.
* `type` - DNS type of the resource record sets that Amazon Route 53 creates when you use a traffic policy to create a traffic policy instance.
* `version` - Version number of the traffic policy. This value is automatically incremented by AWS after each update of this resource.
