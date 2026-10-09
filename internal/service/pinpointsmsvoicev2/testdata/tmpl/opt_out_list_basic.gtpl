resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
