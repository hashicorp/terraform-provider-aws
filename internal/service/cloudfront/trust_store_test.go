// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontTrustStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var truststore cloudfront.GetTrustStoreOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_trust_store.test"
	objectKey := "ca-bundle.pem"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(rName, objectKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &truststore),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNRegexp("cloudfront", regexache.MustCompile(`trust-store/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_of_ca_certificates"), knownvalue.Int32Exact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_trust_store.test"
	objectKey := "ca-bundle.pem"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(rName, objectKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &truststore),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudfront.ResourceTrustStore, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccCloudFrontTrustStore_withS3ObjectVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var truststore cloudfront.GetTrustStoreOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_trust_store.test"
	objectKey := "ca-bundle.pem"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_withS3ObjectVersion(rName, objectKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &truststore),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
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
	var truststore cloudfront.GetTrustStoreOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_trust_store.test"
	objectKey1 := "ca-bundle_v1.pem"
	objectKey2 := "ca-bundle_v2.pem"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(rName, objectKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &truststore),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccTrustStoreConfig_update(rName, objectKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &truststore),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
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

func TestAccCloudFrontTrustStore_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var truststore cloudfront.GetTrustStoreOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_trust_store.test"
	objectKey := "ca-bundle.pem"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_tags1(rName, objectKey, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &truststore),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ca_certificates_bundle_source",
				},
			},
			{
				Config: testAccTrustStoreConfig_tags2(rName, objectKey, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &truststore),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccTrustStoreConfig_tags1(rName, objectKey, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, t, resourceName, &truststore),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckTrustStoreDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_trust_store" {
				continue
			}

			_, err := tfcloudfront.FindTrustStoreByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Trust Store %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTrustStoreExists(ctx context.Context, t *testing.T, n string, v *cloudfront.GetTrustStoreOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		resp, err := tfcloudfront.FindTrustStoreByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

const testAccTrustStoreCertificateContent = `-----BEGIN CERTIFICATE-----
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

func testAccTrustStoreConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccTrustStoreConfig_basic(rName, key string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_base(rName), fmt.Sprintf(`
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
`, rName, key, testAccTrustStoreCertificateContent))
}

func testAccTrustStoreConfig_withS3ObjectVersion(rName, key string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_base(rName), fmt.Sprintf(`
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
`, rName, key, testAccTrustStoreCertificateContent))
}

func testAccTrustStoreConfig_update(rName, key string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_base(rName), fmt.Sprintf(`
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
`, rName, key, testAccTrustStoreCertificateContent))
}

func testAccTrustStoreConfig_tags1(rName, key, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_base(rName), fmt.Sprintf(`
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

  tags = {
    %[4]q = %[5]q
  }
}
`, rName, key, testAccTrustStoreCertificateContent, tagKey1, tagValue1))
}

func testAccTrustStoreConfig_tags2(rName, key, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_base(rName), fmt.Sprintf(`
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

  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }
}
`, rName, key, testAccTrustStoreCertificateContent, tagKey1, tagValue1, tagKey2, tagValue2))
}
