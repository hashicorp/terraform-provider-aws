provider "aws" {
  region = "us-east-1"
}

resource "aws_workspaces_ip_group" "example" {
  name        = "main"
  description = "Main IP access control group"

  rules {
    source = "10.10.10.10/16"
  }

  rules {
    source      = "11.11.11.11/16"
    description = "Contractors"
  }
}
