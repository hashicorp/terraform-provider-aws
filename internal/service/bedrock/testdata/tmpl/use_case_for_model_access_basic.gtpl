# Use import-on-create as this resource type is a global singleton and cannot be deleted
resource "aws_bedrock_use_case_for_model_access" "test" {
  form_data = data.aws_bedrock_use_case_for_model_access.test.form_data
}

data "aws_bedrock_use_case_for_model_access" "test" {}
