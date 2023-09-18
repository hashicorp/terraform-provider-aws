// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opensearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccOpenSearchPackage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_package.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "package_name", name),
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
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_package.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig(name),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopensearch.ResourcePackage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPackageConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "opensearch_packages" {
  bucket = "%s"
}

resource "aws_s3_object" "example_txt" {
  bucket = aws_s3_bucket.opensearch_packages.bucket
  key    = "%s"
  source = "./test-fixtures/example-opensearch-custom-package.txt"
  etag = filemd5("./test-fixtures/example-opensearch-custom-package.txt")
}

resource "aws_opensearch_package" "test" {
  package_name = "%s"
  package_source {
    s3_bucket_name = aws_s3_bucket.opensearch_packages.bucket
    s3_key = aws_s3_object.example_txt.key
  }
  package_type = "TXT-DICTIONARY"
}
`, name, name, name)
}

func testAccCheckPackageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_domain" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)
			_, err := tfopensearch.FindPackageByName(ctx, conn, rs.Primary.Attributes["package_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch package %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
