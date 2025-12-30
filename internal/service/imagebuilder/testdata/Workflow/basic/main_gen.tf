# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_imagebuilder_workflow" "test" {
  name    = var.rName
  version = "1.0.0"
  type    = "TEST"

  data = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
