//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.
terraform {
  required_version = ">= 0.15.3"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "3.50.0"
    }
  }
}