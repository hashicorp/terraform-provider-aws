resource "aws_s3vectors_vector_bucket" "test" {
{{- template "region" }}
  vector_bucket_name = var.rName
}

resource "aws_s3vectors_index" "test" {
{{- template "region" }}
  index_name         = var.rName
  vector_bucket_name = aws_s3vectors_vector_bucket.test.vector_bucket_name

  data_type       = "float32"
  dimension       = 2
  distance_metric = "euclidean"

{{- template "tags" . }}
}