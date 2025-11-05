resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = var.rName
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"
  failover_mode          = "AUTOMATIC"

{{- template "tags" . }}
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}
