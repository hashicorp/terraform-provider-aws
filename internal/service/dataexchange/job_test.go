// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataExchangeJob_assetsFromS3Basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dataexchange_job.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	var job dataexchange.GetJobOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_importAssetsFromS3(bucketName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					testAccCheckJobStarted(ctx, resourceName, false),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeJob_assetsFromS3PostponeStart(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dataexchange_job.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))
	var job dataexchange.GetJobOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_importAssetsFromS3(bucketName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					testAccCheckJobStarted(ctx, resourceName, false),
				),
			},
			{
				Config: testAccJobConfig_importAssetsFromS3(bucketName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobStarted(ctx, resourceName, true),
				),
			},
		},
	})
}

func TestAccDataExchangeJob_exportAssetsToS3Basic(t *testing.T) {
	ctx := acctest.Context(t)
	data, err := helperAccJobCreateDefaultAsset(ctx, awstypes.AssetTypeS3Snapshot)
	if err != nil {
		t.Error(err)
	}
	resourceName := "aws_dataexchange_job.test"
	bucketName := strconv.Itoa(int(time.Now().UnixNano()))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_exportAssetsToS3Basic(bucketName, *data),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &dataexchange.GetJobOutput{}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeJob_exportAssetsToSignedURL(t *testing.T) {
	ctx := acctest.Context(t)
	data, err := helperAccJobCreateDefaultAsset(ctx, awstypes.AssetTypeS3Snapshot)
	if err != nil {
		t.Error(err)
	}
	resourceName := "aws_dataexchange_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_exportAssetToSignedURL(*data),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &dataexchange.GetJobOutput{}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeJob_assetFromSignedURL(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dataexchange_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_importAssetFromSignedURL(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &dataexchange.GetJobOutput{}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
			},
		},
	})
}

func testAccCheckJobExists(ctx context.Context, n string, v *dataexchange.GetJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)
		input := dataexchange.GetJobInput{
			JobId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetJob(ctx, &input)
		if err != nil {
			return err
		}
		if output == nil {
			return fmt.Errorf("DataExchange Job not found")
		}

		*v = *output

		return nil
	}
}

func testAccCheckJobStarted(ctx context.Context, n string, started bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)
		input := dataexchange.GetJobInput{
			JobId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetJob(ctx, &input)
		if err != nil {
			return err
		}
		if output == nil {
			return fmt.Errorf("DataExchange Job not found")
		}

		if started && output.State == awstypes.StateWaiting {
			return fmt.Errorf("DataExchange Job not started")
		} else if !started && output.State != awstypes.StateWaiting {
			return fmt.Errorf("DataExchange Job started but not supposed to")
		}

		return nil
	}
}

func testAccJobConfig_importAssetsFromS3(bucketName string, start bool) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = "test"
  name        = "test"
}

resource "aws_dataexchange_revision" "test" {
  data_set_id = aws_dataexchange_data_set.test.id
}

resource "aws_s3_bucket" "test" {
  bucket        = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = "test"
  content = "test"
}

resource "aws_dataexchange_job" "test" {
  type              = "IMPORT_ASSETS_FROM_S3"
  start_on_creation = %t

  details {
    import_assets_from_s3 {
      data_set_id = aws_dataexchange_data_set.test.id
      revision_id = aws_dataexchange_revision.test.revision_id
      asset_sources {
        bucket = aws_s3_bucket.test.id
        key    = "test"
      }
    }
  }
}
`, bucketName, start)
}

func testAccJobConfig_exportAssetsToS3Basic(bucketName string, data testAsset) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%s"
  force_destroy = true
}

resource "aws_dataexchange_job" "test" {
  type = "EXPORT_ASSETS_TO_S3"

  details {
    export_assets_to_s3 {
      data_set_id = "%s"
      revision_id = "%s"
      asset_destinations {
        bucket   = aws_s3_bucket.test.id
        key      = "test"
        asset_id = "%s"
      }
    }
  }
}
`, bucketName, data.dataSetId, data.revisionId, data.assetId)
}

