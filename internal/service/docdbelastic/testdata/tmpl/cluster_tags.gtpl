resource "aws_docdbelastic_cluster" "test" {
{{- template "region" }}
  name           = var.rName
  shard_capacity = 2
  shard_count    = 1

  admin_user_name     = "testuser"
  admin_user_password = "testpassword"
  auth_type           = "PLAIN_TEXT"

  preferred_maintenance_window = "Tue:04:00-Tue:04:30"

  vpc_security_group_ids = [
    aws_security_group.test.id
  ]

  subnet_ids = aws_subnet.test[*].id
{{- template "tags" . }}
}

# testAccClusterBaseConfig

resource "aws_security_group" "test" {
{{- template "region" }}
  name   = var.rName
  vpc_id = aws_vpc.test.id
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}