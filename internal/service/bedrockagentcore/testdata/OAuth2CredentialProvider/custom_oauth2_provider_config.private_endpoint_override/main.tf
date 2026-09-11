# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = var.rName

  credential_provider_vendor = "CustomOauth2"

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id_wo                  = "private-endpoint-client-id"
      client_secret_wo              = "private-endpoint-client-secret"
      client_credentials_wo_version = 1

      oauth_discovery {
        discovery_url = "https://dev-example.auth0.com/.well-known/openid-configuration"
      }

      private_endpoint {
        managed_vpc_resource {
          endpoint_ip_address_type = "IPV4"
          subnet_ids               = [aws_subnet.test[0].id]
          vpc_identifier           = aws_vpc.test[0].id
          tags = {
            Cause = "test"
          }
        }
      }

      private_endpoint_override {
        domain = "example.com"

        private_endpoint {
          managed_vpc_resource {
            endpoint_ip_address_type = "IPV4"
            subnet_ids               = [aws_subnet.test[1].id]
            vpc_identifier           = aws_vpc.test[1].id
            security_group_ids       = [aws_security_group.test[1].id]
          }
        }
      }
    }
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  count = 2

  cidr_block = "10.${count.index}.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test[count.index].id
  cidr_block        = "10.${count.index}.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = var.rName
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test[count.index].id
  name   = "${var.rName}-${count.index}"

  tags = {
    Name = var.rName
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