func testAccJobConfig_importAssetFromSignedURL() string {
	return `
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = "test"
  name        = "test"
}

resource "aws_dataexchange_revision" "test" {
  data_set_id = aws_dataexchange_data_set.test.id
}

resource "aws_dataexchange_job" "test" {
  type = "IMPORT_ASSET_FROM_SIGNED_URL"

  details {
    import_asset_from_signed_url {
      data_set_id = aws_dataexchange_data_set.test.id
      revision_id = aws_dataexchange_revision.test.revision_id
      asset_name  = "test"
      md5_hash    = "NTdmMmNkNmVkNzZlY2IyNTUK"
    }
  }
}
`
}

func testAccJobConfig_exportAssetToSignedURL(data testAsset) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_job" "test" {
  type = "EXPORT_ASSET_TO_SIGNED_URL"

  details {
    export_asset_to_signed_url {
      data_set_id = "%s"
      revision_id = "%s"
      asset_id    = "%s"
    }
  }
}
`, data.dataSetId, data.revisionId, data.assetId)
}

type testAsset struct {
	dataSetId  string
	revisionId string
	assetId    string
}

func helperAccJobCreateDefaultAsset(ctx context.Context, assetType awstypes.AssetType) (*testAsset, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	conn := dataexchange.NewFromConfig(cfg)

	// Create DataSet
	createDataSetInput := dataexchange.CreateDataSetInput{
		Name:        aws.String("test"),
		Description: aws.String("test"),
		AssetType:   assetType,
	}
	dsOut, err := conn.CreateDataSet(ctx, &createDataSetInput)
	if err != nil {
		return nil, err
	}

	// Create Revision
	createRevisionInput := dataexchange.CreateRevisionInput{
		DataSetId: dsOut.Id,
	}
	rOut, err := conn.CreateRevision(ctx, &createRevisionInput)
	if err != nil {
		return nil, err
	}

	createJobInput := dataexchange.CreateJobInput{
		Type: awstypes.TypeImportAssetFromSignedUrl,
		Details: &awstypes.RequestDetails{
			ImportAssetFromSignedUrl: &awstypes.ImportAssetFromSignedUrlRequestDetails{
				DataSetId:  dsOut.Id,
				RevisionId: rOut.Id,
				AssetName:  aws.String("file.txt"),
				Md5Hash:    aws.String("CY9rzUYh03PK3k6DJie09g=="),
			},
		},
	}
	jOut, err := conn.CreateJob(ctx, &createJobInput)
	if err != nil {
		return nil, err
	}

	if len(jOut.Errors) > 0 {
		return nil, errors.New("error creating a Job")
	}

	// Upload file
	baseUrl, err := url.Parse(*jOut.Details.ImportAssetFromSignedUrl.SignedUrl)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, baseUrl.String(), strings.NewReader("test"))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-MD5", "CY9rzUYh03PK3k6DJie09g==")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(respBody) > 0 {
		return nil, errors.New(string(respBody))
	}

	// Start job
	startJobInput := dataexchange.StartJobInput{
		JobId: jOut.Id,
	}
	_, err = conn.StartJob(ctx, &startJobInput)
	if err != nil {
		return nil, err
	}

	// Get the list of assets
	lOut := &dataexchange.ListRevisionAssetsOutput{
		Assets: []awstypes.AssetEntry{},
	}
	for len(lOut.Assets) == 0 {
		time.Sleep(time.Second)
		input := dataexchange.ListRevisionAssetsInput{
			RevisionId: rOut.Id,
			DataSetId:  dsOut.Id,
		}
		lOut, err = conn.ListRevisionAssets(ctx, &input)
		if err != nil {
			return nil, err
		}
	}

	return &testAsset{
		dataSetId:  *dsOut.Id,
		revisionId: *rOut.Id,
		assetId:    *lOut.Assets[0].Id,
	}, nil
}

func TestAccDataExchangeJob_helperCreateDefaultAsset(t *testing.T) {
	ctx := context.Background()
	data, err := helperAccJobCreateDefaultAsset(ctx, awstypes.AssetTypeS3Snapshot)
	if err != nil {
		t.Error(err)
	}

	if data.dataSetId == "" || data.revisionId == "" || data.assetId == "" {
		t.Errorf("helper func returned empty data")
	}
}
