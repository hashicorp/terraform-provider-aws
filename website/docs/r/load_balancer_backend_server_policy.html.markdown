---
subcategory: "ELB Classic"
layout: "aws"
page_title: "AWS: aws_load_balancer_backend_server_policy"
description: |-
  Attaches a load balancer policy to an ELB backend server.
---

# Resource: aws_load_balancer_backend_server_policy

Attaches a load balancer policy to an ELB backend server.

## Example Usage

```terraform
resource "aws_elb" "wu-tang" {
  name               = "wu-tang"
  availability_zones = ["us-east-1a"]

  listener {
    instance_port      = 443
    instance_protocol  = "http"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = "arn:aws:iam::000000000000:server-certificate/wu-tang.net"
  }

  tags = {
    Name = "wu-tang"
  }
}

resource "aws_load_balancer_policy" "wu-tang-ca-pubkey-policy" {
  load_balancer_name = aws_elb.wu-tang.name
  policy_name        = "wu-tang-ca-pubkey-policy"
  policy_type_name   = "PublicKeyPolicyType"

  # The public key of a CA certificate file can be extracted with:
  # $ cat wu-tang-ca.pem | openssl x509 -pubkey -noout | grep -v '\-\-\-\-' | tr -d '\n' > wu-tang-pubkey
  policy_attribute {
    name  = "PublicKey"
    value = file("wu-tang-pubkey")
  }
}

resource "aws_load_balancer_policy" "wu-tang-root-ca-backend-auth-policy" {
  load_balancer_name = aws_elb.wu-tang.name
  policy_name        = "wu-tang-root-ca-backend-auth-policy"
  policy_type_name   = "BackendServerAuthenticationPolicyType"

  policy_attribute {
    name  = "PublicKeyPolicyName"
    value = aws_load_balancer_policy.wu-tang-root-ca-pubkey-policy.policy_name
  }
}

resource "aws_load_balancer_backend_server_policy" "wu-tang-backend-auth-policies-443" {
  load_balancer_name = aws_elb.wu-tang.name
  instance_port      = 443

  policy_names = [
    aws_load_balancer_policy.wu-tang-root-ca-backend-auth-policy.policy_name,
  ]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `load_balancer_name` - (Required) The load balancer to attach the policy to.
* `policy_names` - (Required) List of Policy Names to apply to the backend server.
* `instance_port` - (Required) The instance port to apply the policy to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the policy.
* `load_balancer_name` - The load balancer on which the policy is defined.
* `instance_port` - The backend port the policies are applied to
