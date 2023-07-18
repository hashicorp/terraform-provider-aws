// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfivs "github.com/hashicorp/terraform-provider-aws/internal/service/ivs"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIVSRecordingConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var recordingConfiguration ivs.RecordingConfiguration
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_recording_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordingConfigurationConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingConfiguration),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.s3.0.bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ivs", regexp.MustCompile(`recording-configuration/.+`)),
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
	var v1, v2 ivs.RecordingConfiguration
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
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordingConfigurationConfig_name(bucketName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
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
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
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
	var recordingconfiguration ivs.RecordingConfiguration
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_recording_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
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
	var recordingconfiguration ivs.RecordingConfiguration
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parentResourceName := "aws_s3_bucket.test"
	resourceName := "aws_ivs_recording_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
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
	var recordingConfiguration ivs.RecordingConfiguration
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_recording_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccRecordingConfigurationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordingConfigurationConfig_tags1(bucketName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingConfiguration),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRecordingConfigurationConfig_tags2(bucketName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingConfiguration),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRecordingConfigurationConfig_tags1(bucketName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordingConfigurationExists(ctx, resourceName, &recordingConfiguration),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckRecordingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ivs_recording_configuration" {
				continue
			}

			input := &ivs.GetRecordingConfigurationInput{
				Arn: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetRecordingConfigurationWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			return create.Error(names.IVS, create.ErrActionCheckingDestroyed, tfivs.ResNameRecordingConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRecordingConfigurationExists(ctx context.Context, name string, recordingconfiguration *ivs.RecordingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNameRecordingConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNameRecordingConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSConn(ctx)

		resp, err := tfivs.FindRecordingConfigurationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNameRecordingConfiguration, rs.Primary.ID, err)
		}

		*recordingconfiguration = *resp

		return nil
	}
}

func testAccCheckRecordingConfigurationRecreated(before, after *ivs.RecordingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.Arn), aws.StringValue(after.Arn); before == after {
			return fmt.Errorf("Expected Recording Configuration IDs to change, %s", before)
		}

		return nil
	}
}

func testAccRecordingConfigurationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IVSConn(ctx)

	input := &ivs.ListRecordingConfigurationsInput{}
	_, err := conn.ListRecordingConfigurationsWithContext(ctx, input)

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
