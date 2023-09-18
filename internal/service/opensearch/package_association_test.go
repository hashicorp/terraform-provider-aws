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

func TestAccOpenSearchPackageAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_package_association.test"
	packageResourceName := "aws_opensearch_package.test"
	domainResourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageAssociationConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "package_name", name),
					resource.TestCheckResourceAttrPair(resourceName, "package_id", packageResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainResourceName, "domain_name"),
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

func TestAccOpenSearchPackageAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_package_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageAssociationConfig(name),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopensearch.ResourcePackage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPackageAssociationConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "opensearch_packages" {
  bucket = "%s"
}

resource "aws_s3_object" "example_txt" {
  bucket = aws_s3_bucket.opensearch_packages.bucket
  key    = "%s"
  source = "./test-fixtures/example-opensearch-custom-package.txt"
  etag   = filemd5("./test-fixtures/example-opensearch-custom-package.txt")
}

resource "aws_opensearch_package" "test" {
  package_name = "%s"
  package_source {
    s3_bucket_name = aws_s3_bucket.opensearch_packages.bucket
    s3_key         = aws_s3_object.example_txt.key
  }
  package_type = "TXT-DICTIONARY"
}

resource "aws_opensearch_domain" "test" {
  domain_name = "%s"

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
`, name, name, name, name)
}
