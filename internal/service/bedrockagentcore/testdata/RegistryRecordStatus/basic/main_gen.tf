# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_registry_record_status" "test" {
  registry_id = aws_bedrockagentcore_registry_record.test.registry_id
  record_id   = aws_bedrockagentcore_registry_record.test.record_id

  status        = "APPROVED"
  status_reason = "LGTM"

  # Ensure that the registry record is in PENDING_APPROVAL state.
  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = [action.aws_bedrockagentcore_submit_registry_record_for_approval.test]
    }
  }
}

action "aws_bedrockagentcore_submit_registry_record_for_approval" "test" {
  config {
    # Works around issue with "region" template and terrafmt.
    region = data.aws_region.current.region

    registry_id = aws_bedrockagentcore_registry_record.test.registry_id
    record_id   = aws_bedrockagentcore_registry_record.test.record_id
  }
}

data "aws_region" "current" {
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
}

resource "aws_bedrockagentcore_registry" "test" {
  name = "${var.rName}-registry"
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
