# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "99.99.99"
    }
  }
}

provider "aws" {
  region = var.region
}

variable "region" {
  description = "AWS region for GameLift resources."
  type        = string
  default     = "us-east-1"
}

variable "name_prefix" {
  description = "Prefix used in GameLift resource names. Keep this unique per test run."
  type        = string
  default     = "my-game-gamelift"
}

variable "fleet_role_arn" {
  description = "IAM role ARN for the fleet with GameLift container fleet permissions (for example GameLiftContainerFleetPolicy)."
  type        = string
  default     = "arn:aws:iam::123456789012:role/gamelift-container-fleet-role"
}

variable "image_uri" {
  description = "ECR image URI for the game server container (for example 123456789012.dkr.ecr.us-west-2.amazonaws.com/game-server:latest)."
  type        = string
  default     = "123456789012.dkr.ecr.us-east-1.amazonaws.com/game-server:latest"
}

variable "server_sdk_version" {
  description = "GameLift server SDK version for the game server container definition."
  type        = string
  default     = "5.2.0"
}

variable "instance_type" {
  description = "EC2 instance type for the container fleet."
  type        = string
  default     = "c7a.medium"
}

variable "operating_system" {
  description = "Operating system for the container group."
  type        = string
  default     = "AMAZON_LINUX_2023"
}

variable "total_memory_limit_mib" {
  description = "Total memory limit for the container group."
  type        = number
  default     = 1024
}

variable "total_vcpu_limit" {
  description = "Total vCPU limit for the container group."
  type        = number
  default     = 1
}

variable "game_server_from_port" {
  description = "Game server container port range start."
  type        = number
  default     = 20000
}

variable "game_server_to_port" {
  description = "Game server container port range end."
  type        = number
  default     = 20040
}

variable "game_server_port_protocol" {
  description = "Game server container port protocol."
  type        = string
  default     = "UDP"
}

variable "billing_type" {
  description = "Container fleet billing type."
  type        = string
  default     = "SPOT"
}

variable "fleet_description" {
  description = "Container fleet description."
  type        = string
  default     = "Example game container fleet"
}

variable "locations" {
  description = "Container fleet deployment locations."
  type        = list(string)
  default     = ["us-east-1"]
}

variable "game_server_container_groups_per_instance" {
  description = "Number of game server container groups per instance."
  type        = number
  default     = 1
}

variable "instance_connection_from_port" {
  description = "Fleet instance connection port range start."
  type        = number
  default     = 30000
}

variable "instance_connection_to_port" {
  description = "Fleet instance connection port range end."
  type        = number
  default     = 30099
}

resource "aws_gamelift_container_group_definition" "this" {
  name                   = "${var.name_prefix}-group"
  container_group_type   = "GAME_SERVER"
  operating_system       = var.operating_system
  total_memory_limit_mib = var.total_memory_limit_mib
  total_vcpu_limit       = var.total_vcpu_limit

  game_server_container_definition {
    container_name     = "game-server"
    image_uri          = var.image_uri
    server_sdk_version = var.server_sdk_version

    port_configuration {
      container_port_ranges {
        from_port = var.game_server_from_port
        to_port   = var.game_server_to_port
        protocol  = var.game_server_port_protocol
      }
    }
  }

  tags = {
    Name = "${var.name_prefix}-group"
  }
}

resource "aws_gamelift_container_fleet" "this" {
  fleet_role_arn = var.fleet_role_arn
  billing_type   = var.billing_type
  description    = var.fleet_description

  game_server_container_group_definition_name = aws_gamelift_container_group_definition.this.arn
  game_server_container_groups_per_instance   = var.game_server_container_groups_per_instance
  instance_type                               = var.instance_type

  dynamic "locations" {
    for_each = var.locations
    content {
      location = locations.value
    }
  }

  instance_connection_port_range {
    from_port = var.instance_connection_from_port
    to_port   = var.instance_connection_to_port
  }

  instance_inbound_permission {
    from_port = var.instance_connection_from_port
    to_port   = var.instance_connection_to_port
    protocol  = var.game_server_port_protocol
    ip_range  = "0.0.0.0/0"
  }

  # Keep TCP open on the connection range because support container definitions
  # may expose TCP ports (for example port 80 in this example).
  instance_inbound_permission {
    from_port = var.instance_connection_from_port
    to_port   = var.instance_connection_to_port
    protocol  = "TCP"
    ip_range  = "0.0.0.0/0"
  }

  tags = {
    Name = "${var.name_prefix}-fleet"
  }
}

output "container_group_definition_id" {
  value = aws_gamelift_container_group_definition.this.id
}

output "container_group_definition_arn" {
  value = aws_gamelift_container_group_definition.this.arn
}

output "container_fleet_id" {
  value = aws_gamelift_container_fleet.this.id
}

output "container_fleet_arn" {
  value = aws_gamelift_container_fleet.this.arn
}
