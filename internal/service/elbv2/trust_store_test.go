// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2TrustStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TrustStore
	resourceName := "aws_lb_trust_store.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("truststore/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccELBV2TrustStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TrustStore
	resourceName := "aws_lb_trust_store.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelbv2.ResourceTrustStore(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBV2TrustStore_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TrustStore
	resourceName := "aws_lb_trust_store.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_nameGenerated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, names.AttrName, "tf-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccELBV2TrustStore_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TrustStore
	resourceName := "aws_lb_trust_store.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustStoreConfig_namePrefix(rName, "tf-px-"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustStoreExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-px-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-px-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func testAccCheckTrustStoreExists(ctx context.Context, n string, v *awstypes.TrustStore) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		output, err := tfelbv2.FindTrustStoreByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTrustStoreDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_trust_store" {
				continue
			}

			_, err := tfelbv2.FindTrustStoreByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELBv2 Trust Store %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTrustStoreConfig_baseS3BucketCA(rName string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "%[1]s.pem"
  content = <<EOT
-----BEGIN CERTIFICATE-----
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
LDgwwPky7T6W4ohoGv+p497rbPtHsLq9
-----END CERTIFICATE-----
EOT
}
`, rName))
}

func testAccTrustStoreConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_baseS3BucketCA(rName), fmt.Sprintf(`
resource "aws_lb_trust_store" "test" {
  name                             = %[1]q
  ca_certificates_bundle_s3_bucket = aws_s3_bucket.test.bucket
  ca_certificates_bundle_s3_key    = aws_s3_object.test.key
}
`, rName))
}

func testAccTrustStoreConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_baseS3BucketCA(rName), `
resource "aws_lb_trust_store" "test" {
  ca_certificates_bundle_s3_bucket = aws_s3_bucket.test.bucket
  ca_certificates_bundle_s3_key    = aws_s3_object.test.key
}
`)
}

func testAccTrustStoreConfig_namePrefix(rName, namePrefix string) string {
	return acctest.ConfigCompose(testAccTrustStoreConfig_baseS3BucketCA(rName), fmt.Sprintf(`
resource "aws_lb_trust_store" "test" {
  name_prefix                      = %[1]q
  ca_certificates_bundle_s3_bucket = aws_s3_bucket.test.bucket
  ca_certificates_bundle_s3_key    = aws_s3_object.test.key
}
`, namePrefix))
}
