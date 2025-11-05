resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket = var.rName
}

resource "aws_glue_catalog_database" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_glue_catalog_table" "test" {
{{- template "region" }}
  name          = var.rName
  database_name = var.rName

  storage_descriptor {
    location = "s3://${aws_s3_bucket.test.bucket}"

    columns {
      name = "my_column_1"
      type = "string"
    }

    columns {
      name = "my_column_2"
      type = "string"
    }
  }
}

resource "aws_cleanrooms_configured_table" "test" {
{{- template "region" }}
  name            = "test-name"
  description     = "test description"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = ["my_column_1", "my_column_2"]

  table_reference {
    database_name = var.rName
    table_name    = var.rName
  }
{{- template "tags" }}

  depends_on = [aws_glue_catalog_table.test]
}
