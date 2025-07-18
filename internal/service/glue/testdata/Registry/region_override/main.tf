provider "aws" {
  region = var.region
}

resource "aws_glue_registry" "test" {
  provider = aws

  registry_name = var.rName
}
