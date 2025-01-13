package dataexchange_test

import (
	"context"
	"fmt"
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestAccDataExchangeEventAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
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

	dataSetId, err := helperAccGetReceivedDataSet(ctx)
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
		CheckDestroy:             testAccCheckEventActionDestroy(ctx, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_basic(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &dataexchange.GetEventActionOutput{}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`event-actions/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeEventAction_full(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dataexchange_event_action.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	if _, okAcc := os.LookupEnv("TF_ACC"); !okAcc {
		t.Skipf("TF_ACC must be set")
	}

	dataSetId, err := helperAccGetReceivedDataSet(ctx)

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
		CheckDestroy:             testAccCheckEventActionDestroy(ctx, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccEventActionConfig_full(bucketName, dataSetId, *identity.Account),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventActionExists(ctx, resourceName, &dataexchange.GetEventActionOutput{}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`event-actions/.+`)),
				),
			},
		},
	})
}

func testAccCheckEventActionDestroy(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)
		_, err := conn.GetEventAction(ctx, &dataexchange.GetEventActionInput{
			EventActionId: aws.String(rs.Primary.ID),
		})
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DataExchange EventAction %s still exists", rs.Primary.ID)
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

func testAccEventActionConfig_full(bucketName, dataSetId, accountId string) string {
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

func helperAccGetReceivedDataSet(ctx context.Context) (string, error) {
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
		if dataSet.AssetType == awstypes.AssetTypeS3Snapshot {
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

func TestHelperAccGetReceivedDataSet(t *testing.T) {
	ctx := context.Background()
	if _, okAcc := os.LookupEnv("TF_ACC"); !okAcc {
		t.Skipf("TF_ACC must be set")
	}

	_, err := helperAccGetReceivedDataSet(ctx)
	if err != nil {
		t.Error(err)
	}
}
