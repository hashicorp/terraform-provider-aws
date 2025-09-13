---
subcategory: "ELB Classic"
layout: "aws"
page_title: "AWS: aws_load_balancer_policy"
description: |-
  Provides a load balancer policy, which can be attached to an ELB listener or backend server.
---

# Resource: aws_load_balancer_policy

Provides a load balancer policy, which can be attached to an ELB listener or backend server.

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

resource "aws_load_balancer_policy" "wu-tang-ssl" {
  load_balancer_name = aws_elb.wu-tang.name
  policy_name        = "wu-tang-ssl"
  policy_type_name   = "SSLNegotiationPolicyType"

  policy_attribute {
    name  = "ECDHE-ECDSA-AES128-GCM-SHA256"
    value = "true"
  }

  policy_attribute {
    name  = "Protocol-TLSv1.2"
    value = "true"
  }
}

resource "aws_load_balancer_policy" "wu-tang-ssl-tls-1-1" {
  load_balancer_name = aws_elb.wu-tang.name
  policy_name        = "wu-tang-ssl"
  policy_type_name   = "SSLNegotiationPolicyType"

  policy_attribute {
    name  = "Reference-Security-Policy"
    value = "ELBSecurityPolicy-TLS-1-1-2017-01"
  }
}

resource "aws_load_balancer_backend_server_policy" "wu-tang-backend-auth-policies-443" {
  load_balancer_name = aws_elb.wu-tang.name
  instance_port      = 443

  policy_names = [
    aws_load_balancer_policy.wu-tang-root-ca-backend-auth-policy.policy_name,
  ]
}

resource "aws_load_balancer_listener_policy" "wu-tang-listener-policies-443" {
  load_balancer_name = aws_elb.wu-tang.name
  load_balancer_port = 443

  policy_names = [
    aws_load_balancer_policy.wu-tang-ssl.policy_name,
  ]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `load_balancer_name` - (Required) The load balancer on which the policy is defined.
* `policy_name` - (Required) The name of the load balancer policy.
* `policy_type_name` - (Required) The policy type.
* `policy_attribute` - (Optional) Policy attribute to apply to the policy.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the policy.
* `policy_name` - The name of the stickiness policy.
* `policy_type_name` - The policy type of the policy.
* `load_balancer_name` - The load balancer on which the policy is defined.
