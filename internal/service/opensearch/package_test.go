// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchPackage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	pkgName := testAccRandomDomainName()
	resourceName := "aws_opensearch_package.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig_basic(pkgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackageExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "available_package_version", ""),
					resource.TestCheckResourceAttr(resourceName, "package_description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "package_id"),
					resource.TestCheckResourceAttr(resourceName, "package_name", pkgName),
					resource.TestCheckResourceAttr(resourceName, "package_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "package_type", "TXT-DICTIONARY"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"package_source", // This isn't returned by the API
				},
			},
		},
	})
}

func TestAccOpenSearchPackage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	pkgName := testAccRandomDomainName()
	resourceName := "aws_opensearch_package.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig_basic(pkgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackageExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopensearch.ResourcePackage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPackageExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)

		_, err := tfopensearch.FindPackageByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckPackageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_package" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)

			_, err := tfopensearch.FindPackageByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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
