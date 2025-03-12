// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
			"and setting the environment variable to the entitled data set ID", testAccDataSetIDEnvVar)
	}
}

// TestAccDataExchangeEventAction_basic, TestAccDataExchangeEventAction_update, TestAccDataExchangeEventAction_disappears, TestAccDataExchangeEventAction_keyPattern, TestAccDataExchangeEventAction_encryption, TestAccDataExchangeEventAction_kmsKeyEncryption require an entitled AWS Data Exchange dataset.
// To run this test:
// 1. Subscribe to an AWS Data Exchange product (e.g. AWS Data Exchange Heartbeat)
// 2. Set TF_AWS_DATAEXCHANGE_DATA_SET_ID environment variable to the entitled data set ID
// 3. Ensure your AWS region is set to where the entitled data set resides (e.g. us-east-1)

func TestAccDataExchangeEventAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.ParallelTest(t, resource.TestCase{
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
				Config: testAccEventActionConfig_basic(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.revision_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.revision_destination.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.revision_destination.0.key_pattern", "${Revision.CreatedAt}/${Asset.Name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dataexchange", "event-actions/{id}"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, "event.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event.0.revision_published.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event.0.revision_published.0.data_set_id", dataSetId),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	createdAtNoChange := statecheck.CompareValue(compare.ValuesSame())
	updatedAtChange := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
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
				Config: testAccEventActionConfig_basic(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventActionConfig_encryption_AES256(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.0.kms_key_arn"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.0.type", "AES256"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					createdAtNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrCreatedAt)),
					updatedAtChange.AddStateValue(resourceName, tfjsonpath.New("updated_at")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.ParallelTest(t, resource.TestCase{
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
				Config: testAccEventActionConfig_basic(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
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

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.ParallelTest(t, resource.TestCase{
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
				Config: testAccEventActionConfig_keyPattern(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.revision_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.revision_destination.0.key_pattern", "${Asset.Name}/${Revision.CreatedAt.Year}/${Revision.CreatedAt.Month}/${Revision.CreatedAt.Day}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventActionConfig_basic(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.revision_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.revision_destination.0.key_pattern", "${Revision.CreatedAt}/${Asset.Name}"),
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

func TestAccDataExchangeEventAction_encryption_AES256(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.ParallelTest(t, resource.TestCase{
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
				Config: testAccEventActionConfig_encryption_AES256(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.0.kms_key_arn"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.0.type", "AES256"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventActionConfig_basic(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.#", "0"),
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

func TestAccDataExchangeEventAction_encryption_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var eventaction dataexchange.GetEventActionOutput
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSetId := os.Getenv(testAccDataSetIDEnvVar)

	resource.ParallelTest(t, resource.TestCase{
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
				Config: testAccEventActionConfig_encryption_kmsKey(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.export_revision_to_s3.0.encryption.0.kms_key_arn", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.0.type", "aws:kms"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventActionConfig_basic(bucketName, dataSetId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &eventaction),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.export_revision_to_s3.0.encryption.#", "0"),
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

func testAccEventActionPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)

	input := dataexchange.ListEventActionsInput{}
	_, err := conn.ListEventActions(ctx, &input)

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

func s3BucketConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
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
        data.aws_caller_identity.current.account_id
      ]
    }
  }
}

data "aws_caller_identity" "current" {}
`, bucketName)
}

func testAccEventActionConfig_basic(bucketName, dataSetId string) string {
	return acctest.ConfigCompose(
		s3BucketConfig(bucketName),
		fmt.Sprintf(`
resource "aws_dataexchange_event_action" "test" {
  action {
    export_revision_to_s3 {
      revision_destination {
        bucket = aws_s3_bucket.test.bucket
      }
    }
  }

  event {
    revision_published {
      data_set_id = %[1]q
    }
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, dataSetId))
}

func testAccEventActionConfig_keyPattern(bucketName, dataSetId string) string {
	return acctest.ConfigCompose(
		s3BucketConfig(bucketName),
		fmt.Sprintf(`
resource "aws_dataexchange_event_action" "test" {
  action {
    export_revision_to_s3 {
      encryption {
        type = "AES256"
      }
      revision_destination {
        bucket      = aws_s3_bucket.test.bucket
        key_pattern = "$${Asset.Name}/$${Revision.CreatedAt.Year}/$${Revision.CreatedAt.Month}/$${Revision.CreatedAt.Day}"
      }
    }
  }

  event {
    revision_published {
      data_set_id = %[1]q
    }
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, dataSetId),
	)
}

func testAccEventActionConfig_encryption_AES256(bucketName, dataSetId string) string {
	return acctest.ConfigCompose(
		s3BucketConfig(bucketName),
		fmt.Sprintf(`
resource "aws_dataexchange_event_action" "test" {
  action {
    export_revision_to_s3 {
      encryption {
        type = "AES256"
      }
      revision_destination {
        bucket = aws_s3_bucket.test.bucket
      }
    }
  }

  event {
    revision_published {
      data_set_id = %[1]q
    }
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, dataSetId),
	)
}

func testAccEventActionConfig_encryption_kmsKey(bucketName, dataSetId string) string {
	return acctest.ConfigCompose(
		s3BucketConfig(bucketName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
}

resource "aws_dataexchange_event_action" "test" {
  action {
    export_revision_to_s3 {
      encryption {
        type        = "aws:kms"
        kms_key_arn = aws_kms_key.test.arn
      }
      revision_destination {
        bucket = aws_s3_bucket.test.bucket
      }
    }
  }

  event {
    revision_published {
      data_set_id = %[1]q
    }
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, dataSetId),
	)
}
