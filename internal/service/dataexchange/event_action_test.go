// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfdataexchange "github.com/hashicorp/terraform-provider-aws/internal/service/dataexchange"
)

func TestAccDataExchangeEventAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"

	bucketName := strconv.Itoa(int(time.Now().UnixNano()))

	if _, okAcc := os.LookupEnv("TF_ACC"); !okAcc {
		t.Skipf("TF_ACC must be set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	identity, err := stsConn.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})

	dataSetId, err := helperAccEventActionGetReceivedDataSet(ctx, awstypes.AssetTypeS3Snapshot)
	if err != nil {
		t.Error(err)
	}

	if dataSetId == "" {
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
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
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "event_revision_published.data_set_id", dataSetId),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn", "dataexchange", regexache.MustCompile(`event-actions/.+`)),
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

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"

	bucketName := strconv.Itoa(int(time.Now().UnixNano()))

	if _, okAcc := os.LookupEnv("TF_ACC"); !okAcc {
		t.Skipf("TF_ACC must be set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	identity, err := stsConn.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})

	dataSetId, err := helperAccEventActionGetReceivedDataSet(ctx, awstypes.AssetTypeS3Snapshot)
	if err != nil {
		t.Error(err)
	}

	if dataSetId == "" {
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
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
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "event_revision_published.data_set_id", dataSetId),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn", "dataexchange", regexache.MustCompile(`event-actions/.+`)),
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
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.s3_encryption_type", "AES256"),
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

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"

	bucketName := strconv.Itoa(int(time.Now().UnixNano()))

	if _, okAcc := os.LookupEnv("TF_ACC"); !okAcc {
		t.Skipf("TF_ACC must be set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	identity, err := stsConn.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})

	dataSetId, err := helperAccEventActionGetReceivedDataSet(ctx, awstypes.AssetTypeS3Snapshot)
	if err != nil {
		t.Error(err)
	}

	if dataSetId == "" {
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
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
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	if _, okAcc := os.LookupEnv("TF_ACC"); !okAcc {
		t.Skipf("TF_ACC must be set")
	}

	dataSetId, err := helperAccEventActionGetReceivedDataSet(ctx, awstypes.AssetTypeS3Snapshot)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	identity, err := stsConn.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		t.Error(err)
	}

	if dataSetId == "" {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_keyPattern(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &dataexchange.GetEventActionOutput{}),
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.key_pattern", "${Revision.CreatedAt}/${Asset.Name}"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`event-actions/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeEventAction_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	if _, okAcc := os.LookupEnv("TF_ACC"); !okAcc {
		t.Skipf("TF_ACC must be set")
	}

	dataSetId, err := helperAccEventActionGetReceivedDataSet(ctx, awstypes.AssetTypeS3Snapshot)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	identity, err := stsConn.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		t.Error(err)
	}

	if dataSetId == "" {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_encryption(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &dataexchange.GetEventActionOutput{}),
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.s3_encryption_type", "AES256"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`event-actions/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeEventAction_kmsKeyEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	if _, okAcc := os.LookupEnv("TF_ACC"); !okAcc {
		t.Skipf("TF_ACC must be set")
	}

	dataSetId, err := helperAccEventActionGetReceivedDataSet(ctx, awstypes.AssetTypeS3Snapshot)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Error(err)
	}
	stsConn := sts.NewFromConfig(cfg)
	identity, err := stsConn.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		t.Error(err)
	}

	if dataSetId == "" {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_kmsKeyEncryption(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &dataexchange.GetEventActionOutput{}),
					resource.TestCheckResourceAttr(resourceName, "action_export_revision_to_s3.s3_encryption_type", "aws:kms"),
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

			_, err := conn.GetEventAction(ctx, &dataexchange.GetEventActionInput{
				EventActionId: aws.String(rs.Primary.ID),
			})

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
		output, err := conn.GetEventAction(ctx, &dataexchange.GetEventActionInput{
			EventActionId: aws.String(rs.Primary.ID),
		})
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

func testAccEventActionConfig_basic(bucketName, dataSetId, accountId string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
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

resource "aws_dataexchange_event_action" "test" {
  action_export_revision_to_s3  {
    bucket = aws_s3_bucket.test.id
  }

  event_revision_published {
    data_set_id = "%s"
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, bucketName, accountId, dataSetId)
}

func testAccEventActionConfig_keyPattern(bucketName, dataSetId, accountId string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
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

resource "aws_dataexchange_event_action" "test" {
  action_export_revision_to_s3  {
    s3_encryption_type = "AES256"
    bucket = aws_s3_bucket.test.id
    key_pattern = "$${Revision.CreatedAt}/$${Asset.Name}"
  }

  event_revision_published  {
    data_set_id = "%s"
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, bucketName, accountId, dataSetId)
}

func testAccEventActionConfig_encryption(bucketName, dataSetId, accountId string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
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

resource "aws_dataexchange_event_action" "test" {
  action_export_revision_to_s3  {
    s3_encryption_type = "AES256"
    bucket = aws_s3_bucket.test.id
    key_pattern = "$${Revision.CreatedAt}/$${Asset.Name}"
  }

  event_revision_published  {
    data_set_id = "%s"
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, bucketName, accountId, dataSetId)
}

func testAccEventActionConfig_kmsKeyEncryption(bucketName, dataSetId, accountId string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
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

resource "aws_kms_key" "test" {
}

resource "aws_dataexchange_event_action" "test" {
  action_export_revision_to_s3  {
    s3_encryption_type = "aws:kms"
    s3_encryption_kms_key_arn = aws_kms_key.test.arn
    bucket = aws_s3_bucket.test.id
    key_pattern = "$${Revision.CreatedAt}/$${Asset.Name}"
  }

  event_revision_published  {
    data_set_id = "%s"
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, bucketName, accountId, dataSetId)
}

func helperAccEventActionGetReceivedDataSet(ctx context.Context, assetType awstypes.AssetType) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}

	conn := dataexchange.NewFromConfig(cfg)
	out, err := conn.ListDataSets(ctx, &dataexchange.ListDataSetsInput{
		MaxResults: 200,
		Origin:     aws.String("ENTITLED"),
	})
	if err != nil {
		return "", err
	}
	for _, dataSet := range out.DataSets {
		if dataSet.AssetType == assetType {
			existingActions, err := conn.ListEventActions(ctx, &dataexchange.ListEventActionsInput{
				EventSourceId: dataSet.SourceId,
			})
			if err != nil {
				continue
			}

			if len(existingActions.EventActions) < 5 {
				return *dataSet.Id, nil
			}
		}
	}

	return "", nil
}

func TestHelperAccEventActionGetReceivedDataSet(t *testing.T) {
	ctx := context.Background()
	if _, okAcc := os.LookupEnv("TF_ACC"); !okAcc {
		t.Skipf("TF_ACC must be set")
	}

	_, err := helperAccEventActionGetReceivedDataSet(ctx, awstypes.AssetTypeS3Snapshot)
	if err != nil {
		t.Error(err)
	}
}
