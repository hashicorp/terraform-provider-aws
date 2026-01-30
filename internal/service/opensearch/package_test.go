// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchPackage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	pkgName := testAccRandomDomainName()
	resourceName := "aws_opensearch_package.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig_basic(pkgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackageExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "available_package_version", "v1"),
					resource.TestCheckResourceAttr(resourceName, "package_description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "package_id"),
					resource.TestCheckResourceAttr(resourceName, "package_name", pkgName),
					resource.TestCheckResourceAttr(resourceName, "package_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeTxtDictionary)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"available_package_version",
					"package_source", // This isn't returned by the API
				},
			},
		},
	})
}

func TestAccOpenSearchPackage_packageTypeZipPlugin(t *testing.T) {
	ctx := acctest.Context(t)
	pkgName := testAccRandomDomainName()
	resourceName := "aws_opensearch_package.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig_packageTypeZipPlugin(pkgName, "OpenSearch_2.17"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackageExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "available_package_version", "v1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "OpenSearch_2.17"),
					resource.TestCheckResourceAttr(resourceName, "package_description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "package_id"),
					resource.TestCheckResourceAttr(resourceName, "package_name", pkgName),
					resource.TestCheckResourceAttr(resourceName, "package_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeZipPlugin)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"available_package_version",
					"package_source", // This isn't returned by the API
				},
			},
			{
				// If engin_version is different from specified in the plugin zip file, it should return an error
				Config:      testAccPackageConfig_packageTypeZipPlugin(pkgName, "OpenSearch_2.11"),
				ExpectError: regexache.MustCompile(`doesn't matches with the provided EngineVersion`),
			},
		},
	})
}

func TestAccOpenSearchPackage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	pkgName := testAccRandomDomainName()
	resourceName := "aws_opensearch_package.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig_basic(pkgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackageExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfopensearch.ResourcePackage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPackageExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

		_, err := tfopensearch.FindPackageByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckPackageDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_package" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

			_, err := tfopensearch.FindPackageByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Package %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPackageConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = %[1]q
  source = "./test-fixtures/example-opensearch-custom-package.txt"
  etag   = filemd5("./test-fixtures/example-opensearch-custom-package.txt")
}

resource "aws_opensearch_package" "test" {
  package_name = %[1]q
  package_source {
    s3_bucket_name = aws_s3_bucket.test.bucket
    s3_key         = aws_s3_object.test.key
  }
  package_type = "TXT-DICTIONARY"
}
`, rName)
}

func testAccPackageConfig_packageTypeZipPlugin(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}


# example-opensearch-plugin.zip was created from the sample repository provided by AWS using the following commands:
# > git clone https://github.com/aws-samples/kr-tech-blog-sample-code.git
# > cd kr-tech-blog-sample-code/opensearch_custom_plugin
# > gradele build
# > cp build/distributions/opensearch-custom-plugin-1.0.0.zip terraform-provider-aws/internal/service/opensearch/test-fixtures/example-opensearch-plugin.zip

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = %[1]q
  source = "./test-fixtures/example-opensearch-plugin.zip"
  etag   = filemd5("./test-fixtures/example-opensearch-plugin.zip")
}

resource "aws_opensearch_package" "test" {
  engine_version = %[2]q
  package_name   = %[1]q
  package_source {
    s3_bucket_name = aws_s3_bucket.test.bucket
    s3_key         = aws_s3_object.test.key
  }
  package_type = "ZIP-PLUGIN"
}
`, rName, engineVersion)
}
