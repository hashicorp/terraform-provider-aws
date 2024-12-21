// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
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
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataExchangeJob_importAssetsFromS3basic(t *testing.T) {
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					testAccCheckJobStarted(ctx, resourceName, false),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeJob_importAssetsFromS3PostponeStart(t *testing.T) {
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					testAccCheckJobStarted(ctx, resourceName, false),
				),
			},
			{
				Config: testAccJobConfig_importAssetsFromS3(bucketName, true),
				Check: resource.ComposeTestCheckFunc(
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &dataexchange.GetJobOutput{}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeJob_exportAssetsToSignedUrl(t *testing.T) {
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
				Config: testAccJobConfig_exportAssetToSignedUrl(*data),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &dataexchange.GetJobOutput{}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataExchangeJob_importAssetFromSignedUrl(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dataexchange_job.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_importAssetFromSignedUrl(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &dataexchange.GetJobOutput{}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
				ExpectNonEmptyPlan: true,
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
		output, err := conn.GetJob(ctx, &dataexchange.GetJobInput{
			JobId: aws.String(rs.Primary.ID),
		})
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
		output, err := conn.GetJob(ctx, &dataexchange.GetJobInput{
			JobId: aws.String(rs.Primary.ID),
		})
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
  bucket = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key = "test"
  content = "test"
}

resource "aws_dataexchange_job" "test" {
  type  = "IMPORT_ASSETS_FROM_S3"
  start_on_creation = %t
  data_set_id = aws_dataexchange_data_set.test.id
  revision_id = aws_dataexchange_revision.test.revision_id
  s3_asset_sources {
      bucket = aws_s3_bucket.test.id
      key = "test"
  }
}
`, bucketName, start)
}

func testAccJobConfig_exportAssetsToS3Basic(bucketName string, data testAsset) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%s"
  force_destroy = true
}

resource "aws_dataexchange_job" "test" {
  type  = "EXPORT_ASSETS_TO_S3"

  data_set_id = "%s"
  revision_id = "%s"
  s3_asset_destinations {
      bucket = aws_s3_bucket.test.id
      key = "test"
      asset_id = "%s"
  }
}
`, bucketName, data.dataSetId, data.revisionId, data.assetId)
}

func testAccJobConfig_importAssetFromSignedUrl() string {
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
  type  = "IMPORT_ASSET_FROM_SIGNED_URL"

  data_set_id = aws_dataexchange_data_set.test.id
  revision_id = aws_dataexchange_revision.test.revision_id
  url_asset_name = "test"
  url_asset_md5_hash = "NTdmMmNkNmVkNzZlY2IyNTUK"
}
`
}

func testAccJobConfig_exportAssetToSignedUrl(data testAsset) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_job" "test" {
  type  = "EXPORT_ASSET_TO_SIGNED_URL"

  data_set_id = "%s"
  revision_id = "%s"
  asset_id = "%s"
}
`, data.dataSetId, data.revisionId, data.assetId)
}

type testAsset struct {
	dataSetId  string
	revisionId string
	assetId    string
}

func helperAccJobCreateDefaultAsset(ctx context.Context, assetType awstypes.AssetType) (*testAsset, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	conn := dataexchange.NewFromConfig(cfg)
	// Create DataSet
	dsOut, err := conn.CreateDataSet(ctx, &dataexchange.CreateDataSetInput{
		Name:        aws.String("test"),
		Description: aws.String("test"),
		AssetType:   assetType,
	})

	if err != nil {
		return nil, err
	}

	// Create Revision
	rOut, err := conn.CreateRevision(ctx, &dataexchange.CreateRevisionInput{
		DataSetId: dsOut.Id,
	})

	if err != nil {
		return nil, err
	}

	jOut, err := conn.CreateJob(ctx, &dataexchange.CreateJobInput{
		Type: awstypes.TypeImportAssetFromSignedUrl,
		Details: &awstypes.RequestDetails{
			ImportAssetFromSignedUrl: &awstypes.ImportAssetFromSignedUrlRequestDetails{
				DataSetId:  dsOut.Id,
				RevisionId: rOut.Id,
				AssetName:  aws.String("file.txt"),
				Md5Hash:    aws.String("CY9rzUYh03PK3k6DJie09g=="),
			},
		},
	})

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

	req, err := http.NewRequest("PUT", baseUrl.String(), strings.NewReader("test"))
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
	_, err = conn.StartJob(ctx, &dataexchange.StartJobInput{
		JobId: jOut.Id,
	})
	if err != nil {
		return nil, err
	}

	// Get the list of assets
	lOut := &dataexchange.ListRevisionAssetsOutput{
		Assets: []awstypes.AssetEntry{},
	}
	for len(lOut.Assets) == 0 {
		time.Sleep(time.Second)
		lOut, err = conn.ListRevisionAssets(ctx, &dataexchange.ListRevisionAssetsInput{
			RevisionId: rOut.Id,
			DataSetId:  dsOut.Id,
		})
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

func TestAccHelperJobAccCreateDefaultAsset(t *testing.T) {
	ctx := context.Background()
	data, err := helperAccJobCreateDefaultAsset(ctx, awstypes.AssetTypeS3Snapshot)
	if err != nil {
		t.Error(err)
	}

	if data.dataSetId == "" || data.revisionId == "" || data.assetId == "" {
		t.Errorf("helper func returned empty data")
	}
}
