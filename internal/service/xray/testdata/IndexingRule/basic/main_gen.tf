# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_indexing_rule" "test" {
  name = "Default"

  rule {
    probabilistic {
      desired_sampling_percentage = 0.66
    }
  }
}

