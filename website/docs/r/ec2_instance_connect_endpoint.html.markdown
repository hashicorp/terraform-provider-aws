---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_connect_endpoint"
description: |-
  Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Instance Connect Endpoint.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
# Resource: aws_ec2_instance_connect_endpoint

Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Instance Connect Endpoint.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "example" {
  vpc_id = aws_vpc.example.id
}

resource "aws_subnet" "example" {
 vpc_id     = aws_vpc.example.id
 cidr_block = "10.0.1.0/24"
}

resource "aws_ec2_instance_connect_endpoint" "example" {
  subnet_id          = aws_subnet.example.id
  security_group_ids = [aws_security_group.example.id]	
  preserve_client_ip = false
}
```

## Argument Reference

The following arguments are required:

* `subnet_id` - (Required) The ID of the subnet in which to create the EC2 Instance Connect Endpoint.

The following arguments are optional:

* `security_group_ids` - (Optional) One or more security groups to associate with the endpoint. If you don't specify a security group, the default security group for your VPC will be associated with the endpoint.If no security groups are specified, the VPC's default security group is associated with the endpoint.

* `preserve_client_ip` - (Optional) Indicates whether your client's IP address is preserved as the source. The value is true or false. Defaults to `true`.
  * If true, your client's IP address is used when you connect to a resource.
  * If false, the elastic network interface IP address is used when you connect to a resource.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the EC2 Instance Connect endpoint.
* `arn` - The Amazon Resource Name for EC2 Instance Connect endpoint.
* `availability_zone` - The Availability Zone of the EC2 Instance Connect Endpoint.
* `dns_name` - The DNS name of the EC2 Instance Connect Endpoint.
* `fips_dns_name` - The FIPS DNS name.
* `network_interface_ids` - The ID of the elastic network interface that Amazon EC2 automatically created when creating the EC2 Instance Connect Endpoint.
* `owner_id` - The ID of the Amazon Web Services account that created the EC2 Instance Connect Endpoint.
* `state` - The state of the EC2 Instance Connect endpoint.
* `state_message` - The message for the current state of the EC2 Instance Connect Endpoint. Can include a failure message.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

EC2 (Elastic Compute Cloud) Instance Connect Endpoint can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_ec2_instance_connect_endpoint.example eice-02d2b75e650eaa75a
```
