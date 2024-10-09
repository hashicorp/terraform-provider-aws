data "aws_db_instance" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
}
