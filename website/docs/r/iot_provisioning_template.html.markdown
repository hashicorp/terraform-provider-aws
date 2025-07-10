---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_provisioning_template"
description: |-
    Manages an IoT fleet provisioning template.
---

# Resource: aws_iot_provisioning_template

Manages an IoT fleet provisioning template. For more info, see the AWS documentation on [fleet provisioning](https://docs.aws.amazon.com/iot/latest/developerguide/provision-wo-cert.html).

## Example Usage

```terraform
data "aws_iam_policy_document" "iot_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["iot.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "iot_fleet_provisioning" {
  name               = "IoTProvisioningServiceRole"
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.iot_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "iot_fleet_provisioning_registration" {
  role       = aws_iam_role.iot_fleet_provisioning.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSIoTThingsRegistration"
}

data "aws_iam_policy_document" "device_policy" {
  statement {
    actions   = ["iot:Subscribe"]
    resources = ["*"]
  }
}

resource "aws_iot_policy" "device_policy" {
  name   = "DevicePolicy"
  policy = data.aws_iam_policy_document.device_policy.json
}

resource "aws_iot_provisioning_template" "fleet" {
  name                  = "FleetTemplate"
  description           = "My provisioning template"
  provisioning_role_arn = aws_iam_role.iot_fleet_provisioning.arn
  enabled               = true

  template_body = jsonencode({
    Parameters = {
      SerialNumber = { Type = "String" }
    }

    Resources = {
      certificate = {
        Properties = {
          CertificateId = { Ref = "AWS::IoT::Certificate::Id" }
          Status        = "Active"
        }
        Type = "AWS::IoT::Certificate"
      }

      policy = {
        Properties = {
          PolicyName = aws_iot_policy.device_policy.name
        }
        Type = "AWS::IoT::Policy"
      }
    }
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the fleet provisioning template.
* `description` - (Optional) The description of the fleet provisioning template.
* `enabled` - (Optional) True to enable the fleet provisioning template, otherwise false.
* `pre_provisioning_hook` - (Optional) Creates a pre-provisioning hook template. Details below.
* `provisioning_role_arn` - (Required) The role ARN for the role associated with the fleet provisioning template. This IoT role grants permission to provision a device.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `template_body` - (Required) The JSON formatted contents of the fleet provisioning template.
* `type` - (Optional) The type you define in a provisioning template.

### pre_provisioning_hook

The `pre_provisioning_hook` configuration block supports the following:

* `payload_version` - (Optional) The version of the payload that was sent to the target function. The only valid (and the default) payload version is `"2020-04-01"`.
* `target_arn` - (Optional) The ARN of the target function.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN that identifies the provisioning template.
* `default_version_id` - The default version of the fleet provisioning template.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IoT fleet provisioning templates using the `name`. For example:

```terraform
import {
  to = aws_iot_provisioning_template.fleet
  id = "FleetProvisioningTemplate"
}
```

Using `terraform import`, import IoT fleet provisioning templates using the `name`. For example:

```console
% terraform import aws_iot_provisioning_template.fleet FleetProvisioningTemplate
```
