resource "aws_continuum_integration" "test" {
{{- template "region" }}
  name = var.rName

  provider_configuration {
    s3 {
      bucket      = aws_s3_bucket.test.id
      role_arn    = aws_iam_role.test.arn
      external_id = var.rName
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.code]
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "code" {
{{- template "region" }}
  bucket     = aws_s3_bucket.test.id
  key        = "code/app.zip"
  content    = "placeholder"
  depends_on = [aws_s3_bucket_versioning.test]
}

data "aws_iam_policy_document" "trust" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["continuum.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "sts:ExternalId"
      values   = [var.rName]
    }
  }
}

data "aws_iam_policy_document" "access" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.test.arn]
  }

  statement {
    actions   = ["s3:GetObject", "s3:GetObjectVersion"]
    resources = ["${aws_s3_bucket.test.arn}/code/*", "${aws_s3_bucket.test.arn}/context/*"]
  }

  statement {
    actions   = ["s3:PutObject", "s3:GetObject", "s3:GetObjectVersion"]
    resources = ["${aws_s3_bucket.test.arn}/continuum-patches/*"]
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  path               = "/continuum/integrations/"
  assume_role_policy = data.aws_iam_policy_document.trust.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.access.json
}
