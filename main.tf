# Specify the provider and access details
provider "aws" {
  region = "us-east-1"
}

resource "aws_servicecatalog_provisioned_product" "test" {
  accept_language             = "en"
  provisioned_product_name    = "fatima"
  product_id                  = "prod-sjyrnwqr6fsa2"
  provisioning_artifact_id    = "pa-ruh7qd5i4czpk"
  path_id                     = "lpv2-ztk6qwhg7da6k"
  provisioning_parameters     = [
    {"ParameterName": "terraform"},
    {"ParameterDescription": "didittwork"}
  ]

}