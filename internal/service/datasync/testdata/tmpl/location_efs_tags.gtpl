resource "aws_datasync_location_efs" "test" {
{{- template "region" }}
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test[0].arn
  }
{{- template "tags" . }}
}

# testAccLocationEFSConfig_base

#resource "aws_vpc" "test" {
#  cidr_block = "10.0.0.0/16"
#}
#
#resource "aws_subnet" "test" {
#  cidr_block = "10.0.0.0/24"
#  vpc_id     = aws_vpc.test.id
#}

resource "aws_security_group" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
}

resource "aws_efs_file_system" "test" {
{{- template "region" }}
}

resource "aws_efs_mount_target" "test" {
{{- template "region" }}
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test[0].id
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
