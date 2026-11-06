resource "aws_ec2_application_status_check_association" "test" {
{{- template "region" }}
  application_status_check_id = aws_ec2_application_status_check.test.id
  target_tag_key              = "Name"
  target_tag_value            = var.rName
}

resource "aws_ec2_application_status_check" "test" {
{{- template "region" }}
  protocol = "http"
  port     = 80
}
