# Specify the provider and access details
provider "aws" {
  region = "us-east-1"
}

resource "aws_servicecatalog_provisioned_product" "test" {
  product_id = "parameter"
}