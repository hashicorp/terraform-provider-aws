data "aws_efs_access_point" "test" {
{{- template "region" }}
  access_point_id = aws_efs_access_point.test.id
}