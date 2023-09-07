// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opensearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
)

func TestAccOpenSearchPackage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_package.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_1", &domain),
					resource.TestCheckResourceAttr(resourceName, "package_name", "example"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchPackage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_package.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_1", &domain),
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
  bucket = "%s-1"
}

resource "aws_s3_object" "example_txt" {
  bucket = aws_s3_bucket.opensearch_packages.bucket
  key    = "example-opensearch-custom-package.txt"
  source = "./test-fixtures/example-opensearch-custom-package.txt"
  etag = filemd5("./test-fixtures/example-opensearch-custom-package.txt")
}

resource "aws_opensearch_package" "example_txt" {
  package_name = "example"
  package_source {
    s3_bucket_name = aws_s3_bucket.opensearch_packages.bucket
    s3_key = aws_s3_object.example_txt.key
  }
  package_type = "TXT-DICTIONARY"
}
`, name)
}
