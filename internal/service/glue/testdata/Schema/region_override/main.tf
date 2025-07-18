provider "aws" {
  region = var.region
}

resource "aws_glue_schema" "test" {
  provider = aws

  schema_name       = var.rName
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"
  data_format      = "AVRO"
  compatibility    = "NONE"
}
