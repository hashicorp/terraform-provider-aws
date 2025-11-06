# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_codegurureviewer_repository_association" "test" {
  region = var.region

  repository {
    codecommit {
      name = aws_codecommit_repository.test.repository_name
    }
  }
}

# testAccRepositoryAssociation_codecommit_repository

resource "aws_codecommit_repository" "test" {
  region = var.region

  repository_name = var.rName
  description     = "This is a test description"
  lifecycle {
    ignore_changes = [
      tags["codeguru-reviewer"]
    ]
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
