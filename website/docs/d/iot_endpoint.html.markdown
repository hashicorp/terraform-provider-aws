---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_endpoint"
description: |-
  Get the unique IoT endpoint
---

# Data Source: aws_iot_endpoint

Returns a unique endpoint specific to the AWS account making the call.

## Example Usage

```terraform
data "aws_iot_endpoint" "example" {}

resource "kubernetes_pod" "agent" {
  metadata {
    name = "my-device"
  }

  spec {
    container {
      image = "gcr.io/my-project/image-name"
      name  = "image-name"

      env {
        name  = "IOT_ENDPOINT"
        value = data.aws_iot_endpoint.example.endpoint_address
      }
    }
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `endpoint_type` - (Optional) Endpoint type. Valid values: `iot:CredentialProvider`, `iot:Data`, `iot:Data-ATS`, `iot:Jobs`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `endpoint_address` - Endpoint based on `endpoint_type`:
    * No `endpoint_type`: Either `iot:Data` or `iot:Data-ATS` [depending on region](https://aws.amazon.com/blogs/iot/aws-iot-core-ats-endpoints/)
    * `iot:CredentialsProvider`: `IDENTIFIER.credentials.iot.REGION.amazonaws.com`
    * `iot:Data`: `IDENTIFIER.iot.REGION.amazonaws.com`
    * `iot:Data-ATS`: `IDENTIFIER-ats.iot.REGION.amazonaws.com`
    * `iot:Jobs`: `IDENTIFIER.jobs.iot.REGION.amazonaws.com`
