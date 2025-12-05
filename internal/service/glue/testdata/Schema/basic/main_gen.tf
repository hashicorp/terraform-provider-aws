# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_glue_schema" "test" {
  schema_name       = var.rName
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"
  data_format       = "AVRO"
  compatibility     = "NONE"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
