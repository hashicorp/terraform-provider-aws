# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockdataautomation_library" "test" {
  count = 2

  name = "${var.rName}-${count.index}"
}

resource "aws_bedrockdataautomation_project" "test" {
  name = var.rName

  data_automation_library_configuration {
    library {
      library_arn = aws_bedrockdataautomation_library.test[var.library_index].arn
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "library_index" {
  type     = number
  nullable = false
}
