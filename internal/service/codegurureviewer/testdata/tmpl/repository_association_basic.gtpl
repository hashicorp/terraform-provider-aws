resource "aws_codegurureviewer_repository_association" "test" {
{{- template "region" }}
  repository {
    codecommit {
      name = aws_codecommit_repository.test.repository_name
    }
  }
{{- template "tags" . }}
}

# testAccRepositoryAssociation_codecommit_repository

resource "aws_codecommit_repository" "test" {
{{- template "region" }}
  repository_name = var.rName
  description     = "This is a test description"
  lifecycle {
    ignore_changes = [
      tags["codeguru-reviewer"]
    ]
  }
}
