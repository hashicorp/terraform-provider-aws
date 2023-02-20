terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = var.aws_region
}

resource "aws_fms_protocol" "foo" {
  name      = var.list_name
  protocols = var.list_protocols
}