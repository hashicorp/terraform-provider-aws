---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_thing_principal_attachment"
description: |-
    Provides AWS IoT Thing Principal attachment.
---

# Resource: aws_iot_thing_principal_attachment

Attaches Principal to AWS IoT Thing.

## Example Usage

```terraform
resource "aws_iot_thing" "example" {
  name = "example"
}

resource "aws_iot_certificate" "cert" {
  csr    = file("csr.pem")
  active = true
}

resource "aws_iot_thing_principal_attachment" "att" {
  principal = aws_iot_certificate.cert.arn
  thing     = aws_iot_thing.example.name
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `principal` - (Required) The AWS IoT Certificate ARN or Amazon Cognito Identity ID.
* `thing` - (Required) The name of the thing.

## Attribute Reference

This resource exports no additional attributes.
