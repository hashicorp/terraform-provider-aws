data "aws_pinpointsmsvoicev2_pool" "test" {
{{- template "region" }}
  id = aws_pinpointsmsvoicev2_pool.test.id
}
