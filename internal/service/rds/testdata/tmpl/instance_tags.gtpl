resource "aws_db_instance" "test" {
  identifier          = var.rName
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"

{{- template "tags" . }}
}

# testAccInstanceConfig_orderableClassMySQL

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.t4g.micro"]
}
