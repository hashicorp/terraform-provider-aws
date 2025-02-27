// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdataexchange "github.com/hashicorp/terraform-provider-aws/internal/service/dataexchange"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const testAccDataSetIDEnvVar = "TF_AWS_DATAEXCHANGE_DATA_SET_ID"

func testAccPreCheckDataSetID(t *testing.T) {
	if dataSetId := os.Getenv(testAccDataSetIDEnvVar); dataSetId == "" {
		t.Skipf("Environment variable %s is required for DataExchange Event Action tests. "+
			"This requires subscribing to an AWS Data Exchange product (e.g. AWS Data Exchange Heartbeat) "+
			"and setting the environment variable to the entitled dataset ID", testAccDataSetIDEnvVar)
	}
}

// TestAccDataExchangeEventAction_basic, TestAccDataExchangeEventAction_update, TestAccDataExchangeEventAction_disappears, TestAccDataExchangeEventAction_keyPattern, TestAccDataExchangeEventAction_encryption, TestAccDataExchangeEventAction_kmsKeyEncryption require an entitled AWS Data Exchange dataset.
// To run this test:
// 1. Subscribe to an AWS Data Exchange product (e.g. AWS Data Exchange Heartbeat)
// 2. Set TF_AWS_DATAEXCHANGE_DATA_SET_ID environment variable to the entitled dataset ID
// 3. Ensure your AWS region is set to where the entitled dataset resides (e.g. us-east-1)

func TestAccDataExchangeEventAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	input := sts.GetCallerIdentityInput{}
	identity, err := stsConn.GetCallerIdentity(ctx, &input)
	if err != nil {
		t.Error(err)
	}

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDataSetID(t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
			testAccEventActionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_basic(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.revision_destination.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "event_revision_published.data_set_id", dataSetId),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`event-actions/.+`)),
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

func TestAccDataExchangeEventAction_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	input := sts.GetCallerIdentityInput{}
	identity, err := stsConn.GetCallerIdentity(ctx, &input)
	if err != nil {
		t.Error(err)
	}

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDataSetID(t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
			testAccEventActionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_basic(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.revision_destination.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "event_revision_published.data_set_id", dataSetId),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`event-actions/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventActionConfig_encryption(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.encryption.type", "AES256"),
				),
			},
		},
	})
}

func TestAccDataExchangeEventAction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	input := sts.GetCallerIdentityInput{}
	identity, err := stsConn.GetCallerIdentity(ctx, &input)
	if err != nil {
		t.Error(err)
	}

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDataSetID(t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
			testAccEventActionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_basic(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdataexchange.ResourceEventAction, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataExchangeEventAction_keyPattern(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	input := sts.GetCallerIdentityInput{}
	identity, err := stsConn.GetCallerIdentity(ctx, &input)
	if err != nil {
		t.Error(err)
	}

	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDataSetID(t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
			testAccEventActionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_keyPattern(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &dataexchange.GetEventActionOutput{}),
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.revision_destination.key_pattern", "${Revision.CreatedAt}/${Asset.Name}"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`event-actions/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeEventAction_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	input := sts.GetCallerIdentityInput{}
	identity, err := stsConn.GetCallerIdentity(ctx, &input)
	if err != nil {
		t.Error(err)
	}

	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDataSetID(t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
			testAccEventActionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_encryption(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &dataexchange.GetEventActionOutput{}),
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.encryption.type", "AES256"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`event-actions/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeEventAction_kmsKeyEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	input := sts.GetCallerIdentityInput{}
	identity, err := stsConn.GetCallerIdentity(ctx, &input)
	if err != nil {
		t.Error(err)
	}

	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDataSetID(t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
			testAccEventActionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_kmsKeyEncryption(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &dataexchange.GetEventActionOutput{}),
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.encryption.type", "aws:kms"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`event-actions/.+`)),
				),
			},
		},
	})
}

func testAccEventActionPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)

	input := &dataexchange.ListEventActionsInput{}

	_, err := conn.ListEventActions(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckEventActionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dataexchange_event_action" {
				continue
			}

			_, err := tfdataexchange.FindEventActionByID(ctx, conn, rs.Primary.ID)

			if errs.IsA[*retry.NotFoundError](err) {
				return nil
			}

			if err != nil {
				return create.Error(names.DataExchange, create.ErrActionCheckingDestroyed, tfdataexchange.ResNameEventAction, rs.Primary.ID, err)
			}

			return create.Error(names.DataExchange, create.ErrActionCheckingDestroyed, tfdataexchange.ResNameEventAction, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckEventActionExists(ctx context.Context, n string, v *dataexchange.GetEventActionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)
		output, err := tfdataexchange.FindEventActionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if output == nil {
			return fmt.Errorf("DataExchange EventAction not found")
		}

		*v = *output

		return nil
	}
}

func s3BucketConfig(bucketName, accountId string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    principals {
      type        = "Service"
      identifiers = ["dataexchange.amazonaws.com"]
    }

    actions = [
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]

    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"

      values = [
        "%s"
      ]
    }
  }
}
`, bucketName, accountId)
}

func testAccEventActionConfig_basic(bucketName, dataSetId, accountId string) string {
	return acctest.ConfigCompose(
		s3BucketConfig(bucketName, accountId),
		fmt.Sprintf(`
resource "aws_dataexchange_event_action" "test" {
  action_export_revision_to_s3 {
    revision_destination {
      bucket = aws_s3_bucket.test.bucket
    }
  }

  event_revision_published {
    data_set_id = "%s"
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, dataSetId))
}

func testAccEventActionConfig_keyPattern(bucketName, dataSetId, accountId string) string {
	return acctest.ConfigCompose(
		s3BucketConfig(bucketName, accountId),
		fmt.Sprintf(`
resource "aws_dataexchange_event_action" "test" {
  action_export_revision_to_s3 {
    encryption {
      type = "AES256"
    }
    revision_destination {
      bucket      = aws_s3_bucket.test.bucket
      key_pattern = "$${Revision.CreatedAt}/$${Asset.Name}"
    }
  }

  event_revision_published {
    data_set_id = "%s"
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, dataSetId),
	)
}

func testAccEventActionConfig_encryption(bucketName, dataSetId, accountId string) string {
	return acctest.ConfigCompose(
		s3BucketConfig(bucketName, accountId),
		fmt.Sprintf(`
resource "aws_dataexchange_event_action" "test" {
  action_export_revision_to_s3 {
    encryption {
      type = "AES256"
    }
    revision_destination {
      bucket      = aws_s3_bucket.test.bucket
      key_pattern = "$${Revision.CreatedAt}/$${Asset.Name}"
    }
  }

  event_revision_published {
    data_set_id = "%s"
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, dataSetId),
	)
}

func testAccEventActionConfig_kmsKeyEncryption(bucketName, dataSetId, accountId string) string {
	return acctest.ConfigCompose(
		s3BucketConfig(bucketName, accountId),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
}

resource "aws_dataexchange_event_action" "test" {
  action_export_revision_to_s3 {
    encryption {
      type        = "aws:kms"
      kms_key_arn = aws_kms_key.test.arn
    }
    revision_destination {
      bucket      = aws_s3_bucket.test.bucket
      key_pattern = "$${Revision.CreatedAt}/$${Asset.Name}"
    }
  }

  event_revision_published {
    data_set_id = "%s"
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, dataSetId),
	)
}
