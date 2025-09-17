---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_stop_instance"
description: |-
  Stops an EC2 instance.
---

# Action: aws_ec2_stop_instance

Stops an EC2 instance. This action will gracefully stop the instance and wait for it to reach the stopped state.

For information about Amazon EC2, see the [Amazon EC2 User Guide](https://docs.aws.amazon.com/ec2/latest/userguide/). For specific information about stopping instances, see the [StopInstances](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_StopInstances.html) page in the Amazon EC2 API Reference.

~> **Note:** `aws_ec2_stop_instance` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

~> **Note:** This action directly stops EC2 instances which will interrupt running workloads. Ensure proper coordination with your applications before using this action.

## Example Usage

### Basic Usage

```terraform
resource "aws_instance" "example" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"

  tags = {
    Name = "example-instance"
  }
}

action "aws_ec2_stop_instance" "example" {
  config {
    instance_id = aws_instance.example.id
  }
}
```

### Force Stop

```terraform
action "aws_ec2_stop_instance" "force_stop" {
  config {
    instance_id = aws_instance.example.id
    force       = true
    timeout     = 300
  }
}
```

### Maintenance Window

```terraform
resource "aws_instance" "web_server" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"

  tags = {
    Name = "web-server"
  }
}

action "aws_ec2_stop_instance" "maintenance" {
  config {
    instance_id = aws_instance.web_server.id
    timeout     = 900
  }
}

resource "terraform_data" "maintenance_trigger" {
  input = var.maintenance_window

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_ec2_stop_instance.maintenance]
    }
  }

  depends_on = [aws_instance.web_server]
}
```

## Argument Reference

This action supports the following arguments:

* `instance_id` - (Required) ID of the EC2 instance to stop. Must be a valid EC2 instance ID (e.g., i-1234567890abcdef0).
* `force` - (Optional) Forces the instance to stop. The instance does not have an opportunity to flush file system caches or file system metadata. If you use this option, you must perform file system check and repair procedures. This option is not recommended for Windows instances. Default: `false`.
* `timeout` - (Optional) Timeout in seconds to wait for the instance to stop. Must be between 30 and 3600 seconds. Default: `600`.
