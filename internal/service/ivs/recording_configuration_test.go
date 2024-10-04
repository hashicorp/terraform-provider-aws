// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ivs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfivs "github.com/hashicorp/terraform-provider-aws/internal/service/ivs"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIVSRecordingConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var recordingConfiguration awstypes.RecordingConfiguration
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_recording_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSEndpointID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordingConfigurationConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingConfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.s3.0.bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ivs", regexache.MustCompile(`recording-configuration/.+`)),
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

func TestAccIVSRecordingConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.RecordingConfiguration
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_recording_configuration.test"
	recordingReconnectWindowSeconds := "45"
	recordingMode := "INTERVAL"
	targetIntervalSeconds := "30"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSEndpointID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordingConfigurationConfig_name(bucketName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.s3.0.bucket_name", bucketName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRecordingConfigurationConfig_update(bucketName2, rName2, recordingReconnectWindowSeconds, recordingMode, targetIntervalSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &v2),
					testAccCheckRecordingConfigurationRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.s3.0.bucket_name", bucketName2),
					resource.TestCheckResourceAttr(resourceName, "recording_reconnect_window_seconds", recordingReconnectWindowSeconds),
					resource.TestCheckResourceAttr(resourceName, "thumbnail_configuration.0.recording_mode", recordingMode),
					resource.TestCheckResourceAttr(resourceName, "thumbnail_configuration.0.target_interval_seconds", targetIntervalSeconds),
				),
			},
		},
	})
}

func TestAccIVSRecordingConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var recordingconfiguration awstypes.RecordingConfiguration
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_recording_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSEndpointID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordingConfigurationConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingconfiguration),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfivs.ResourceRecordingConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIVSRecordingConfiguration_disappears_S3Bucket(t *testing.T) {
	ctx := acctest.Context(t)
	var recordingconfiguration awstypes.RecordingConfiguration
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parentResourceName := "aws_s3_bucket.test"
	resourceName := "aws_ivs_recording_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSEndpointID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordingConfigurationConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingconfiguration),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucket(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIVSRecordingConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var recordingConfiguration awstypes.RecordingConfiguration
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_recording_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSEndpointID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordingConfigurationConfig_tags1(bucketName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingConfiguration),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRecordingConfigurationConfig_tags2(bucketName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingConfiguration),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRecordingConfigurationConfig_tags1(bucketName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingConfiguration),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckRecordingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ivs_recording_configuration" {
				continue
			}

			input := &ivs.GetRecordingConfigurationInput{
				Arn: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetRecordingConfiguration(ctx, input)
			if err != nil {
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					return nil
				}
				return err
			}

			return create.Error(names.IVS, create.ErrActionCheckingDestroyed, tfivs.ResNameRecordingConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRecordingConfigurationExists(ctx context.Context, name string, recordingconfiguration *awstypes.RecordingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNameRecordingConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNameRecordingConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSClient(ctx)

		resp, err := tfivs.FindRecordingConfigurationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNameRecordingConfiguration, rs.Primary.ID, err)
		}

		*recordingconfiguration = *resp

		return nil
	}
}

func testAccCheckRecordingConfigurationRecreated(before, after *awstypes.RecordingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Arn), aws.ToString(after.Arn); before == after {
			return fmt.Errorf("Expected Recording Configuration IDs to change, %s", before)
		}

		return nil
	}
}

func testAccRecordingConfigurationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IVSClient(ctx)

	input := &ivs.ListRecordingConfigurationsInput{}
	_, err := conn.ListRecordingConfigurations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRecordingConfigurationConfig_s3Bucket(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, bucketName)
}

func testAccRecordingConfigurationConfig_basic(bucketName string) string {
	return acctest.ConfigCompose(
		testAccRecordingConfigurationConfig_s3Bucket(bucketName),
		`
resource "aws_ivs_recording_configuration" "test" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`)
}

func testAccRecordingConfigurationConfig_name(bucketName, rName string) string {
	return acctest.ConfigCompose(
		testAccRecordingConfigurationConfig_s3Bucket(bucketName),
		fmt.Sprintf(`
resource "aws_ivs_recording_configuration" "test" {
  name = %[1]q
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`, rName))
}

func testAccRecordingConfigurationConfig_update(bucketName, rName, recordingReconnectWindowSeconds, recordingMode, targetIntervalSeconds string) string {
	return acctest.ConfigCompose(
		testAccRecordingConfigurationConfig_s3Bucket(bucketName),
		fmt.Sprintf(`
resource "aws_ivs_recording_configuration" "test" {
  name = %[1]q
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
  recording_reconnect_window_seconds = %[2]s
  thumbnail_configuration {
    recording_mode          = %[3]q
    target_interval_seconds = %[4]s
  }
}
`, rName, recordingReconnectWindowSeconds, recordingMode, targetIntervalSeconds))
}

func testAccRecordingConfigurationConfig_tags1(bucketName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccRecordingConfigurationConfig_s3Bucket(bucketName),
		fmt.Sprintf(`
resource "aws_ivs_recording_configuration" "test" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccRecordingConfigurationConfig_tags2(bucketName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccRecordingConfigurationConfig_s3Bucket(bucketName),
		fmt.Sprintf(`
resource "aws_ivs_recording_configuration" "test" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
