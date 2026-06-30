resource "aws_bedrock_use_case_for_model_access" "test" {
  form_data = data.aws_bedrock_use_case_for_model_access.test.form_data
}

data "aws_bedrock_use_case_for_model_access" "test" {}
