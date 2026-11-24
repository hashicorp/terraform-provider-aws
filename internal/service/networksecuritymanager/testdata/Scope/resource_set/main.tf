# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = var.rName
  }
}

resource "aws_networksecuritymanager_scope" "test" {
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::EC2::EIP"
      include_all   = var.selection == "include_all" ? true : null

      dynamic "include" {
        for_each = var.selection == "include" ? [1] : []

        content {
          explicit_arns = [aws_eip.test.arn]
        }
      }

      dynamic "exclude" {
        for_each = var.selection == "exclude" ? [1] : []

        content {
          explicit_arns = [aws_eip.test.arn]

          expression {
            criteria {
              tags = {
                Name = var.rName
              }
            }
          }
        }
      }
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "selection" {
  description = "Which of include_all, include and exclude the resource scope sets"
  type        = string
  nullable    = false
}
