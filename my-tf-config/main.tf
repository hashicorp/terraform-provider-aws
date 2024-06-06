provider "aws" {
  region = "us-east-1"
}


data "aws_iam_server_certificates" "my-domain" {
  latest      = true
}