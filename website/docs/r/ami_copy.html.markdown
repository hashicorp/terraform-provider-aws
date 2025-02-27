---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ami_copy"
description: |-
  Duplicates an existing Amazon Machine Image (AMI)
---

# Resource: aws_ami_copy

The "AMI copy" resource allows duplication of an Amazon Machine Image (AMI),
including cross-region copies.

If the source AMI has associated EBS snapshots, those will also be duplicated
along with the AMI.

This is useful for taking a single AMI provisioned in one region and making
it available in another for a multi-region deployment.

Copying an AMI can take several minutes. The creation of this resource will
block until the new AMI is available for use on new instances.

## Example Usage

```terraform
resource "aws_ami_copy" "example" {
  name              = "terraform-example"
  source_ami_id     = "ami-xxxxxxxx"
  source_ami_region = "us-west-1"

  tags = {
    Name = "HelloWorld"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Region-unique name for the AMI.
* `source_ami_id` - (Required) Id of the AMI to copy. This id must be valid in the region
  given by `source_ami_region`.
* `source_ami_region` - (Required) Region from which the AMI will be copied. This may be the
  same as the AWS provider region in order to create a copy within the same region.
* `destination_outpost_arn` - (Optional) ARN of the Outpost to which to copy the AMI.
  Only specify this parameter when copying an AMI from an AWS Region to an Outpost. The AMI must be in the Region of the destination Outpost.  
* `encrypted` - (Optional) Whether the destination snapshots of the copied image should be encrypted. Defaults to `false`
* `kms_key_id` - (Optional) Full ARN of the KMS Key to use when encrypting the snapshots of an image during a copy operation. If not specified, then the default AWS KMS Key will be used
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

This resource also exposes the full set of arguments from the [`aws_ami`](ami.html) resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the AMI.
* `id` - ID of the created AMI.

This resource also exports a full set of attributes corresponding to the arguments of the
[`aws_ami`](/docs/providers/aws/r/ami.html) resource, allowing the properties of the created AMI to be used elsewhere in the
configuration.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `40m`)
* `update` - (Default `40m`)
* `delete` - (Default `90m`)
