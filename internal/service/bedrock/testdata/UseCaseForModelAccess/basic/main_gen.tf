# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrock_use_case_for_model_access" "test" {
  form_data = data.aws_bedrock_use_case_for_model_access.test.form_data
}

data "aws_bedrock_use_case_for_model_access" "test" {}

