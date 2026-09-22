resource "aws_efs_mount_target" "test" {
{{- template "region" }}
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test[0].id
}

resource "aws_efs_file_system" "test" {
{{- template "region" }}
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
