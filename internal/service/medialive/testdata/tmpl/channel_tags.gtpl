resource "aws_medialive_channel" "test" {
  name          = var.rName
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = var.rName

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = var.rName
      name                = var.rName
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = var.rName
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [var.rName]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
{{- template "tags" . }}
}

# testAccChannelConfig_base

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "medialive.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:*",
          "s3:*",
          "mediastore:*",
          "mediaconnect:*",
          "cloudwatch:*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

# testAccChannelConfig_baseS3

resource "aws_s3_bucket" "test1" {
  bucket = "${var.rName}-1"
}

resource "aws_s3_bucket" "test2" {
  bucket = "${var.rName}-2"
}

# testAccChannelConfig_baseMultiplex

resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = "10.0.0.8/32"
  }
}

resource "aws_medialive_input" "test" {
  name                  = var.rName
  input_security_groups = [aws_medialive_input_security_group.test.id]
  type                  = "UDP_PUSH"
}
