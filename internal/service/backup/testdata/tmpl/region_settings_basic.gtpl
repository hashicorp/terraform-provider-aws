resource "aws_backup_region_settings" "test" {
{{- template "region" }}
  resource_type_opt_in_preference = {
    "Aurora"                 = true
    "CloudFormation"         = true
    "DocumentDB"             = true
    "DSQL"                   = true
    "DynamoDB"               = true
    "EBS"                    = true
    "EC2"                    = true
    "EFS"                    = true
    "FSx"                    = true
    "Neptune"                = true
    "RDS"                    = true
    "Redshift"               = true
    "Redshift Serverless"    = true
    "S3"                     = true
    "SAP HANA on Amazon EC2" = true
    "Storage Gateway"        = true
    "Timestream"             = true
    "VirtualMachine"         = true
  }
}
