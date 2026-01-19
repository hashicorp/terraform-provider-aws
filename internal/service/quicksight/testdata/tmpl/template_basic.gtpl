resource "aws_quicksight_template" "test" {
  template_id         = var.rName
  name                = var.rName
  version_description = "test"
  definition {
    data_set_configuration {
      data_set_schema {
        column_schema_list {
          name      = "Column1"
          data_type = "STRING"
        }
        column_schema_list {
          name      = "Column2"
          data_type = "INTEGER"
        }
      }
      placeholder = "1"
    }
    sheets {
      title    = "Test"
      sheet_id = "Test1"
      visuals {
        bar_chart_visual {
          visual_id = "BarChart"
          chart_configuration {
            field_wells {
              bar_chart_aggregated_field_wells {
                category {
                  categorical_dimension_field {
                    field_id = "1"
                    column {
                      column_name         = "Column1"
                      data_set_identifier = "1"
                    }
                  }
                }
                values {
                  numerical_measure_field {
                    field_id = "2"
                    column {
                      column_name         = "Column2"
                      data_set_identifier = "1"
                    }
                    aggregation_function {
                      simple_numerical_aggregation = "SUM"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
{{- template "tags" . }}
}

# testAccTemplateConfig_base

resource "aws_quicksight_data_set" "test" {
  data_set_id = var.rName
  name        = var.rName
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = var.rName
    s3_source {
      data_source_arn = aws_quicksight_data_source.test.arn
      input_columns {
        name = "Column1"
        type = "STRING"
      }
      input_columns {
        name = "Column2"
        type = "STRING"
      }
      upload_settings {}
    }
  }
  logical_table_map {
    logical_table_map_id = var.rName
    alias                = "Group1"
    source {
      physical_table_id = var.rName
    }
    data_transforms {
      cast_column_type_operation {
        column_name     = "Column2"
        new_column_type = "INTEGER"
      }
    }
  }

  lifecycle {
    ignore_changes = [
      physical_table_map
    ]
  }
}

# testAccDataSetConfig_base

resource "aws_quicksight_data_source" "test" {
  data_source_id = var.rName
  name           = var.rName

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  type = "S3"
}

# testAccDataSourceConfig_base

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_object" "test_data" {
  depends_on = [aws_s3_bucket_acl.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = "${var.rName}-test-data"
  content = <<EOF
[
	{
		"Column1": "aaa",
		"Column2": 1
	},
	{
		"Column1": "bbb",
		"Column2": 1
	}
]
EOF
  acl     = "public-read"
}

resource "aws_s3_object" "test" {
  depends_on = [aws_s3_bucket_acl.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = var.rName
  content = <<EOF
{
  "fileLocations": [
      {
          "URIs": [
              "https://${aws_s3_bucket.test.bucket}.s3.${data.aws_partition.current.dns_suffix}/${var.rName}-test-data"
          ]
      }
  ],
  "globalUploadSettings": {
      "format": "JSON"
  }
}
EOF
  acl     = "public-read"
}
