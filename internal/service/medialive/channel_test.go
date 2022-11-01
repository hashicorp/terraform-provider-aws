package medialive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmedialive "github.com/hashicorp/terraform-provider-aws/internal/service/medialive"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaLiveChannel_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccChannelsPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMediaLiveChannel_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccChannelsPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(resourceName, &channel),
					acctest.CheckResourceDisappears(acctest.Provider, tfmedialive.ResourceChannel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckChannelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_medialive_channel" {
			continue
		}

		_, err := tfmedialive.FindChannelByID(ctx, conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingDestroyed, tfmedialive.ResNameChannel, rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckChannelExists(name string, channel *medialive.DescribeChannelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameChannel, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameChannel, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
		ctx := context.Background()
		resp, err := tfmedialive.FindChannelByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameChannel, rs.Primary.ID, err)
		}

		*channel = *resp

		return nil
	}
}

func testAccChannelsPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
	ctx := context.Background()

	input := &medialive.ListChannelsInput{}
	_, err := conn.ListChannels(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccChannelBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
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
`, rName)
}

func testAccChannelConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test1" {
  bucket = "%[1]s-1"
}

resource "aws_s3_bucket_acl" "test1" {
  bucket = aws_s3_bucket.test1.id
  acl    = "private"
}

resource "aws_s3_bucket" "test2" {
  bucket = "%[1]s-2"
}

resource "aws_s3_bucket_acl" "test2" {
  bucket = aws_s3_bucket.test2.id
  acl    = "private"
}

resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = "10.0.0.8/32"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_medialive_input" "test" {
  name                  = %[1]q
  input_security_groups = [aws_medialive_input_security_group.test.id]
  type                  = "UDP_PUSH"

  tags = {
    Name = %[1]q
  }
}

resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachment {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id

    input_setting {
      audio_selector {
        name = %[1]q
      }
    }
  }

  destination {
    id = %[1]q

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

    audio_description {
      audio_selector_name = %[1]q
      name                = %[1]q
    }

    video_description {
      name = "test-video-name"
    }

    output_group {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      output {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
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
}
`, rName))
}
