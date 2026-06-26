# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

action "aws_bedrockagentcore_submit_registry_record_for_approval" "test" {
  config {
    registry_id = aws_bedrockagentcore_registry_record.test.registry_id
    record_id   = aws_bedrockagentcore_registry_record.test.record_id
  }
}

resource "aws_bedrockagentcore_registry_record" "test" {
  name        = "${var.rName}-record"
  registry_id = aws_bedrockagentcore_registry.test.registry_id

  descriptor_type = "CUSTOM"

  descriptors {
    custom {
      inline_content = "{}"
    }
  }

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.aws_bedrockagentcore_submit_registry_record_for_approval.test]
    }
  }
}

resource "aws_bedrockagentcore_registry" "test" {
  name = "${var.rName}-registry"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
