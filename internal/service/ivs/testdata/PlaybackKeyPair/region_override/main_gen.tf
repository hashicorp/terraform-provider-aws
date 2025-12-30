# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ivs_playback_key_pair" "test" {
  region = var.region

  public_key = var.rTlsEcdsaPublicKeyPem
}

variable "rTlsEcdsaPublicKeyPem" {
  type     = string
  nullable = false
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
