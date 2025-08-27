resource "aws_ecr_lifecycle_policy" "test" {
{{- template "region" }}
  repository = aws_ecr_repository.test.name

  policy = <<EOF
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Expire images older than 14 days",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 14
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}
EOF
}

resource "aws_ecr_repository" "test" {
{{- template "region" }}
  name = var.rName
}
