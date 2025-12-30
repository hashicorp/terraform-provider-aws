# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ivs_playback_key_pair" "test" {
  public_key = var.rTlsEcdsaPublicKeyPem
}

variable "rTlsEcdsaPublicKeyPem" {
  type     = string
  nullable = false
}

