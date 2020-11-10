resource "aws_cloudwatch_log_group" "test" {
  name = "%q"

  lifecycle {
    create_before_destroy = true
  }
}