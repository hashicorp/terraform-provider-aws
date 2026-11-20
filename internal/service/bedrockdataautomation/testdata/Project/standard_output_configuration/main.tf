# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockdataautomation_project" "test" {
  name = var.rName

  standard_output_configuration {
    document {
      extraction {
        bounding_box {
          state = "ENABLED"
        }

        granularity {
          types = var.granularity_types
        }
      }

      generative_field {
        state = var.generative_field_state
      }
    }

    image {
      extraction {
        bounding_box {
          state = "DISABLED"
        }

        category {
          state = "ENABLED"
          types = ["TEXT_DETECTION", "LOGOS"]
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

variable "granularity_types" {
  type     = list(string)
  nullable = false
}

variable "generative_field_state" {
  type     = string
  nullable = false
}
