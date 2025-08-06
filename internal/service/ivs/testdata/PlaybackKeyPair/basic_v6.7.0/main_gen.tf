# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ivs_playback_key_pair" "test" {
  public_key = var.rTlsEcdsaPublicKeyPem
}

variable "rTlsEcdsaPublicKeyPem" {
  type     = string
  nullable = false
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.7.0"
    }
  }
}

provider "aws" {}
