data "aws_pinpointsmsvoicev2_opt_out_list" "test" {
{{- template "region" }}
  name = aws_pinpointsmsvoicev2_opt_out_list.test.name
}
