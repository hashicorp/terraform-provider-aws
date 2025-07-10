---
subcategory: "Chime"
layout: "aws"
page_title: "AWS: aws_chime_voice_connector_origination"
description: |-
    Enable origination settings to control inbound calling to your SIP infrastructure.
---

# Resource: aws_chime_voice_connector_origination

Enable origination settings to control inbound calling to your SIP infrastructure.

## Example Usage

```terraform
resource "aws_chime_voice_connector" "default" {
  name               = "test"
  require_encryption = true
}

resource "aws_chime_voice_connector_origination" "default" {
  disabled           = false
  voice_connector_id = aws_chime_voice_connector.default.id

  route {
    host     = "127.0.0.1"
    port     = 8081
    protocol = "TCP"
    priority = 1
    weight   = 1
  }

  route {
    host     = "127.0.0.2"
    port     = 8082
    protocol = "TCP"
    priority = 2
    weight   = 10
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `voice_connector_id` - (Required) The Amazon Chime Voice Connector ID.
* `route` - (Required) Set of call distribution properties defined for your SIP hosts. See [route](#route) below for more details. Minimum of 1. Maximum of 20.
* `disabled` - (Optional) When origination settings are disabled, inbound calls are not enabled for your Amazon Chime Voice Connector.

### `route`

Origination routes define call distribution properties for your SIP hosts to receive inbound calls using your Amazon Chime Voice Connector. Limit: Ten origination routes for each Amazon Chime Voice Connector.

* `host` - (Required) The FQDN or IP address to contact for origination traffic.
* `port` - (Required) The designated origination route port. Defaults to `5060`.
* `priority` - (Required) The priority associated with the host, with 1 being the highest priority. Higher priority hosts are attempted first.
* `protocol` - (Required) The protocol to use for the origination route. Encryption-enabled Amazon Chime Voice Connectors use TCP protocol by default.
* `weight` - (Required) The weight associated with the host. If hosts are equal in priority, calls are redistributed among them based on their relative weight.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Chime Voice Connector ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Chime Voice Connector Origination using the `voice_connector_id`. For example:

```terraform
import {
  to = aws_chime_voice_connector_origination.default
  id = "abcdef1ghij2klmno3pqr4"
}
```

Using `terraform import`, import Chime Voice Connector Origination using the `voice_connector_id`. For example:

```console
% terraform import aws_chime_voice_connector_origination.default abcdef1ghij2klmno3pqr4
```
