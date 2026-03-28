// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchPackageAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := testAccRandomDomainName(t)
	pkgName := testAccRandomDomainName(t)
	resourceName := "aws_opensearch_package_association.test"
	packageResourceName := "aws_opensearch_package.test"
	domainResourceName := "aws_opensearch_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageAssociationConfig_basic(pkgName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackageAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, domainResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, "package_id", packageResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccPackageAssociationImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccOpenSearchPackageAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := testAccRandomDomainName(t)
	pkgName := testAccRandomDomainName(t)
	resourceName := "aws_opensearch_package_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageAssociationConfig_basic(pkgName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackageAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfopensearch.ResourcePackageAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPackageAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

		_, err := tfopensearch.FindPackageAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["package_id"])

		return err
	}
}

func testAccCheckPackageAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_package_association" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

			_, err := tfopensearch.FindPackageAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["package_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Package Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPackageAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		domainName := rs.Primary.Attributes[names.AttrDomainName]
		packageID := rs.Primary.Attributes["package_id"]

		return fmt.Sprintf("%s:%s", domainName, packageID), nil
	}
}

func testAccPackageAssociationConfig_basic(pkgName, domainName string) string {
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

resource "aws_opensearch_domain" "test" {
  domain_name = %[2]q

  cluster_config {
    instance_type = "t3.small.search" # supported in both aws and aws-us-gov
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_opensearch_package_association" "test" {
  package_id  = aws_opensearch_package.test.id
  domain_name = aws_opensearch_domain.test.domain_name
}
`, pkgName, domainName)
}
