# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_quicksight_theme" "test" {
  theme_id = var.rName
  name     = var.rName

  base_theme_id = "MIDNIGHT"

  configuration {
    data_color_palette {
      colors = [
        "#FFFFFF",
        "#111111",
        "#222222",
        "#333333",
        "#444444",
        "#555555",
        "#666666",
        "#777777",
        "#888888",
        "#999999"
      ]
      empty_fill_color = "#FFFFFF"
      min_max_gradient = [
        "#FFFFFF",
        "#111111",
      ]
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
