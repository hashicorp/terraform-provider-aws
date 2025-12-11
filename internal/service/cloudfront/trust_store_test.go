// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontTrustStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var truststore cloudfront.GetTrustStoreOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_trust_store.test"
	objectKey := "ca-bundle.pem"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(rName, objectKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &truststore),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "trust-store/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.TrustStoreStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.key", objectKey),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.region", acctest.Region()),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, "number_of_ca_certificates", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ca_certificates_bundle_source",
				},
			},
		},
	})
}

func TestAccCloudFrontTrustStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var truststore cloudfront.GetTrustStoreOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_trust_store.test"
	objectKey := "ca-bundle.pem"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(rName, objectKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &truststore),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceTrustStore, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontTrustStore_withVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var truststore cloudfront.GetTrustStoreOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_trust_store.test"
	objectKey := "ca-bundle.pem"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_withVersion(rName, objectKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &truststore),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "trust-store/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.TrustStoreStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.key", objectKey),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.region", acctest.Region()),
					resource.TestCheckResourceAttrSet(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.version"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, "number_of_ca_certificates", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ca_certificates_bundle_source",
				},
			},
		},
	})
}

func TestAccCloudFrontTrustStore_update(t *testing.T) {
	ctx := acctest.Context(t)
	var truststore1, truststore2 cloudfront.GetTrustStoreOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_trust_store.test"
	objectKey1 := "ca-bundle_v1.pem"
	objectKey2 := "ca-bundle_v2.pem"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(rName, objectKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &truststore1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "trust-store/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.TrustStoreStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.key", objectKey1),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.region", acctest.Region()),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, "number_of_ca_certificates", "1"),
				),
			},
			{
				Config: testAccTrustStoreConfig_update(rName, objectKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &truststore2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "trust-store/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.TrustStoreStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.key", objectKey2),
					resource.TestCheckResourceAttr(resourceName, "ca_certificates_bundle_source.0.ca_certificates_bundle_s3_location.0.region", acctest.Region()),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, "number_of_ca_certificates", "1"),
					testAccCheckTrustStoreRecreated(&truststore1, &truststore2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ca_certificates_bundle_source",
				},
			},
		},
	})
}

func testAccCheckTrustStoreDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_trust_store" {
				continue
			}

			_, err := tfcloudfront.FindTrustStoreByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.CloudFront, create.ErrActionCheckingDestroyed, tfcloudfront.ResNameTrustStore, rs.Primary.ID, err)
			}

			return create.Error(names.CloudFront, create.ErrActionCheckingDestroyed, tfcloudfront.ResNameTrustStore, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTrustStoreExists(ctx context.Context, name string, truststore *cloudfront.GetTrustStoreOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameTrustStore, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameTrustStore, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		resp, err := tfcloudfront.FindTrustStoreByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameTrustStore, rs.Primary.ID, err)
		}

		*truststore = *resp

		return nil
	}
}

func testAccCheckTrustStoreRecreated(before, after *cloudfront.GetTrustStoreOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.TrustStore.Id == after.TrustStore.Id {
			return errors.New("trust store was not recreated")
		}
		return nil
	}
}

