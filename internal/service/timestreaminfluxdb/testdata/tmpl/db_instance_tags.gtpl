resource "aws_timestreaminfluxdb_db_instance" "test" {
  name                   = var.rName
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"

{{- template "tags" . }}
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}
