# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_devopsguru_event_sources_config" "test" {
  event_sources {
    amazon_code_guru_profiler {
      status = "ENABLED"
    }
  }
}