// testCertificateContent returns the common certificate content used across tests
func testCertificateContent() string {
	return `-----BEGIN CERTIFICATE-----
MIIDQTCCAimgAwIBAgITBmyfz5m/jAo54vB4ikPmljZbyjANBgkqhkiG9w0BAQsF
ADA5MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRkwFwYDVQQDExBBbWF6
b24gUm9vdCBDQSAxMB4XDTE1MDUyNjAwMDAwMFoXDTM4MDExNzAwMDAwMFowOTEL
MAkGA1UEBhMCVVMxDzANBgNVBAoTBkFtYXpvbjEZMBcGA1UEAxMQQW1hem9uIFJv
b3QgQ0EgMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALJ4gHHKeNXj
ca9HgFB0fW7Y14h29Jlo91ghYPl0hAEvrAIthtOgQ3pOsqTQNroBvo3bSMgHFzZM
9O6II8c+6zf1tRn4SWiw3te5djgdYZ6k/oI2peVKVuRF4fn9tBb6dNqcmzU5L/qw
IFAGbHrQgLKm+a/sRxmPUDgH3KKHOVj4utWp+UhnMJbulHheb4mjUcAwhmahRWa6
VOujw5H5SNz/0egwLX0tdHA114gk957EWW67c4cX8jJGKLhD+rcdqsq08p8kDi1L
93FcXmn/6pUCyziKrlA4b9v7LWIbxcceVOF34GfID5yHI9Y/QCB/IIDEgEw+OyQm
jgSubJrIqg0CAwEAAaNCMEAwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMC
AYYwHQYDVR0OBBYEFIQYzIU07LwMlJQuCFmcx7IQTgoIMA0GCSqGSIb3DQEBCwUA
A4IBAQCY8jdaQZChGsV2USggNiMOruYou6r4lK5IpDB/G/wkjUu0yKGX9rbxenDI
U5PMCCjjmCXPI6T53iHTfIUJrU6adTrCC2qJeHZERxhlbI1Bjjt/msv0tadQ1wUs
N+gDS63pYaACbvXy8MWy7Vu33PqUXHeeE6V/Uq2V8viTO96LXFvKWlJbYK8U90vv
o/ufQJVtMVT8QtPHRh8jrdkPSHCa2XV4cdFyQzR1bldZwgJcJmApzyMZFo6IQ6XU
5MsI+yMRQ+hDKXJioaldXgjUkK642M4UwtBV8ob2xJNDd2ZhwLnoQdeXeGADbkpy
rqXRfboQnoZsG4q5WTP468SQvvG5
-----END CERTIFICATE-----`
}

// testAccTrustStoreConfigBase returns the common base configuration for S3 bucket and data sources
func testAccTrustStoreConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccTrustStoreConfig_basic(rName, key string) string {
	return testAccTrustStoreConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = %[2]q
  content = <<-EOT
%[3]s
EOT
}

resource "aws_cloudfront_trust_store" "test" {
  name = %[1]q

  ca_certificates_bundle_source {
    ca_certificates_bundle_s3_location {
      bucket = aws_s3_bucket.test.id
      key    = aws_s3_object.test.key
      region = data.aws_region.current.name
    }
  }
}
`, rName, key, testCertificateContent())
}

func testAccTrustStoreConfig_withVersion(rName, key string) string {
	return testAccTrustStoreConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "test" {
  bucket     = aws_s3_bucket.test.id
  key        = %[2]q
  content    = <<-EOT
%[3]s
EOT
  depends_on = [aws_s3_bucket_versioning.test]
}

resource "aws_cloudfront_trust_store" "test" {
  name = %[1]q

  ca_certificates_bundle_source {
    ca_certificates_bundle_s3_location {
      bucket  = aws_s3_bucket.test.id
      key     = aws_s3_object.test.key
      region  = data.aws_region.current.name
      version = aws_s3_object.test.version_id
    }
  }
}
`, rName, key, testCertificateContent())
}

func testAccTrustStoreConfig_update(rName, key string) string {
	return testAccTrustStoreConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "test" {
  bucket     = aws_s3_bucket.test.id
  key        = %[2]q
  content    = <<-EOT
%[3]s
EOT
  depends_on = [aws_s3_bucket_versioning.test]
}

resource "aws_cloudfront_trust_store" "test" {
  name = %[1]q

  ca_certificates_bundle_source {
    ca_certificates_bundle_s3_location {
      bucket  = aws_s3_bucket.test.id
      key     = aws_s3_object.test.key
      region  = data.aws_region.current.name
      version = aws_s3_object.test.version_id
    }
  }
}
`, rName, key, testCertificateContent())
}
